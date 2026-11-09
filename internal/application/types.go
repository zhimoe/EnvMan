package application

import (
	"envman/internal/domain"
	"time"
)

type Snapshot struct {
	ID        string            `json:"id"`
	Time      time.Time         `json:"time"`
	Scope     domain.Scope      `json:"scope"`
	Note      string            `json:"note"`
	Variables []domain.Variable `json:"variables"`
}

type PendingEdits struct {
	Scope     domain.Scope      `json:"scope"`
	Note      string            `json:"note"`
	Focus     string            `json:"focus,omitempty"`
	Before    []domain.Variable `json:"before,omitempty"`
	Variables []domain.Variable `json:"variables"`
}
