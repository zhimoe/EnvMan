package app

import (
	"errors"
	"io/fs"
	"os"
	"strings"
)

// scanMissingPaths records the single-valued absolute paths that do not
// exist. It runs once per reload on the main thread, so the table can paint
// missing values without touching the file system while rendering.
func (a *App) scanMissingPaths(stat func(string) error) {
	missing := make(map[string]bool)
	for _, scope := range []Scope{ScopeUser, ScopeSystem} {
		for _, v := range a.variables(scope) {
			if path, ok := a.checkablePath(v); ok {
				if errors.Is(stat(path), fs.ErrNotExist) {
					missing[v.Key(scope)] = true
				}
			}
		}
	}
	a.missingPaths = missing
}

// checkablePath returns the expanded path of a variable whose whole value is
// one absolute Windows path. Lists, relative paths, command lines, and values
// whose references cannot be resolved are left alone.
func (a *App) checkablePath(v Variable) (string, bool) {
	entries := splitValue(v.Value)
	if len(entries) != 1 || entries[0] == "" {
		return "", false
	}
	value := entries[0]
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
