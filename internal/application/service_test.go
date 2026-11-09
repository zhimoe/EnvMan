package application_test

import (
	"envman/internal/application"
	"envman/internal/domain"
	"envman/internal/environment"
	"envman/internal/storage"
	"errors"
	"testing"
)

type failingBackups struct{}

func (failingBackups) SaveSnapshot(domain.Scope, string, []domain.Variable) (application.Snapshot, error) {
	return application.Snapshot{}, errors.New("disk unavailable")
}

func TestApplyRequiresBackup(t *testing.T) {
	before := []domain.Variable{{Name: "A", Value: "old"}}
	backend := environment.NewMemory(map[domain.Scope][]domain.Variable{domain.ScopeUser: before})
	service := application.New(backend, failingBackups{})
	_, err := service.Apply(application.Change{Scope: domain.ScopeUser, Before: before, Desired: []domain.Variable{{Name: "A", Value: "new"}}})
	if err == nil {
		t.Fatal("expected backup failure")
	}
	current, _ := backend.Load(domain.ScopeUser)
	if current[0].Value != "old" {
		t.Fatal("write occurred without backup")
	}
}

func TestApplyPreservesConcurrentChangesAndSupportsUndo(t *testing.T) {
	before := []domain.Variable{{Name: "A", Value: "old"}}
	backend := environment.NewMemory(map[domain.Scope][]domain.Variable{domain.ScopeUser: {{Name: "A", Value: "old"}, {Name: "B", Value: "external"}}})
	service := application.New(backend, storage.New(t.TempDir()))
	result, err := service.Apply(application.Change{Scope: domain.ScopeUser, Before: before, Desired: []domain.Variable{{Name: "A", Value: "new"}}})
	if err != nil {
		t.Fatal(err)
	}
	current, _ := backend.Load(domain.ScopeUser)
	if b, _ := domain.Find(current, "B"); b.Value != "external" {
		t.Fatal("unrelated change lost")
	}
	_, err = service.Apply(application.Change{Scope: domain.ScopeUser, Before: result.Desired, Desired: result.Snapshot.Variables})
	if err != nil {
		t.Fatal(err)
	}
	current, _ = backend.Load(domain.ScopeUser)
	if a, _ := domain.Find(current, "A"); a.Value != "old" {
		t.Fatal("undo failed")
	}
}

func TestApplyRejectsInvalidCommands(t *testing.T) {
	for _, change := range []application.Change{
		{Scope: domain.Scope(99)},
		{Desired: []domain.Variable{{Name: "A"}, {Name: "a"}}},
		{Desired: []domain.Variable{{Name: "A=B"}}},
		{Desired: []domain.Variable{{Name: "A", Value: "bad\x00value"}}},
		{Desired: []domain.Variable{{Name: "A", Type: "REG_BINARY"}}},
	} {
		backend := environment.NewMemory(nil)
		store := storage.New(t.TempDir())
		if _, err := application.New(backend, store).Apply(change); err == nil {
			t.Fatal("invalid command accepted")
		}
		snapshots, _ := store.LoadSnapshots()
		if len(snapshots) != 0 {
			t.Fatal("invalid command created backup")
		}
	}
}
