// Package menubar renders the top line of menu labels.
package menubar

import (
	"strings"

	"github.com/ori-team/oride/internal/tui/layout"
)

// View is what the bar needs to draw.
type View struct {
	// Labels come from the locale catalog, so the bar has no hard-coded text.
	Labels []string
	// Open is the index of the expanded menu, or -1.
	Open int
}

// Markers around the open label. The open menu is marked, not only tinted.
const (
	openPrefix = "["
	openSuffix = "]"
)

// Render draws the menu bar at an exact width.
func Render(width int, view View) string {
	if width <= 0 {
		return ""
	}
	if len(view.Labels) == 0 {
		return layout.Pad("", width)
	}

	parts := make([]string, 0, len(view.Labels))
	for index, label := range view.Labels {
		if index == view.Open {
			parts = append(parts, openPrefix+label+openSuffix)
			continue
		}
		parts = append(parts, label)
	}

	bar := " " + strings.Join(parts, "  ")
	return layout.Pad(bar, width)
}

// LabelsFor returns the menu labels for a locale catalog, in bar order.
//
// The order lives here rather than in the catalog: a translation supplies the
// words, and the bar decides the arrangement — otherwise a locale file could
// reorder the menus.
func LabelsFor(file, edit, view, goMenu, git, help string) []string {
	return []string{file, edit, view, goMenu, git, help}
}
