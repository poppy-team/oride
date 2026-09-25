package lsp

import (
	"encoding/json"
	"strings"
	"unicode/utf16"
)

// Position is a line and a UTF-16 code-unit column, as the protocol defines it.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range is a half-open span.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Severity classifies a diagnostic.
type Severity int

// Severity values, matching the protocol's numbering.
const (
	SeverityError       Severity = 1
	SeverityWarning     Severity = 2
	SeverityInformation Severity = 3
	SeverityHint        Severity = 4
)

// Diagnostic is one problem the server reports.
type Diagnostic struct {
	Range    Range    `json:"range"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Source   string   `json:"source"`
}

// CompletionItem is one suggestion.
type CompletionItem struct {
	Label      string `json:"label"`
	Detail     string `json:"detail"`
	InsertText string `json:"insertText"`
}

// Hover is the documentation shown for a position.
type Hover struct {
	Contents MarkupContent `json:"contents"`
}

// MarkupContent carries hover text in either plain text or Markdown.
type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// Location is a span in a file.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// TextEdit is one change the server asks for.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// UTF16Column converts a column counted in Unicode scalars to the UTF-16 code
// units the protocol counts.
//
// A scalar outside the Basic Multilingual Plane occupies two units, which is why
// this cannot be a byte count, a rune count, or a length in either.
func UTF16Column(line string, characterColumn int) int {
	units := 0
	count := 0
	for _, r := range line {
		if count == characterColumn {
			break
		}
		units += len(utf16.Encode([]rune{r}))
		count++
	}
	return units
}

// CharacterColumn converts a UTF-16 column back to a scalar column.
//
// It returns false for a position that splits a surrogate pair. Rejecting is the
// only honest answer: the position names half a character, and rounding it either
// way would move the caret to a place the server did not mean.
func CharacterColumn(line string, utf16Column int) (int, bool) {
	target := utf16Column
	units := 0
	column := 0

	for _, r := range line {
		if units == target {
			return column, true
		}
		units += len(utf16.Encode([]rune{r}))
		if units > target {
			return 0, false
		}
		column++
	}
	if units == target {
		return column, true
	}
	return 0, false
}

// ParseDiagnostics reads a `textDocument/publishDiagnostics` parameter.
func ParseDiagnostics(params json.RawMessage) ([]Diagnostic, error) {
	var payload struct {
		Diagnostics []Diagnostic `json:"diagnostics"`
	}
	if err := json.Unmarshal(params, &payload); err != nil {
		return nil, err
	}
	out := make([]Diagnostic, 0, len(payload.Diagnostics))
	for _, diagnostic := range payload.Diagnostics {
		if diagnostic.Severity == 0 {
			// Servers may omit the severity; a diagnostic is a problem either
			// way, and dropping it would hide a real finding.
			diagnostic.Severity = SeverityWarning
		}
		out = append(out, diagnostic)
	}
	return out, nil
}

// ParseCompletion reads a `textDocument/completion` result.
//
// The result may be a bare list or a `{items: [...]}` object; both are legal and
// servers differ, so accepting one shape would silently produce no completions
// against half the servers.
func ParseCompletion(raw json.RawMessage) ([]CompletionItem, error) {
	var list []CompletionItem
	if err := json.Unmarshal(raw, &list); err == nil {
		return list, nil
	}

	var wrapper struct {
		Items []CompletionItem `json:"items"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Items, nil
}

// ParseHover reads a hover result, flattening every legal shape to text.
//
// `contents` has three forms in the wild: a markup object, a bare string, and a
// list of marked-up strings from before the object existed. The shape is decided
// from the value, not from the envelope, so accepting only one form would show
// no documentation against servers using another.
func ParseHover(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", nil
	}

	var envelope struct {
		Contents json.RawMessage `json:"contents"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		// Some servers send the contents directly, without the envelope.
		return parseHoverContents(raw)
	}
	return parseHoverContents(envelope.Contents)
}

func parseHoverContents(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", nil
	}

	var markup MarkupContent
	if err := json.Unmarshal(raw, &markup); err == nil && markup.Value != "" {
		return markup.Value, nil
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}

	var marked []struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(raw, &marked); err == nil {
		parts := make([]string, 0, len(marked))
		for _, part := range marked {
			parts = append(parts, part.Value)
		}
		return strings.Join(parts, "\n"), nil
	}
	return "", nil
}

// ParseLocations reads a definition result, which may be one location or a list.
func ParseLocations(raw json.RawMessage) ([]Location, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var single Location
	if err := json.Unmarshal(raw, &single); err == nil && single.URI != "" {
		return []Location{single}, nil
	}

	var list []Location
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, err
	}
	return list, nil
}
