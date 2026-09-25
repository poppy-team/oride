// Package statusbar renders the bottom line.
package statusbar

import (
	"strconv"

	"github.com/ori-team/oride/internal/tui/layout"
	"github.com/ori-team/oride/internal/tui/theme"
)

// View is what the status bar needs to draw.
//
// Every field is a plain value rather than a model reference: the bar is a pure
// function of what it is told, which is what lets it be tested at any width
// without a terminal.
type View struct {
	File       string
	Dirty      bool
	Focus      string
	Message    string
	CaretLine  int
	CaretCol   int
	HasCaret   bool
	CursorMode string
	// Theme styles the bar. The zero value renders plain, which is what a
	// no-colour terminal and every test want.
	Theme theme.Theme
}

// markers. The dirty flag is spelled out, not only tinted.
const (
	dirtyMark = "*"
	untitled  = "(sem título)"
	separator = "  "
)

// Render draws the status bar at an exact width.
//
// The caret position is placed at the right edge and everything else is
// truncated to fit what remains, because the position is the field a reader
// looks for at a fixed place — a value that moves is a value that gets missed.
func Render(width int, view View) string {
	if width <= 0 {
		return ""
	}

	right := position(view)
	left := description(view)

	// The right side wins the space it needs: it is short, and it is the field
	// read by position rather than by scanning.
	available := width - layout.Width(right) - 1
	if available < 0 {
		return layout.Pad(layout.Truncate(left, width), width)
	}

	left = layout.Truncate(left, available)
	padding := width - layout.Width(left) - layout.Width(right)
	if padding < 0 {
		padding = 0
	}

	// The style wraps the whole padded line, not just the text: a status bar whose
	// background stopped at the last word would look like a rendering fault.
	bar := layout.Pad(left+spaces(padding)+right, width)
	if view.Dirty {
		return view.Theme.StatusDirty().Render(bar)
	}
	return view.Theme.Status().Render(bar)
}

// description is the left-hand part: what is open, where focus is, and any
// transient message.
func description(view View) string {
	file := view.File
	if file == "" {
		file = untitled
	}
	if view.Dirty {
		file += dirtyMark
	}

	parts := []string{" " + file}
	if view.Focus != "" {
		parts = append(parts, view.Focus)
	}
	if view.Message != "" {
		parts = append(parts, view.Message)
	}
	out := parts[0]
	for _, part := range parts[1:] {
		out += separator + part
	}
	return out
}

// position is the right-hand part: the caret's line and column.
func position(view View) string {
	if !view.HasCaret {
		return " "
	}
	return "Ln " + strconv.Itoa(view.CaretLine) + ", Col " + strconv.Itoa(view.CaretCol) + " "
}

// spaces repeats a space.
func spaces(count int) string {
	if count <= 0 {
		return ""
	}
	out := make([]byte, count)
	for index := range out {
		out[index] = ' '
	}
	return string(out)
}
