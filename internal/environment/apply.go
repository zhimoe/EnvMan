package environment

import (
	"fmt"

	"envman/internal/domain"
)

// ApplyError records operations completed before a write failed. Completed
// operations are not rolled back, and callers must reload the environment.
type ApplyError struct {
	Upserts   []domain.Variable
	Deletions []string
	Err       error
}

func (e *ApplyError) Error() string { return e.Err.Error() }
func (e *ApplyError) Unwrap() error { return e.Err }
func (e *ApplyError) Partial() bool { return len(e.Upserts)+len(e.Deletions) > 0 }

// applyDiff broadcasts even when a later operation fails after earlier writes.
func applyDiff(upserts []domain.Variable, deletions []string, set func(domain.Variable) error, remove func(string) error, broadcast func()) error {
	result := &ApplyError{}
	defer func() {
		if result.Partial() {
			broadcast()
		}
	}()
	for _, v := range upserts {
		if err := set(v); err != nil {
			result.Err = fmt.Errorf("写入 %s：%w", v.Name, err)
			return result
		}
		result.Upserts = append(result.Upserts, v)
	}
	for _, name := range deletions {
		if err := remove(name); err != nil {
			result.Err = fmt.Errorf("删除 %s：%w", name, err)
			return result
		}
		result.Deletions = append(result.Deletions, name)
	}
	return nil
}
