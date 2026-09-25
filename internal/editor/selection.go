package editor

import "github.com/ori-team/oride/internal/buffer"

// Selection is an oriented range: Anchor is where it started, Head is the caret.
//
// When the two are equal the selection is a caret with no range. Keeping the
// direction matters — extending left and then right is not the same selection
// as extending right and then left, and the user can tell.
type Selection struct {
	Anchor buffer.Offset
	Head   buffer.Offset
}

// CaretAt returns a collapsed selection at an offset.
func CaretAt(at buffer.Offset) Selection {
	return Selection{Anchor: at, Head: at}
}

// NewSelection returns an oriented range.
func NewSelection(anchor, head buffer.Offset) Selection {
	return Selection{Anchor: anchor, Head: head}
}

// IsEmpty reports whether the selection has no range.
func (s Selection) IsEmpty() bool { return s.Anchor == s.Head }

// Start returns the lower end of the range.
func (s Selection) Start() buffer.Offset {
	if s.Anchor <= s.Head {
		return s.Anchor
	}
	return s.Head
}

// End returns the higher end of the range.
func (s Selection) End() buffer.Offset {
	if s.Anchor <= s.Head {
		return s.Head
	}
	return s.Anchor
}

// MoveHead moves the caret, keeping or discarding the anchor.
//
// Not extending collapses the selection, which is what an unshifted arrow does.
func (s Selection) MoveHead(head buffer.Offset, extend bool) Selection {
	if extend {
		return Selection{Anchor: s.Anchor, Head: head}
	}
	return CaretAt(head)
}
