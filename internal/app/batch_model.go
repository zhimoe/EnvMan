package app

import (
	"slices"
	"strings"
)

type batchInfo struct {
	original   []string
	added      []string
	removed    []string
	reordered  bool
	hasChanges bool
	duplicates []string
}

func batchParse(text string) (entries []string, ignored int) {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	for p := range strings.SplitSeq(strings.ReplaceAll(text, ";", "\n"), "\n") {
		p = strings.TrimSpace(p)
		if p == "" {
			ignored++
			continue
		}
		entries = append(entries, p)
	}
	return entries, ignored
}

func normalizeEntry(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func analyzeBatch(originalValue, text string) batchInfo {
	original := splitValue(originalValue)
	if originalValue == "" {
		original = nil
	}
	entries, _ := batchParse(text)
	oldKeys := make(map[string]bool)
	for _, o := range original {
		oldKeys[normalizeEntry(o)] = true
	}
	newKeys := make(map[string]bool)
	for _, n := range entries {
		newKeys[normalizeEntry(n)] = true
	}
	info := batchInfo{original: original}
	for _, e := range entries {
		if !oldKeys[normalizeEntry(e)] {
			info.added = append(info.added, e)
		}
	}
	for _, o := range original {
		if !newKeys[normalizeEntry(o)] {
			info.removed = append(info.removed, o)
		}
	}
	commonOld := make([]string, 0)
	seenOld := make(map[string]bool)
	for _, o := range original {
		key := normalizeEntry(o)
		if newKeys[key] && !seenOld[key] {
			commonOld = append(commonOld, key)
			seenOld[key] = true
		}
	}
	commonNew := make([]string, 0)
	seenNew := make(map[string]bool)
	for _, n := range entries {
		key := normalizeEntry(n)
		if oldKeys[key] && !seenNew[key] {
			commonNew = append(commonNew, key)
			seenNew[key] = true
		}
	}
	info.reordered = strings.Join(commonOld, "\x00") != strings.Join(commonNew, "\x00")
	// Saving compares exact values; path duplicate detection remains semantic.
	info.hasChanges = !slices.Equal(original, entries)
	seen := make(map[string]bool)
	for _, n := range entries {
		k := normalizeEntry(n)
		if seen[k] {
			info.duplicates = append(info.duplicates, n)
		}
		seen[k] = true
	}
	return info
}

func containsFold(list []string, s string) bool {
	for _, v := range list {
		if normalizeEntry(v) == normalizeEntry(s) {
			return true
		}
	}
	return false
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, s := range in {
		if !seen[normalizeEntry(s)] {
			seen[normalizeEntry(s)] = true
			out = append(out, s)
		}
	}
	return out
}

func firstN(in []string, n int) []string {
	if len(in) <= n {
		return in
	}
	return in[:n]
}
