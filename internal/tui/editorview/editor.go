// Package editorview renders the text viewport.
package editorview

import (
	"strconv"

	"github.com/ori-team/oride/internal/tui/layout"
	"github.com/ori-team/oride/internal/tui/theme"
)

// View is what the viewport needs to draw.
//
// Lines arrive already extracted rather than as a buffer: the surface must not
// reach into the document, and a viewport does not need the whole file to draw
// one screen.
type View struct {
	Lines     []string
	Offset    int
	Caret     Position
	HasCaret  bool
	Selection Selection
	Gutter    bool
	Focused   bool
	// Theme dims the gutter. The zero value renders plain.
	Theme theme.Theme
}

// TotalLines is the document's length, which the gutter is sized from.
func (v View) TotalLines() int { return v.Offset + len(v.Lines) }

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

// paintSelection draws the line with the selected cells tinted.
//
// Only the cells inside the selection are styled, and the row is padded outside
// them: a background painted across the whole row would claim the selection
// extends to the edge of the screen, which it does not.
func paintSelection(text string, line, width int, view View) string {
	if view.Selection.Empty || view.Selection.isCollapsed() || !lineIsSelected(line, view.Selection) {
		return layout.Pad(text, width)
	}

	start, end := selectedColumns(line, text, view.Selection)

	before := layout.Slice(text, 0, start)
	inside := layout.Slice(text, start, end-start)
	after := layout.Slice(text, start+layout.Width(inside), width-start-layout.Width(inside))

	return layout.Pad(before, start) +
		view.Theme.Selection().Render(layout.Pad(inside, end-start)) +
		layout.Pad(after, width-end)
}

// lineIsSelected reports whether the selection touches a line at all.
func lineIsSelected(line int, selection Selection) bool {
	return line >= selection.Start.Line && line <= selection.End.Line
}

// selectedColumns is the selected range on one line, in cells.
//
// The first and last lines of a selection are partial; every line between them is
// selected whole. The end column is exclusive, and on the first line the range
// starts where the selection starts.
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

	// A column past the end of the line would make the padding arithmetic produce
	// a negative width.
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
