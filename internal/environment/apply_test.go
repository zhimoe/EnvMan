package environment

import (
	"errors"
	"slices"
	"testing"

	"envman/internal/domain"
)

func TestApplyDiffFailureAndBroadcast(t *testing.T) {
	failure := errors.New("injected write failure")
	for _, tc := range []struct {
		name          string
		failAt        int
		wantUpserts   int
		wantDeletes   int
		wantBroadcast int
	}{
		{"first write fails", 1, 0, 0, 0},
		{"second write fails", 2, 1, 0, 1},
		{"first delete fails", 3, 2, 0, 1},
		{"second delete fails", 4, 2, 1, 1},
		{"success", 0, 2, 2, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls, broadcasts := 0, 0
			var completed []string
			operation := func(name string) error {
				calls++
				if calls == tc.failAt {
					return failure
				}
				completed = append(completed, name)
				return nil
			}
			err := applyDiff([]domain.Variable{{Name: "A"}, {Name: "B"}}, []string{"C", "D"}, func(v domain.Variable) error { return operation(v.Name) }, operation, func() { broadcasts++ })
			if broadcasts != tc.wantBroadcast {
				t.Fatalf("broadcasts = %d", broadcasts)
			}
			if tc.failAt == 0 {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			result, ok := errors.AsType[*ApplyError](err)
			if !ok || !errors.Is(err, failure) {
				t.Fatalf("error = %v", err)
			}
			if len(result.Upserts) != tc.wantUpserts || len(result.Deletions) != tc.wantDeletes {
				t.Fatalf("result = %+v", result)
			}
			var reported []string
			for _, v := range result.Upserts {
				reported = append(reported, v.Name)
			}
			reported = append(reported, result.Deletions...)
			if !slices.Equal(reported, completed) {
				t.Fatalf("reported %v, completed %v", reported, completed)
			}
			if calls != tc.failAt {
				t.Fatalf("continued after failure: %d calls", calls)
			}
		})
	}
}
