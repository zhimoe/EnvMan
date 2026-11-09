package app

import (
	"errors"
	"io/fs"
	"os"
	"strings"
)

// scanMissingPaths records the absolute path entries that do not exist. It
// runs once per reload on the main thread, so the table can paint missing
// entries without touching the file system while rendering. Every entry is
// checked, so a list like Path flags only the entries that are gone.
func (a *App) scanMissingPaths(stat func(string) error) {
	missing := make(map[string]bool)
	for _, scope := range []Scope{ScopeUser, ScopeSystem} {
		for _, v := range a.variables(scope) {
			for _, entry := range splitValue(v.Value) {
				path, ok := a.checkablePath(v, entry)
				if !ok {
					continue
				}
				if errors.Is(stat(path), fs.ErrNotExist) {
					missing[missingEntryKey(scope, v.Name, entry)] = true
				}
			}
		}
	}
	a.missingPaths = missing
}

// missingEntryKey identifies one entry of a variable's value; the same
// variable can have some entries missing and others present.
func missingEntryKey(scope Scope, name, entry string) string {
	return varKey(scope, name) + "\x00" + entry
}

func (a *App) missingEntry(scope Scope, name, entry string) bool {
	return a.missingPaths[missingEntryKey(scope, name, entry)]
}

// checkablePath returns the expanded path of an entry that is one whole
// absolute Windows path. Relative paths, command lines, and entries whose
// references cannot be resolved are left alone.
func (a *App) checkablePath(v Variable, entry string) (string, bool) {
	if entry == "" {
		return "", false
	}
	value := entry
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = strings.TrimSpace(value[1 : len(value)-1])
	} else if strings.Contains(value, `"`) {
		return "", false
	}
	if strings.Contains(value, "%") {
		// Windows expands references only in REG_EXPAND_SZ values.
		if valueType(v.Type, v.Value) == "REG_SZ" {
			return "", false
		}
		value = expandProcessRefs(a.expandValue(value))
		if strings.Contains(value, "%") {
			return "", false
		}
	}
	if !isAbsoluteWindowsPath(value) || hasPathBadChars(value) || hasFlagArgs(value) {
		return "", false
	}
	return value, true
}

// expandProcessRefs fills references the registry snapshot does not hold,
// such as %USERPROFILE%, from the process environment.
func expandProcessRefs(value string) string {
	return refPattern.ReplaceAllStringFunc(value, func(ref string) string {
		if expanded, ok := os.LookupEnv(ref[1 : len(ref)-1]); ok {
			return expanded
		}
		return ref
	})
}

func isAbsoluteWindowsPath(s string) bool {
	if len(s) < 3 {
		return false
	}
	c := s[0]
	return (c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z') &&
		s[1] == ':' && (s[2] == '\\' || s[2] == '/')
}

func hasPathBadChars(s string) bool {
	for _, r := range s {
		switch r {
		case ';', '*', '?', '<', '>', '|', '"':
			return true
		}
		if r < 0x20 {
			return true
		}
	}
	return false
}

// hasFlagArgs treats a space followed by - or / as command-line arguments,
// so "C:\tools\code.exe --wait" is not mistaken for a path.
func hasFlagArgs(s string) bool {
	for i := 1; i < len(s); i++ {
		if s[i-1] == ' ' && (s[i] == '-' || s[i] == '/') {
			return true
		}
	}
	return false
}

func statPath(path string) error {
	_, err := os.Stat(path)
	return err
}
