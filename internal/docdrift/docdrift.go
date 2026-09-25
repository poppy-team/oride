// Package docdrift keeps documentation checked against the code it describes.
//
// A keymap table copied into a guide is correct on the day it is written and
// wrong the first time a binding changes. This package makes the document a
// derived artifact: the table is generated from the bindings, and a test fails
// when the committed document and the code disagree in either direction.
//
// The same approach applies to the rest of the UI contract — the focus graph, the
// state vocabulary and the degradation rules — which is why they share a package
// rather than each owning a bespoke check.
package docdrift

import (
	"fmt"
	"sort"
	"strings"
)

// TableHeader is the line a generated table starts with, and the anchor the
// comparison uses to find the table inside the document.
const TableHeader = "| Tecla | Ação |"

// KeymapTable renders the canonical chord → action table as Markdown.
//
// The output is deterministic: chords are grouped by domain and sorted inside
// each group, so regenerating an unchanged table produces an unchanged file and
// a diff always means a real change.
func KeymapTable(bindings map[string]string) string {
	grouped := map[string][]binding{}
	for chord, id := range bindings {
		grouped[DomainOf(id)] = append(grouped[DomainOf(id)], binding{Chord: chord, Action: id})
	}

	var out strings.Builder

	for _, domain := range domainOrder {
		items, ok := grouped[domain]
		if !ok {
			continue
		}
		sort.Slice(items, func(i, j int) bool { return items[i].Chord < items[j].Chord })

		// A heading per group, because a flat 98-row table is a wall. The heading
		// is presentation: the check compares the (chord, id) pairs.
		fmt.Fprintf(&out, "### %s\n\n", domain)
		out.WriteString(TableHeader + "\n|---|---|\n")
		for _, item := range items {
			fmt.Fprintf(&out, "| %s | %s |\n", CodeSpan(PrettyChord(item.Chord)), CodeSpan(item.Action))
		}
		out.WriteString("\n")
	}
	return out.String()
}

type binding struct {
	Chord  string
	Action string
}

// domainOrder is the presentation order of the groups.
var domainOrder = []string{
	"Arquivo", "Edição", "Movimento", "Seleção", "Busca",
	"Layout", "Visão", "Terminal", "Git", "LSP", "Navegação", "Descoberta", "Outros",
}

// domainPrefixes maps an action id prefix to its group.
//
// A prefix rule rather than a hand-maintained list of ids: a new action lands in
// a group by its name, and no one has to remember to update a second table. The
// test guards the (chord, id) pairs; the grouping is presentation.
var domainPrefixes = []struct {
	Prefix string
	Domain string
}{
	{"move_", "Movimento"},
	{"select_", "Seleção"},
	{"add_cursor", "Seleção"},
	{"clear_extra_cursors", "Seleção"},
	{"undo", "Edição"},
	{"redo", "Edição"},
	{"insert_", "Edição"},
	{"backspace", "Edição"},
	{"delete_forward", "Edição"},
	{"copy", "Edição"},
	{"paste", "Edição"},
	{"cut", "Edição"},
	{"toggle_comment", "Edição"},
	{"surround", "Edição"},
	{"save", "Arquivo"},
	{"new_tab", "Arquivo"},
	{"open_", "Arquivo"},
	{"close_tab", "Arquivo"},
	{"reload", "Arquivo"},
	{"next_tab", "Arquivo"},
	{"prev_tab", "Arquivo"},
	{"find", "Busca"},
	{"project_find", "Busca"},
	{"project_replace", "Busca"},
	{"replace", "Busca"},
	{"git_", "Git"},
	{"scm", "Git"},
	{"show_diff", "Git"},
	{"lsp_", "LSP"},
	{"diagnostics", "LSP"},
	{"terminal_", "Terminal"},
	{"toggle_terminal", "Terminal"},
	{"run_tasks", "Terminal"},
	{"focus_", "Layout"},
	{"resize_", "Layout"},
	{"split_", "Layout"},
	{"toggle_tree", "Layout"},
	{"toggle_md_preview", "Layout"},
	{"toggle_scm", "Layout"},
	{"jump_", "Navegação"},
	{"command_palette", "Descoberta"},
	{"which_key", "Descoberta"},
	{"help", "Descoberta"},
	{"welcome", "Descoberta"},
	{"buffer_picker", "Descoberta"},
	{"multi_picker", "Descoberta"},
	{"undo_tree", "Descoberta"},
	{"select_theme", "Descoberta"},
	{"select_locale", "Descoberta"},
	{"health_check", "Descoberta"},
	{"toggle_mouse", "Visão"},
	{"toggle_modal", "Visão"},
}

