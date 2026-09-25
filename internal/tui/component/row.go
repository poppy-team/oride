package component

import (
	"strings"

	"github.com/ori-team/oride/internal/tui/layout"
)

// Mark is the cursor drawn beside the selected row.
//
// A visible mark rather than colour alone: the rule is that no state is carried
// by colour, and a selection tinted but unmarked is invisible to anyone who
// cannot see the tint — and to anyone in a no-colour terminal.
const Mark = "▶"

// indentUnit is the spaces one level of depth adds.
const indentUnit = 2

// cursorWidth is the column reserved for the cursor, measured rather than
// assumed.
//
// The mark is an ambiguous-width character and this package counts those as wide,
// so writing `Mark + " "` reserved three cells while an unselected row reserved
// two — and every row shifted sideways the moment it was selected. Deriving the
// width from the same measurement that fills it removes the disagreement.
var cursorWidth = layout.Width(Mark) + 1

// Row renders one list entry at an exact width.
//
// The cursor column is reserved whether or not the row is selected, so selecting
// a row does not shift the text sideways — a list that jumps on every arrow key
// is exhausting to read.
func Row(name string, depth int, selected bool, width int) string {
	cursor := layout.Pad("", cursorWidth)
	if selected {
		cursor = layout.Pad(Mark, cursorWidth)
	}

	prefix := cursor + strings.Repeat(" ", depth*indentUnit)
	return layout.Pad(prefix+name, width)
}
