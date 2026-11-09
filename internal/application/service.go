// Package application coordinates environment use cases without UI or platform dependencies.
package application

import (
	"envman/internal/domain"
	"fmt"
)

type Environment interface {
	Load(domain.Scope) ([]domain.Variable, error)
	ApplyChanges(domain.Scope, []domain.Variable, []domain.Variable) error
}

type Backups interface {
	SaveSnapshot(domain.Scope, string, []domain.Variable) (Snapshot, error)
}

type Change struct {
	Scope   domain.Scope
	Before  []domain.Variable
	Desired []domain.Variable
	Note    string
}

type Result struct {
	Snapshot Snapshot
	Desired  []domain.Variable
}

type Service struct {
	environment Environment
	backups     Backups
}

func New(environment Environment, backups Backups) *Service {
	return &Service{environment: environment, backups: backups}
}

func (s *Service) Snapshot(scope domain.Scope, note string) (Snapshot, error) {
	variables, err := s.environment.Load(scope)
	if err != nil {
		return Snapshot{}, err
	}
	return s.backups.SaveSnapshot(scope, note, variables)
}

// Apply always creates a recovery snapshot before attempting any writes.
// A result with a snapshot is returned even when writes partially fail.
func (s *Service) Apply(change Change) (Result, error) {
	if err := domain.Validate(change.Scope, change.Desired); err != nil {
		return Result{}, err
	}
	snapshot, err := s.Snapshot(change.Scope, change.Note)
	if err != nil {
		return Result{}, fmt.Errorf("创建备份失败，未应用更改：%w", err)
	}
	desired := domain.Clone(change.Desired)
	if change.Before != nil {
		desired = domain.MergeChanges(snapshot.Variables, change.Before, desired)
	}
	result := Result{Snapshot: snapshot, Desired: desired}
	return result, s.environment.ApplyChanges(change.Scope, snapshot.Variables, desired)
}
