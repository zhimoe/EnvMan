package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"envman/internal/domain"
	"envman/internal/environment"
	"envman/internal/storage"
)

func TestBatchCaseOnlyChangeCanBeApplied(t *testing.T) {
	a, tt := newTestApp(t)
	memoryBackend(t, a).Replace(ScopeUser, []Variable{{Name: "CUSTOM", Value: "foo;bar", Type: "REG_SZ"}})
	a.reload()
	a.openBatch("CUSTOM")
	a.batch.text = "FOO\nbar"
	tt.Frame()
	if err := tt.Click("应用更改"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	value, _ := a.findVariable(ScopeUser, "CUSTOM")
	if value.Value != "FOO;bar" || a.batch != nil {
		t.Fatalf("case-only edit not saved: %+v", value)
	}
	if analyzeBatch("foo;bar", "foo\nbar").hasChanges {
		t.Fatal("line separators counted as edits")
	}
}

type partialBackend struct {
	*environment.Memory
	fail bool
}

func (b *partialBackend) ApplyChanges(scope domain.Scope, before, desired []domain.Variable) error {
	if !b.fail {
		return b.Memory.ApplyChanges(scope, before, desired)
	}
	upserts, _ := domain.Diff(before, desired)
	if len(upserts) == 0 {
		return errors.New("expected upserts")
	}
	partial := domain.Clone(before)
	_, index := domain.Find(partial, upserts[0].Name)
	if index < 0 {
		partial = append(partial, upserts[0])
	} else {
		partial[index] = upserts[0]
	}
	if err := b.Memory.ApplyChanges(scope, before, partial); err != nil {
		return err
	}
	return &environment.ApplyError{Upserts: upserts[:1], Err: errors.New("injected failure")}
}

func TestPartialCommitRefreshesAndOffersRecovery(t *testing.T) {
	backend := &partialBackend{Memory: environment.NewMemory(map[Scope][]Variable{ScopeUser: {{Name: "A", Value: "old"}, {Name: "B", Value: "old"}}}), fail: true}
	a := newAppWith(backend, storage.New(t.TempDir()))
	before := cloneVars(a.variables(ScopeUser))
	desired := cloneVars(before)
	desired[0].Value, desired[1].Value = "new", "new"
	if a.commit(ScopeUser, desired, "test", "success", "", before) {
		t.Fatal("partial failure reported as success")
	}
	first, _ := a.findVariable(ScopeUser, "A")
	second, _ := a.findVariable(ScopeUser, "B")
	if first.Value != "new" || second.Value != "old" {
		t.Fatalf("UI not refreshed: %+v", a.variables(ScopeUser))
	}
	if len(a.snapshots) != 1 || a.toastAction == nil || !strings.Contains(a.toastAction.message, "部分成功") {
		t.Fatal("missing backup or recovery message")
	}
	backend.fail = false
	a.toastAction.action()
	a.dismissModal(true)
	first, _ = a.findVariable(ScopeUser, "A")
	if first.Value != "old" {
		t.Fatal("backup recovery failed")
	}
}

func TestCorruptBackupBlocksCommit(t *testing.T) {
	a, tt := newTestApp(t)
	before := cloneVars(a.variables(ScopeUser))
	path := filepath.Join(a.store.Dir(), "snapshots.json")
	if err := os.WriteFile(path, []byte("broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	desired := cloneVars(before)
	desired[0].Value = "changed"
	if a.commit(ScopeUser, desired, "test", "success", "", before) {
		t.Fatal("commit ignored corrupt backup")
	}
	current, err := a.env.Load(ScopeUser)
	if err != nil {
		t.Fatal(err)
	}
	if current[0].Value != before[0].Value {
		t.Fatal("environment modified without backup")
	}
	a.reload()
	a.backupsOpen = true
	tt.Frame()
	if a.snapshotsError == "" || tt.HasText("还没有备份") {
		t.Fatal("backup error shown as empty history")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "broken" {
		t.Fatal("corrupt history overwritten")
	}
}

func TestElevatedPartialApplyOffersRecovery(t *testing.T) {
	backend := &partialBackend{Memory: environment.NewMemory(map[Scope][]Variable{ScopeSystem: {{Name: "A", Value: "old"}, {Name: "B", Value: "old"}}}), fail: true}
	store := storage.New(t.TempDir())
	before, err := backend.Load(ScopeSystem)
	if err != nil {
		t.Fatal(err)
	}
	desired := cloneVars(before)
	desired[0].Value, desired[1].Value = "new", "new"
	if err := store.WritePending(PendingEdits{Scope: ScopeSystem, Before: before, Variables: desired}); err != nil {
		t.Fatal(err)
	}
	a := newAppWith(backend, store)
	if a.toastAction == nil || !strings.Contains(a.toastAction.message, "部分成功") || len(a.snapshots) != 1 {
		t.Fatal("startup failure missing recovery")
	}
	first, _ := a.findVariable(ScopeSystem, "A")
	if first.Value != "new" {
		t.Fatal("startup failure did not refresh state")
	}
}
