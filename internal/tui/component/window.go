// Package component holds the primitives more than one surface needs.
//
// It exists because the duplication was measured, not imagined: scrolling a
// selected row into view and painting a selected row both appear in the tree and
// the editor, and both appeared five and three times respectively in the Rust UI.
// Nothing is added here on the chance that a third surface will want it — the
// contract's §1.6 allows an abstraction at two proven uses and forbids it at one.
package component

// Window computes the visible slice of a list.
//
// It returns the first and last exclusive indices to draw. The rule is the one
// every list follows: the selection is always visible, and the list does not jump
// while the selection moves inside the already-visible range.
func Window(total, height, selected, scroll int) (start, end int) {
	if total <= 0 || height <= 0 {
		return 0, 0
	}

	start = clamp(scroll, 0, max(0, total-height))
	if selected < start {
		start = selected
	}
	if selected >= start+height {
		start = selected - height + 1
	}

	// Re-clamping after following the selection: following it can push the window
	// past the end, which would leave blank rows below the last item.
	start = clamp(start, 0, max(0, total-height))
	return start, min(start+height, total)
}

// clamp keeps a value inside a range.
func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

// Offset is where a row sits inside a window.
type Offset struct {
	// Visible reports whether the row is drawn at all.
	Visible bool
	// Row is the screen row, counted from the top of the window.
	Row int
	// Last is the final visible row, so a renderer can stop.
	Last bool
}
