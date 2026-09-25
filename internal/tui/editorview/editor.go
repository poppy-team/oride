// Package editorview renders the text viewport.
package editorview

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/ori-team/oride/internal/tui/layout"
	"github.com/ori-team/oride/internal/tui/theme"
)

// View is what the viewport needs to draw.
//
// Lines arrive already extracted rather than as a buffer: the surface must not
// reach into the document, and a viewport does not need the whole file to draw
// one screen.
type View struct {
	Lines        []string
	Offset       int
	Caret        Position
	HasCaret     bool
	Selection    Selection
	Matches      []Match
	CurrentMatch int
	Gutter       bool
	Focused      bool
	// Theme dims the gutter. The zero value renders plain.
	Theme theme.Theme
}

// TotalLines is the document's length, which the gutter is sized from.
func (v View) TotalLines() int { return v.Offset + len(v.Lines) }

// Match is one search result, in document coordinates.
type Match struct {
	Start Position
	End   Position
}

// Position is a caret location, zero-based, in document coordinates.
type Position struct {
	Line   int
	Column int
}

// Selection is the selected range, in document coordinates.
type Selection struct {
	Start Position
	End   Position
	Empty bool
}

// isCollapsed reports whether the range covers no cells.
//
// A zero-value Selection has Empty false and Start equal to End, so a surface that
// trusted the flag alone would paint a collapsed range — and, worse, would take the
// painting path for a line that needed the plain one, losing the truncation marker
// on every long line. The surface decides from the range, not from the caller.
func (s Selection) isCollapsed() bool {
	return s.Start.Line == s.End.Line && s.Start.Column == s.End.Column
}

// caretMarker sits in the gutter on the line the caret is on.
//
// The caret is marked in the gutter rather than painted over a character: a
// terminal cursor is a real thing the runtime draws, and a surface that also
// drew its own would show two. And the marker is text, so it survives a no-colour
// terminal — where a tinted caret would be invisible and the position unknowable.
const caretMarker = ">"

// gutterPadding is the space between the line numbers and the text.
const gutterPadding = 1

// Render draws the viewport, one string per line, each exactly width cells.
func Render(width, height int, view View) []string {
	if width <= 0 || height <= 0 {
		return nil
	}

	gutterWidth := 0
	if view.Gutter {
		gutterWidth = gutterSize(view.TotalLines()) + gutterPadding
	}
	if gutterWidth >= width {
		// The gutter would leave no room for text, so it is what gives way.
		gutterWidth = 0
	}

	rows := make([]string, 0, height)
	for index := range view.Lines {
		if len(rows) == height {
			break
		}
		rows = append(rows, renderLine(width, gutterWidth, index, view))
	}

	// Padding rows keep the frame exactly as tall as asked. A viewport that came
	// back short would let whatever is below it move up.
	for len(rows) < height {
		blank := ""
		if gutterWidth > 0 {
			blank = layout.Pad("", gutterWidth)
		}
		rows = append(rows, layout.Pad(blank, width))
	}
	return rows
}

// renderLine draws one line with its gutter.
//
// index is a position in the visible slice; the gutter shows the document line
// number, which is what a reader expects to see beside the text.
func renderLine(width, gutterWidth, index int, view View) string {
	text := view.Lines[index]
	number := view.Offset + index

	if gutterWidth == 0 {
		// Without a gutter there is no line number, but there is still a
		// selection: returning the padded text here skipped the painting
		// entirely, and every line came out plain.
		return paintSelection(text, number, width, view)
	}

	marker := " "
	if view.HasCaret && number == view.Caret.Line {
		marker = caretMarker
	}
	gutter := marker + layout.Pad(strconv.Itoa(number+1), gutterWidth-1)

	// The gutter is styled and the text is not, so the line stays selectable by
	// copy: a background over the text would be captured by a terminal copy.
	body := paintSelection(text, number, width-gutterWidth, view)
	return layout.Pad(view.Theme.Gutter().Render(gutter), gutterWidth) + body
}

// cell is what a column of the line is, before any styling.
//
// A cell array rather than a chain of conditions per rune: selection and matches
// overlap, and deciding the winner once per cell is what makes the precedence
// explicit instead of an accident of the order the styles were applied.
type cell int

const (
	cellPlain cell = iota
	cellMatch
	cellCurrentMatch
	cellSelection
)

