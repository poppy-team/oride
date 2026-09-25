// Package editorview renders the text viewport.
package editorview

import (
	"strconv"

	"github.com/ori-team/oride/internal/tui/layout"
)

// View is what the viewport needs to draw.
//
// Lines arrive already extracted rather than as a buffer: the surface must not
// reach into the document, and a viewport does not need the whole file to draw
// one screen.
type View struct {
	Lines    []string
	Offset   int
	Caret    Position
	HasCaret bool
	Gutter   bool
	Focused  bool
}

// TotalLines is the document's length, which the gutter is sized from.
func (v View) TotalLines() int { return v.Offset + len(v.Lines) }

// Position is a caret location, zero-based.
type Position struct {
	Line   int
	Column int
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
		return layout.Pad(text, width)
	}

	marker := " "
	if view.HasCaret && number == view.Caret.Line {
		marker = caretMarker
	}
	gutter := marker + layout.Pad(strconv.Itoa(number+1), gutterWidth-1)

	return layout.Pad(gutter+layout.Pad(text, width-gutterWidth), width)
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
