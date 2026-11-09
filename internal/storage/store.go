package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"envman/internal/application"
	"envman/internal/domain"
)

const MaxSnapshots = 20

type Snapshot = application.Snapshot
type PendingEdits = application.PendingEdits

type Store struct {
	dir string
}

func New(dir string) *Store {
	if dir == "" {
		base, err := os.UserConfigDir()
		if err != nil || base == "" {
			base = os.TempDir()
		}
		dir = filepath.Join(base, "EnvMan")
	}
	_ = os.MkdirAll(dir, 0o700)
	return &Store{dir: dir}
}

func (s *Store) Dir() string { return s.dir }

func (s *Store) snapshotsPath() string { return filepath.Join(s.dir, "snapshots.json") }
func (s *Store) pendingPath() string   { return filepath.Join(s.dir, "pending.json") }
func (s *Store) settingsPath() string  { return filepath.Join(s.dir, "settings.json") }

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, data, 0o600); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}

func (s *Store) LoadSnapshots() ([]Snapshot, error) {
	var snapshots []Snapshot
	if err := readJSON(s.snapshotsPath(), &snapshots); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("读取备份历史失败：%w", err)
	}
	return snapshots, nil
}

func (s *Store) SaveSnapshot(scope domain.Scope, note string, variables []domain.Variable) (Snapshot, error) {
	now := time.Now()
	snapshot := Snapshot{
		ID:        now.Format("20060102-150405.000000000"),
		Time:      now,
		Scope:     scope,
		Note:      note,
		Variables: domain.Clone(variables),
	}
	snapshots, err := s.LoadSnapshots()
	if err != nil {
		return snapshot, err
	}
	snapshots = append([]Snapshot{snapshot}, snapshots...)
	if len(snapshots) > MaxSnapshots {
		snapshots = snapshots[:MaxSnapshots]
	}
	if err := writeJSON(s.snapshotsPath(), snapshots); err != nil {
		return snapshot, err
	}
	return snapshot, nil
}

func (s *Store) ReadPending() (*PendingEdits, error) {
	var pending PendingEdits
	if err := readJSON(s.pendingPath(), &pending); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return &pending, nil
}

func (s *Store) WritePending(pending PendingEdits) error {
	return writeJSON(s.pendingPath(), pending)
}

func (s *Store) ClearPending() error {
	err := os.Remove(s.pendingPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *Store) LoadSettings(settings any) error {
	return readJSON(s.settingsPath(), settings)
}

func (s *Store) SaveSettings(settings any) error {
	return writeJSON(s.settingsPath(), settings)
}

// ClaimPending keeps the original command for recovery but removes it from
// automatic startup processing. The returned name identifies the retained file.
func (s *Store) ClaimPending() (string, error) {
	name := "pending-" + time.Now().Format("20060102-150405.000000000") + ".interrupted.json"
	if err := os.Rename(s.pendingPath(), filepath.Join(s.dir, name)); err != nil {
		return "", err
	}
	return name, nil
}

func (s *Store) CompletePending(name string) error {
	if filepath.Base(name) != name || name == "." {
		return fmt.Errorf("无效的待应用文件名")
	}
	return os.Remove(filepath.Join(s.dir, name))
}