// paintSelection draws the line with its matches and selection tinted.
//
// Precedence is decided per cell: the selection wins over a match, because a
// reader who has selected text is asking about the selection.
func paintSelection(text string, line, width int, view View) string {
	kinds := classify(text, line, width, view)
	if !hasStyled(kinds) {
		return layout.Pad(text, width)
	}

	var out strings.Builder
	position := 0
	for position < width {
		kind := kinds[position]
		run := 0
		for position+run < width && kinds[position+run] == kind {
			run++
		}

		segment := layout.Pad(layout.Slice(text, position, run), run)
		out.WriteString(styleFor(kind, view.Theme).Render(segment))
		position += run
	}
	return out.String()
}

// classify decides what each column of the line is.
//
// The array is in cells, matching the unit the layout measures in: a wide
// character occupies two of them, and marking one of the two would tint half a
// character.
func classify(text string, line, width int, view View) []cell {
	kinds := make([]cell, width)

	for _, match := range matchesOn(line, view.Matches) {
		mark(kinds, text, match.Start, match.End, cellMatch)
	}
	if view.CurrentMatch >= 0 && view.CurrentMatch < len(view.Matches) {
		current := view.Matches[view.CurrentMatch]
		if current.Start.Line <= line && line <= current.End.Line {
			mark(kinds, text, current.Start, current.End, cellCurrentMatch)
		}
	}

	if !view.Selection.Empty && !view.Selection.isCollapsed() && lineIsSelected(line, view.Selection) {
		start, end := selectedColumns(line, text, view.Selection)
		for column := start; column < end && column < width; column++ {
			kinds[column] = cellSelection
		}
	}
	return kinds
}

// matchesOn lists the matches touching a line.
func matchesOn(line int, matches []Match) []Match {
	out := make([]Match, 0, 2)
	for _, match := range matches {
		if match.Start.Line <= line && line <= match.End.Line {
			out = append(out, match)
		}
	}
	return out
}

// mark paints the columns of one range.
func mark(kinds []cell, text string, start, end Position, kind cell) {
	line := start.Line
	from, to := selectedColumns(line, text, Selection{Start: start, End: end})
	for column := from; column < to && column < len(kinds); column++ {
		kinds[column] = kind
	}
}

// hasStyled reports whether anything on the line needs styling.
//
// The plain path stays the common one: most lines of most frames have neither a
// match nor a selection, and building styled output for them would cost escape
// sequences on every row.
func hasStyled(kinds []cell) bool {
	for _, kind := range kinds {
		if kind != cellPlain {
			return true
		}
	}
	return false
}

// styleFor maps a kind onto the style that draws it.
func styleFor(kind cell, theme theme.Theme) lipgloss.Style {
	switch kind {
	case cellSelection:
		return theme.Selection()
	case cellCurrentMatch:
		return theme.CurrentMatch()
	case cellMatch:
		return theme.Match()
	}
	return lipgloss.NewStyle()
}

// lineIsSelected reports whether the selection touches a line at all.
func lineIsSelected(line int, selection Selection) bool {
	return line >= selection.Start.Line && line <= selection.End.Line
}

// selectedColumns is the selected range on one line, in cells.
//
// The first and last lines of a range are partial; every line between them is
// covered whole. A column past the end of the line would make the padding
// arithmetic produce a negative width, so it is clamped.
func selectedColumns(line int, text string, selection Selection) (start, end int) {
	switch {
	case line == selection.Start.Line && line == selection.End.Line:
		start, end = selection.Start.Column, selection.End.Column
	case line == selection.Start.Line:
		start, end = selection.Start.Column, layout.Width(text)
	case line == selection.End.Line:
		start, end = 0, selection.End.Column
	default:
		start, end = 0, layout.Width(text)
	}

	lineWidth := layout.Width(text)
	start = min(max(0, start), lineWidth)
	end = min(max(start, end), lineWidth)
	return start, end
}

// gutterSize is the width of the widest line number the file will reach.
//
// Derived from the document's length rather than from the visible slice: numbering
// that re-widened as the reader scrolled would shift every line of text sideways.
func gutterSize(totalLines int) int {
	if totalLines <= 1 {
		return 1
	}
	return len(strconv.Itoa(totalLines))
}
