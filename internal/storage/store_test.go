package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"envman/internal/domain"
)

func TestSnapshotRetentionAndPendingRoundTrip(t *testing.T) {
	store := New(t.TempDir())
	variables := []domain.Variable{{Name: "Path", Value: `C:\Windows`, Type: "REG_SZ"}}
	for i := range MaxSnapshots + 5 {
		if _, err := store.SaveSnapshot(domain.ScopeUser, fmt.Sprintf("snapshot-%02d", i), variables); err != nil {
			t.Fatal(err)
		}
	}
	snapshots, err := store.LoadSnapshots()
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshots) != MaxSnapshots {
		t.Fatalf("snapshot count = %d, want %d", len(snapshots), MaxSnapshots)
	}
	if snapshots[0].Note != "snapshot-24" || snapshots[len(snapshots)-1].Note != "snapshot-05" {
		t.Fatalf("retained snapshots = %q ... %q", snapshots[0].Note, snapshots[len(snapshots)-1].Note)
	}

	pending := PendingEdits{Scope: domain.ScopeSystem, Note: "pending", Variables: variables}
	if err := store.WritePending(pending); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.ReadPending()
	if err != nil {
		t.Fatal(err)
	}
	if loaded == nil || loaded.Note != pending.Note || len(loaded.Variables) != 1 {
		t.Fatalf("pending round trip = %+v", loaded)
	}
	store.ClearPending()
	if loaded, err := store.ReadPending(); err != nil || loaded != nil {
		t.Fatalf("pending file survived clear: %+v", loaded)
	}
}

func TestSnapshotReadFailurePreservesHistory(t *testing.T) {
	for _, tc := range []struct{ name, data string }{
		{"invalid JSON", `[{"id":"keep"}`},
		{"wrong shape", `{"id":"keep"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := New(t.TempDir())
			if err := os.WriteFile(store.snapshotsPath(), []byte(tc.data), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := store.LoadSnapshots(); err == nil {
				t.Fatal("corruption was ignored")
			}
			if _, err := store.SaveSnapshot(domain.ScopeUser, "new", nil); err == nil {
				t.Fatal("save overwrote unreadable history")
			}
			data, err := os.ReadFile(store.snapshotsPath())
			if err != nil || string(data) != tc.data {
				t.Fatalf("history changed: %q, %v", data, err)
			}
		})
	}
}

func TestSnapshotMissingAndReadError(t *testing.T) {
	store := New(t.TempDir())
	if snapshots, err := store.LoadSnapshots(); err != nil || len(snapshots) != 0 {
		t.Fatalf("missing history: %v, %v", snapshots, err)
	}
	// A directory reliably produces a read error, including when tests run as root.
	if err := os.Mkdir(store.snapshotsPath(), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := store.LoadSnapshots(); err == nil {
		t.Fatal("read error was ignored")
	}
	if _, err := store.SaveSnapshot(domain.ScopeUser, "new", nil); err == nil {
		t.Fatal("save ignored read error")
	}
	info, err := os.Stat(store.snapshotsPath())
	if err != nil || !info.IsDir() {
		t.Fatalf("original path replaced: %v", err)
	}
}

func TestPendingErrorsAndClaim(t *testing.T) {
	store := New(t.TempDir())
	if err := os.WriteFile(store.pendingPath(), []byte("broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ReadPending(); err == nil {
		t.Fatal("corruption hidden")
	}
	if err := store.WritePending(PendingEdits{Scope: domain.ScopeSystem}); err != nil {
		t.Fatal(err)
	}
	name, err := store.ClaimPending()
	if err != nil {
		t.Fatal(err)
	}
	if pending, err := store.ReadPending(); err != nil || pending != nil {
		t.Fatal("claimed command can replay")
	}
	if _, err := os.Stat(filepath.Join(store.Dir(), name)); err != nil {
		t.Fatal("recovery command lost", err)
	}
	if err := store.CompletePending(name); err != nil {
		t.Fatal(err)
	}
}
