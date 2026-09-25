package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// EditorIndent is the indentation resolved for a file.
type EditorIndent struct {
	TabSize      int
	InsertSpaces bool
}

// ResolveIndentForFile reads `.editorconfig` and resolves the indentation for a
// file, falling back to the given values.
//
// This is a deliberately small subset of the format: `indent_style`,
// `indent_size` and `tab_width`, under `[*]`, `*.ext`, `**/*.ext` or an exact
// file name. The full specification is large, and a partial reader that is
// honest about what it reads is better than one that pretends to parse globs it
// does not.
func ResolveIndentForFile(file string, fallback EditorIndent) EditorIndent {
	out := fallback

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(file), "."))
	name := strings.ToLower(filepath.Base(file))

	// Outermost first, so the nearest file is applied last and wins. This is the
	// editorconfig rule, and the reference has it backwards: it applies nearest
	// first, so a repository-wide file overrides a subproject's. Ledger B22.
	for _, dir := range editorconfigChain(file) {
		data, err := os.ReadFile(filepath.Join(dir, ".editorconfig"))
		if err != nil {
			continue
		}
		applyEditorconfig(string(data), ext, name, &out)
	}
	return out
}

// editorconfigChain returns the directories to read, outermost first.
//
// The walk stops at a file that declares `root = true`: that setting exists to
// say nothing above it applies.
func editorconfigChain(file string) []string {
	const maxDepth = 32

	chain := make([]string, 0, maxDepth)
	current := filepath.Dir(file)
	for range maxDepth {
		chain = append(chain, current)

		if data, err := os.ReadFile(filepath.Join(current, ".editorconfig")); err == nil {
			if hasRootTrue(string(data)) {
				break
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	// Reverse: innermost first becomes outermost first.
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain
}

func hasRootTrue(text string) bool {
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if strings.EqualFold(line, "root=true") || strings.EqualFold(line, "root = true") {
			return true
		}
	}
	return false
}

func applyEditorconfig(text, ext, filename string, out *EditorIndent) {
	applies := false

	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			applies = sectionMatches(line[1:len(line)-1], ext, filename)
			continue
		}

		// Properties before any section are global — `root` lives there, and
		// indentation only ever appears under a section.
		if !applies {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.ToLower(strings.TrimSpace(value))

		switch key {
		case "indent_style":
			switch value {
			case "space":
				out.InsertSpaces = true
			case "tab":
				out.InsertSpaces = false
			}
		case "indent_size":
			// `indent_size = tab` means "follow tab_width", which the next key
			// supplies. Anything else must be a number.
			if value == "tab" {
				continue
			}
			if number, err := strconv.Atoi(value); err == nil && number >= 1 {
				out.TabSize = number
			}
		case "tab_width":
			// tab_width describes a literal tab, so it only sets the size when
			// tabs are what get inserted.
			if number, err := strconv.Atoi(value); err == nil && number >= 1 && !out.InsertSpaces {
				out.TabSize = number
			}
		}
	}
}

// sectionMatches reports whether a section applies to a file.
func sectionMatches(pattern, ext, filename string) bool {
	p := strings.TrimSpace(pattern)
	if p == "*" {
		return true
	}
	if rest, ok := strings.CutPrefix(p, "*."); ok {
		return ext == strings.ToLower(rest)
	}
	// Checked after `*.`: a pattern starting with `**/` is not a `*.` pattern,
	// because the two characters after the first star are not `*.`.
	if rest, ok := strings.CutPrefix(p, "**/*."); ok {
		return ext == strings.ToLower(rest)
	}
	return strings.EqualFold(p, filename)
}