// CodeSpan wraps text as a Markdown code span.
//
// A single backtick would break on content that contains one — and a keybinding
// does: `Ctrl+\“ collides with the delimiter and the row stops parsing as a
// table row at all. The delimiter grows to one longer than the longest run
// inside the content, which is how CommonMark resolves the ambiguity.
func CodeSpan(text string) string {
	longest := 0
	run := 0
	for _, r := range text {
		if r == '`' {
			run++
			if run > longest {
				longest = run
			}
			continue
		}
		run = 0
	}

	delimiter := strings.Repeat("`", longest+1)
	// A space keeps a leading or trailing backtick from being read as part of
	// the delimiter, and is stripped by the renderer.
	if strings.HasPrefix(text, "`") || strings.HasSuffix(text, "`") {
		return delimiter + " " + text + " " + delimiter
	}
	return delimiter + text + delimiter
}

// StripCodeSpan removes the Markdown code-span delimiters from a cell.
//
// The inverse of CodeSpan: it counts the opening run and requires the same run to
// close, so a cell whose content starts with a backtick is not mistaken for an
// unclosed span.
func StripCodeSpan(cell string) (string, bool) {
	opening := 0
	for opening < len(cell) && cell[opening] == '`' {
		opening++
	}
	if opening == 0 {
		return "", false
	}

	delimiter := strings.Repeat("`", opening)
	if len(cell) < 2*opening || !strings.HasSuffix(cell, delimiter) {
		return "", false
	}

	inner := cell[opening : len(cell)-opening]
	if strings.HasPrefix(inner, " ") && strings.HasSuffix(inner, " ") && len(inner) >= 2 {
		inner = inner[1 : len(inner)-1]
	}
	return inner, true
}

// DomainOf groups an action id by its name.
func DomainOf(actionID string) string {
	for _, entry := range domainPrefixes {
		if strings.HasPrefix(actionID, entry.Prefix) {
			return entry.Domain
		}
	}
	return "Outros"
}

// chordSpellings rewrites a canonical chord into the form a reader expects.
var chordSpellings = strings.NewReplacer(
	"ctrl+", "Ctrl+",
	"alt+", "Alt+",
	"shift+", "Shift+",
	"+left", "+←",
	"+right", "+→",
	"+up", "+↑",
	"+down", "+↓",
)

// PrettyChord renders a canonical chord for display.
//
// Display only: the canonical spelling stays the one `keymap.Parse` accepts, so
// the document never becomes a second dialect to parse.
func PrettyChord(chord string) string { return chordSpellings.Replace(chord) }

// prettyToCanonical is the inverse of chordSpellings.
var prettyToCanonical = strings.NewReplacer(
	"Ctrl+", "ctrl+",
	"Alt+", "alt+",
	"Shift+", "shift+",
	"+←", "+left",
	"+→", "+right",
	"+↑", "+up",
	"+↓", "+down",
)

// CanonicalChord converts a displayed chord back to the canonical spelling.
//
// It exists so the drift check can compare the document against the bindings in
// both directions using the same vocabulary the keymap uses.
func CanonicalChord(pretty string) string { return prettyToCanonical.Replace(pretty) }

// Markers delimit the generated region of a document.
const (
	BeginMarker = "<!-- GERADO: keymap -->"
	EndMarker   = "<!-- FIM: keymap -->"
)

// ReplaceGenerated swaps the region between the markers for content.
//
// It returns the document unchanged when either marker is missing, and reports
// that: silently appending a second table would make the check pass while the
// document grew a duplicate.
func ReplaceGenerated(document, content string) (string, error) {
	begin := strings.Index(document, BeginMarker)
	end := strings.Index(document, EndMarker)
	if begin < 0 || end < 0 {
		return "", fmt.Errorf("documento sem os marcadores %s … %s", BeginMarker, EndMarker)
	}
	if end < begin {
		return "", fmt.Errorf("marcador de fim antes do de início")
	}
	return document[:begin+len(BeginMarker)] + "\n" + content + document[end:], nil
}

// GeneratedRegion extracts the text between the markers.
func GeneratedRegion(document string) (string, error) {
	begin := strings.Index(document, BeginMarker)
	end := strings.Index(document, EndMarker)
	if begin < 0 || end < 0 {
		return "", fmt.Errorf("documento sem os marcadores %s … %s", BeginMarker, EndMarker)
	}
	if end < begin {
		return "", fmt.Errorf("marcador de fim antes do de início")
	}
	return document[begin+len(BeginMarker) : end], nil
}
