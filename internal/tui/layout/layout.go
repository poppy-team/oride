// Package layout computes where each surface goes, and how text fits inside it.
//
// It is pure geometry: no terminal, no Bubble Tea, no model. The model says what
// it wants shown; this package answers where it lands. Keeping it separate means
// the layout rules in docs/ui-ux/layout.md are testable at every width without
// rendering a frame.
package layout

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
)

// WidthClass is how much room the terminal offers.
type WidthClass int

// The classes. Boundaries are inclusive at the bottom: 80 is Standard, 120 is
// Wide. Stated once, because an off-by-one here moves every surface on screen.
const (
	Compact WidthClass = iota
	Standard
	Wide
)

// CompactMax and StandardMax are the last widths of their class.
const (
	CompactMax  = 79
	StandardMax = 119
)

// Classify reports the class of a terminal width.
func Classify(width int) WidthClass {
	switch {
	case width <= CompactMax:
		return Compact
	case width <= StandardMax:
		return Standard
	default:
		return Wide
	}
}

// Size is a terminal size, in cells.
type Size struct {
	Width  int
	Height int
}

// Region is a rectangle, in cells.
type Region struct {
	X      int
	Y      int
	Width  int
	Height int
}

// Empty reports a region with no area.
func (r Region) Empty() bool { return r.Width <= 0 || r.Height <= 0 }

// Wants is what the model asks to show.
//
// A value rather than a set of booleans threaded through the call: four
// parameters of which three are booleans is the arity the contract calls a
// smell, and this keeps the layout input readable at the call site.
type Wants struct {
	ShowTree       bool
	TreeWidth      int
	ShowTerminal   bool
	TerminalHeight int
	// FindHeight is how many rows the find bar needs, or zero when it is closed.
	// Asked for rather than assumed, because the bar grows when the replacement
	// field is revealed and the body has to give up the row.
	FindHeight int
}

// Regions is where every surface landed for one frame.
type Regions struct {
	Class     WidthClass
	MenuBar   Region
	Tabs      Region
	Tree      Region
	Editor    Region
	Terminal  Region
	FindBar   Region
	StatusBar Region
	// TreeOverlaid says the tree floats over the editor instead of sitting
	// beside it, which is what the compact class does.
	TreeOverlaid bool
}

// minEditorWidth is the narrowest the editor may become before a side panel is
// dropped. A tree that pushes the editor below this has stopped being useful.
const minEditorWidth = 40

// Compute places every surface for a terminal of the given size.
//
// The stack is fixed — menu bar, tabs, body, status bar — and the body splits
// horizontally. A zero or negative size is not an error: a terminal that has not
// been measured yet must produce empty regions rather than a panic or a negative
// width that renders as garbage.
func Compute(size Size, wants Wants) Regions {
	regions := Regions{Class: Classify(size.Width)}
	if size.Width <= 0 || size.Height <= 0 {
		return regions
	}

	y := 0
	regions.MenuBar, y = row(y, size.Width, size.Height)
	regions.Tabs, y = row(y, size.Width, size.Height)

	// The status bar is reserved before the body, so the body never has to know
	// whether it is the last surface.
	regions.StatusBar = Region{X: 0, Y: size.Height - 1, Width: size.Width, Height: 1}

	bodyHeight := size.Height - y - regions.StatusBar.Height
	if bodyHeight < 0 {
		bodyHeight = 0
	}
	body := Region{X: 0, Y: y, Width: size.Width, Height: bodyHeight}

	// The find bar sits above the terminal and below the editor, so opening it
	// takes rows from the editor rather than covering the text being searched.
	if wants.FindHeight > 0 && body.Height > wants.FindHeight {
		regions.FindBar = Region{
			X:      0,
			Y:      body.Y + body.Height - wants.FindHeight,
			Width:  body.Width,
			Height: wants.FindHeight,
		}
		body.Height -= wants.FindHeight
	}

	if wants.ShowTerminal && wants.TerminalHeight > 0 && body.Height > wants.TerminalHeight {
		regions.Terminal = Region{
			X:      0,
			Y:      body.Y + body.Height - wants.TerminalHeight,
			Width:  body.Width,
			Height: wants.TerminalHeight,
		}
		body.Height -= wants.TerminalHeight
	}

	regions.Tree, regions.Editor, regions.TreeOverlaid = splitBody(body, wants)
	return regions
}

