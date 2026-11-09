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
		{Name: "PATH", Value: `C:\ok\bin;%ENVMAN_TEST_ROOT%\bin;%BASE%\bin;C:\bad\bin`, Type: "REG_EXPAND_SZ"},
		{Name: "RELATIVE", Value: `bin\tools`, Type: "REG_SZ"},
		{Name: "MIXED", Value: `bin\tools;C:\bad`, Type: "REG_SZ"},
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

	for _, tc := range []struct{ name, entry string }{
		{"MISSING", `C:\bad`},
		{"QUOTED", `"C:\bad"`},
		{"TRAILING", `C:\bad`},
		{"LIST", `C:\bad`},
		{"PATH", `C:\bad\bin`},
		{"MIXED", `C:\bad`},
	} {
		if !a.missingEntry(ScopeUser, tc.name, tc.entry) {
			t.Errorf("%s entry %q: missing path not recorded", tc.name, tc.entry)
		}
	}
	for _, tc := range []struct{ name, entry string }{
		{"OK", `C:\ok`},
		{"LIST", `C:\ok`},
		{"PATH", `C:\ok\bin`},
		{"PATH", `%ENVMAN_TEST_ROOT%\bin`},
		{"PATH", `%BASE%\bin`},
		{"RELATIVE", `bin\tools`},
		{"MIXED", `bin\tools`},
		{"COMMAND", `C:\tools\code.exe --wait`},
		{"REF", `%BASE%\bin`},
		{"PROCESS_REF", `%ENVMAN_TEST_ROOT%\bin`},
		{"UNRESOLVED", `%NOPE%\bin`},
		{"LITERAL", `C:\%USERNAME%\bin`},
	} {
		if a.missingEntry(ScopeUser, tc.name, tc.entry) {
			t.Errorf("%s entry %q: non-missing entry was flagged", tc.name, tc.entry)
		}
	}
}

func TestMissingPathEntriesAreShown(t *testing.T) {
	a, tt := newTestApp(t)
	backend := memoryBackend(t, a)
	backend.Replace(ScopeUser, []Variable{{Name: "Path", Value: `C:\ok;C:\bad`, Type: "REG_SZ"}})
	a.reload()
	a.scanMissingPaths(fakeStat(map[string]bool{`C:\ok`: true}))

	tt.Frame()
	if !tt.HasText("1 个路径不存在") {
		t.Fatalf("missing entry count not shown: %q", tt.Texts())
	}
	if !tt.HasText("路径不存在") {
		t.Fatalf("missing entry badge not shown: %q", tt.Texts())
	}
}

func TestScanMissingPathsReplacesStaleResults(t *testing.T) {
	a, _ := newTestApp(t)
	backend := memoryBackend(t, a)
	backend.Replace(ScopeUser, []Variable{{Name: "PATH_ONE", Value: `C:\one;C:\two`, Type: "REG_SZ"}})
	a.reload()

	a.scanMissingPaths(fakeStat(map[string]bool{`C:\one`: true, `C:\two`: true}))
	if a.missingEntry(ScopeUser, "PATH_ONE", `C:\one`) {
		t.Fatal("existing path was flagged")
	}
	a.scanMissingPaths(fakeStat(nil))
	if !a.missingEntry(ScopeUser, "PATH_ONE", `C:\one`) || !a.missingEntry(ScopeUser, "PATH_ONE", `C:\two`) {
		t.Fatal("stale result was kept after re-scan")
	}
}
