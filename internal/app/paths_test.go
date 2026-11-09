package app

import (
	"io/fs"
	"testing"
)

func fakeStat(exists map[string]bool) func(string) error {
	return func(path string) error {
		if exists[path] {
			return nil
		}
		return fs.ErrNotExist
	}
}

func TestScanMissingPaths(t *testing.T) {
	t.Setenv("ENVMAN_TEST_ROOT", `C:\ok`)
	a, _ := newTestApp(t)
	backend := memoryBackend(t, a)
	backend.Replace(ScopeUser, []Variable{
		{Name: "OK", Value: `C:\ok`, Type: "REG_SZ"},
		{Name: "MISSING", Value: `C:\bad`, Type: "REG_SZ"},
		{Name: "QUOTED", Value: `"C:\bad"`, Type: "REG_SZ"},
		{Name: "TRAILING", Value: `C:\bad;`, Type: "REG_SZ"},
		{Name: "LIST", Value: `C:\ok;C:\bad`, Type: "REG_SZ"},
		{Name: "RELATIVE", Value: `bin\tools`, Type: "REG_SZ"},
		{Name: "COMMAND", Value: `C:\tools\code.exe --wait`, Type: "REG_SZ"},
		{Name: "BASE", Value: `C:\ok`, Type: "REG_SZ"},
		{Name: "REF", Value: `%BASE%\bin`, Type: "REG_EXPAND_SZ"},
		{Name: "PROCESS_REF", Value: `%ENVMAN_TEST_ROOT%\bin`, Type: "REG_EXPAND_SZ"},
		{Name: "UNRESOLVED", Value: `%NOPE%\bin`, Type: "REG_EXPAND_SZ"},
		{Name: "LITERAL", Value: `C:\%USERNAME%\bin`, Type: "REG_SZ"},
	})
	a.reload()
	a.scanMissingPaths(fakeStat(map[string]bool{
		`C:\ok`:     true,
		`C:\ok\bin`: true,
	}))

	for _, name := range []string{"MISSING", "QUOTED", "TRAILING"} {
		if !a.missingPaths[varKey(ScopeUser, name)] {
			t.Errorf("%s: missing path not recorded", name)
		}
	}
	for _, name := range []string{"OK", "LIST", "RELATIVE", "COMMAND", "REF", "PROCESS_REF", "UNRESOLVED", "LITERAL"} {
		if a.missingPaths[varKey(ScopeUser, name)] {
			t.Errorf("%s: non-missing value was flagged", name)
		}
	}
}

func TestScanMissingPathsReplacesStaleResults(t *testing.T) {
	a, _ := newTestApp(t)
	backend := memoryBackend(t, a)
	backend.Replace(ScopeUser, []Variable{{Name: "PATH_ONE", Value: `C:\one`, Type: "REG_SZ"}})
	a.reload()

	a.scanMissingPaths(fakeStat(map[string]bool{`C:\one`: true}))
	if a.missingPaths[varKey(ScopeUser, "PATH_ONE")] {
		t.Fatal("existing path was flagged")
	}
	a.scanMissingPaths(fakeStat(nil))
	if !a.missingPaths[varKey(ScopeUser, "PATH_ONE")] {
		t.Fatal("stale result was kept after re-scan")
	}
}
