// Package tabs renders the tab bar.
package tabs

import (
	"strings"

	"github.com/ori-team/oride/internal/tui/layout"
)

// Tab is one open document as the bar sees it.
type Tab struct {
	Title  string
	Dirty  bool
	Active bool
}

// View is what the bar needs to draw.
type View struct {
	Tabs    []Tab
	Focused bool
}

// Markers. Text, not colour: a state carried only by a tint is invisible in a
// no-colour terminal, and the rule is that a tinted state is also spelled out.
const (
	dirtySuffix = "*"
	activeOpen  = "["
	activeClose = "]"
)

// Render draws the tab bar at an exact width.
//
// The active tab is drawn first when they do not all fit: a bar that hides which
// document is being edited is worse than one that hides the others.
func Render(width int, view View) string {
	if width <= 0 {
		return ""
	}
	if len(view.Tabs) == 0 {
		return layout.Pad(" (sem documentos) ", width)
	}

	ordered := view.Tabs
	if active := activeIndex(view.Tabs); active > 0 {
		ordered = append([]Tab{view.Tabs[active]}, append(append([]Tab(nil), view.Tabs[:active]...), view.Tabs[active+1:]...)...)
	}

	var out strings.Builder
	for _, tab := range ordered {
		chip := chipText(tab)
		if layout.Width(out.String())+layout.Width(chip) > width {
			// The remaining tabs do not fit. The cut is marked, because a bar that
			// silently drops tabs looks like a bar with fewer tabs.
			return layout.Pad(appendMarker(out.String(), width), width)
		}
		out.WriteString(chip)
	}
	return layout.Pad(out.String(), width)
}

// activeIndex finds the active tab, or -1.
func activeIndex(tabs []Tab) int {
	for index, tab := range tabs {
		if tab.Active {
			return index
		}
	}
	return -1
}

// chipText renders one tab as a fixed-spacing chip.
func chipText(tab Tab) string {
	title := tab.Title
	if tab.Dirty {
		title += dirtySuffix
	}
	if tab.Active {
		return " " + activeOpen + title + activeClose + " "
	}
	return " " + title + " "
}

// appendMarker adds the cut marker when it fits.
func appendMarker(text string, width int) string {
	marker := layout.Truncate("…", width)
	if layout.Width(text)+layout.Width(marker) > width {
		return text
	}
	return text + marker
}
