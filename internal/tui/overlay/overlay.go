// Package overlay renders everything that floats above the surfaces.
//
// One package rather than four, because the four share the same chrome: a
// filtered list with a title and a hint, or a centred box. Splitting them would
// mean either three copies of the list renderer or a fourth package holding it —
// and the contract asks for cohesion before it asks for symmetry.
package overlay

import (
	"strings"

	"github.com/ori-team/oride/internal/tui/component"
	"github.com/ori-team/oride/internal/tui/layout"
)

// Kind says which overlay is showing.
type Kind int

// The overlays. None means the surfaces receive input.
const (
	None Kind = iota
	Palette
	WhichKey
	Help
	Welcome
	QuitConfirm
	CloseConfirm
)

// Captures reports whether this overlay takes every keystroke.
//
// Every overlay does. The first rule of docs/ui-ux/focus-graph.md is that an
// active overlay captures all input: while one is open, neither the focused
// surface nor the global keymap is consulted. Anything else would let a key typed
// into a filter also edit the buffer behind it.
func (k Kind) Captures() bool { return k != None }

// Active reports whether an overlay is open.
func (k Kind) Active() bool { return k != None }

// Item is one row of a list overlay.
type Item struct {
	Label    string
	Detail   string
	Selected bool
}

// ListView is what a list overlay needs to draw.
type ListView struct {
	Title  string
	Hint   string
	Items  []Item
	Scroll int
	// Empty is shown when there is nothing to list, so an empty palette is
	// distinguishable from one that failed to load.
	Empty string
}

// layout constants. The box leaves a margin so the surface behind stays visible.
const (
	minWidth   = 20
	maxWidth   = 76
	marginCols = 4
	marginRows = 2
	marker     = "▶ "
	padding    = 1
)

// List renders a list overlay at the given size.
func List(width, height int, view ListView) []string {
	if width <= 0 || height <= 0 {
		return nil
	}

	inner := listInnerWidth(width)
	rows := make([]string, 0, height)

	if view.Title != "" {
		rows = append(rows, layout.Pad(" "+view.Title, inner))
	}

	if len(view.Items) == 0 {
		message := view.Empty
		if message == "" {
			message = "(nada)"
		}
		rows = append(rows, layout.Pad(" "+message, inner))
		return fit(rows, inner, height, view.Hint)
	}

	start, end := component.Window(len(view.Items), listRows(height, view), 0, view.Scroll)
	for index := start; index < end; index++ {
		item := view.Items[index]
		rows = append(rows, layout.Pad(" "+row(item, inner), inner))
	}
	return fit(rows, inner, height, view.Hint)
}

// Box renders a centred box: a title, some lines, and a hint.
func Box(width, height int, title string, lines []string, hint string) []string {
	if width <= 0 || height <= 0 {
		return nil
	}

	inner := listInnerWidth(width)
	rows := make([]string, 0, height)
	if title != "" {
		rows = append(rows, layout.Pad(" "+title, inner))
	}
	for _, line := range lines {
		rows = append(rows, layout.Pad(" "+line, inner))
	}
	return fit(rows, inner, height, hint)
}

// row draws one item, marking the selected one in text as well as tint.
func row(item Item, inner int) string {
	cursor := "  "
	if item.Selected {
		cursor = marker
	}

	text := item.Label
	if item.Detail != "" {
		text += "  " + item.Detail
	}
	return layout.Truncate(cursor+text, inner-1)
}

// listRows is how many item rows fit, once the title and the hint are accounted
// for.
func listRows(height int, view ListView) int {
	reserved := 0
	if view.Title != "" {
		reserved++
	}
	if view.Hint != "" {
		reserved++
	}
	return max(1, height-reserved)
}

// fit pads the rows to the height and adds the hint at the bottom.
//
// The hint goes last so it always occupies the same row, whatever the content:
// a hint that moves is a hint nobody finds twice.
func fit(rows []string, inner, height int, hint string) []string {
	reserved := 0
	if hint != "" {
		reserved = 1
	}
	for len(rows) < height-reserved {
		rows = append(rows, layout.Pad("", inner))
	}
	if hint != "" {
		rows = append(rows, layout.Pad(" "+hint, inner))
	}
	for len(rows) < height {
		rows = append(rows, layout.Pad("", inner))
	}
	if len(rows) > height {
		rows = rows[:height]
	}
	return rows
}

// listInnerWidth is the content width of an overlay at a given terminal width.
func listInnerWidth(width int) int {
	return min(max(width-marginCols, minWidth), maxWidth)
}

// Region returns where the overlay lands, centred over the given area.
func Region(area layout.Region, width, height int) layout.Region {
	inner := listInnerWidth(area.Width)
	if height > area.Height {
		height = area.Height
	}
	return layout.Region{
		X:      area.X + max(0, (area.Width-inner)/2),
		Y:      area.Y + max(0, (area.Height-height)/2),
		Width:  inner,
		Height: height,
	}
}

// Frame is the overlay's preferred height for a list of n items.
func Frame(items int) int { return items + marginRows }

// Joins rows for a test or a caller that wants one string.
func Joins(rows []string) string { return strings.Join(rows, "\n") }