// row reserves one line at the given offset, clamped to the terminal.
func row(y, width, height int) (Region, int) {
	if y >= height {
		return Region{}, y
	}
	return Region{X: 0, Y: y, Width: width, Height: 1}, y + 1
}

// splitBody divides the body between the tree and the editor.
//
// The tree gives way before the editor does: below the compact boundary it
// overlays instead of taking a column, and in any class it is dropped if the
// editor would fall under the minimum. An editor squeezed to a dozen columns is
// not a smaller editor, it is a broken one.
func splitBody(body Region, wants Wants) (tree, editor Region, overlaid bool) {
	editor = body

	if !wants.ShowTree || wants.TreeWidth <= 0 {
		return Region{}, editor, false
	}

	if Classify(body.Width) == Compact || body.Width-wants.TreeWidth < minEditorWidth {
		// Overlaid: the tree floats over the editor's top-left corner, so the
		// editor keeps its full width underneath.
		return Region{
			X:      body.X,
			Y:      body.Y,
			Width:  min(wants.TreeWidth, body.Width),
			Height: min(body.Height, overlaidTreeHeight),
		}, editor, true
	}

	tree = Region{X: body.X, Y: body.Y, Width: wants.TreeWidth, Height: body.Height}
	editor = Region{
		X:      body.X + wants.TreeWidth,
		Y:      body.Y,
		Width:  body.Width - wants.TreeWidth,
		Height: body.Height,
	}
	return tree, editor, false
}

// overlaidTreeHeight bounds the floating tree so it cannot cover the whole
// editor on a short terminal.
const overlaidTreeHeight = 12

// Condition measures text the way the layout assumes.
//
// Ambiguous-width characters count as wide. Erring that way costs a column;
// erring the other way misaligns the rest of the row, which is visible and
// wrong.
var condition = &runewidth.Condition{EastAsianWidth: true}

// Width measures a string in terminal cells.
//
// Escape sequences occupy no cells, so they are stripped before measuring. Without
// this, styling a row would change its measured width and every line to its right
// would shift — the bug would appear only once colour arrived, and only visually.
//
// The escape scan is a fast path rather than an optimisation for its own sake:
// every row of every frame is measured, and most rows carry no escape at all.
func Width(text string) int {
	if strings.IndexByte(text, escape) < 0 {
		return condition.StringWidth(text)
	}
	return condition.StringWidth(ansi.Strip(text))
}

// escape is the control character that begins an ANSI sequence.
const escape = 0x1b

// Truncate cuts text to fit width cells, marking that it was cut.
//
// The marker is not decoration: a label that vanishes without a sign is
// indistinguishable from a label that never existed, and the reader has no way
// to tell a short value from a clipped one.
func Truncate(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if Width(text) <= width {
		return text
	}

	marker := ellipsis
	if ellipsisWidth > width {
		// The marker counts as two cells, so it does not fit in a one-column
		// field. Returning it anyway would overflow the row by one column.
		marker = fallbackMarker
	}
	limit := width - condition.StringWidth(marker)
	var out strings.Builder
	used := 0
	for _, r := range text {
		cellWidth := condition.RuneWidth(r)
		if used+cellWidth > limit {
			break
		}
		out.WriteRune(r)
		used += cellWidth
	}
	return out.String() + marker
}

const ellipsis = "…"

// ellipsisWidth is measured, never assumed.
//
// The marker is an ambiguous-width character, and this package's own rule counts
// those as wide — so a hard-coded 1 disagreed with Width() and overflowed every
// truncated row by one column. The constant is now derived from the same
// measurement the rest of the package uses.
var ellipsisWidth = condition.StringWidth(ellipsis)

// fallbackMarker fits a single column, for the case where the ellipsis does not.
const fallbackMarker = "."

// Pad fits text into exactly width cells, truncating when needed.
//
// Exactly, not at most: a row that comes back short lets whatever is to its
// right shift left, which is how a two-column layout turns into a diagonal one.
func Pad(text string, width int) string {
	if width <= 0 {
		return ""
	}

	fitted := Truncate(text, width)
	if padding := width - Width(fitted); padding > 0 {
		return fitted + strings.Repeat(" ", padding)
	}
	return fitted
}
