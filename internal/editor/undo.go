package editor

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/ori-team/oride/internal/buffer"
)

// Edit is an atomic change, stored in a form that can be applied in either
// direction. A transaction without its inverse is not undoable, so the two are
// recorded together.
type Edit struct {
	// At is where the text was inserted or removed.
	At buffer.Offset
	// Removed is true for a deletion, false for an insertion.
	Removed bool
	// Text is the inserted or removed text.
	Text string
}

// Apply performs the edit on a buffer.
func (e Edit) Apply(b *buffer.Buffer) error {
	if e.Removed {
		_, err := b.DeleteRange(e.At, e.At+buffer.Offset(len(e.Text)))
		return err
	}
	return b.Insert(e.At, e.Text)
}

// Inverse returns the edit that undoes this one.
func (e Edit) Inverse() Edit {
	return Edit{At: e.At, Removed: !e.Removed, Text: e.Text}
}

// EditGroup is a set of edits that undo as one step, plus the selection the step
// started and ended with.
//
// Carrying the selection is what makes redo land where the edit did. Without it
// the caret has to be salvaged from wherever the undo left it, which puts the
// cursor at the wrong end of the restored text — ledger item B2.
type EditGroup struct {
	edits []Edit
	// before is the selection when the group's first edit was recorded.
	before *Selection
	// after is the selection when the group was committed.
	after *Selection
}

// NewEditGroup returns an empty group.
func NewEditGroup() *EditGroup { return &EditGroup{} }

// Push appends an edit, recording where the group started.
func (g *EditGroup) Push(e Edit, before Selection) {
	if g.before == nil {
		start := before
		g.before = &start
	}
	g.edits = append(g.edits, e)
}

// Len returns the number of edits in the group.
func (g *EditGroup) Len() int { return len(g.edits) }

// IsEmpty reports whether the group holds no edits.
func (g *EditGroup) IsEmpty() bool { return len(g.edits) == 0 }

// Summary is the label the undo history shows for this group.
//
// The exact text matters: it is part of the observable state, so a case
// comparing undo labels against the oracle fails on a wording difference. The
// quotes are the typographic ones the reference emits, not ASCII.
func (g *EditGroup) Summary() string {
	if len(g.edits) == 0 {
		return "(vazio)"
	}
	first := g.edits[0]
	preview := truncateRunes(first.Text, 24)

	verb := "insert"
	if first.Removed {
		verb = "delete"
	}

	switch {
	case len(g.edits) > 1:
		return fmt.Sprintf("%s “%s…” (+%d edits)", verb, preview, len(g.edits)-1)
	case !first.Removed && strings.Contains(first.Text, "\n"):
		return fmt.Sprintf("insert %d lines", countLines(first.Text))
	default:
		return fmt.Sprintf("%s “%s”", verb, preview)
	}
}

func (g *EditGroup) applyForward(b *buffer.Buffer) error {
	for _, e := range g.edits {
		if err := e.Apply(b); err != nil {
			return err
		}
	}
	return nil
}

// applyInverse applies the edits backwards, each inverted.
//
// Backwards is not an optimisation: undoing an insert and a delete that touched
// the same region only lands correctly if the later edit is reversed first.
func (g *EditGroup) applyInverse(b *buffer.Buffer) error {
	for i := len(g.edits) - 1; i >= 0; i-- {
		if err := g.edits[i].Inverse().Apply(b); err != nil {
			return err
		}
	}
	return nil
}

// UndoStack is a linear undo history with a coalescing open group.
//
// Linear on purpose: the product's undo has no branches, and claiming otherwise
// in the documentation was one of the divergences this migration is fixing.
type UndoStack struct {
	undo []*EditGroup
	redo []*EditGroup
	// open collects edits that have not yet crossed a boundary.
	open *EditGroup
}

// NewUndoStack returns an empty history.
func NewUndoStack() *UndoStack { return &UndoStack{} }

// PushApplied records an edit already applied to the buffer, and drops the redo
// branch — editing after undoing abandons the future that was undone.
//
// `before` is the selection prior to the edit that produced `e`; only the first
// one for a group is kept.
func (s *UndoStack) PushApplied(e Edit, before Selection) {
	s.redo = nil
	if s.open == nil {
		s.open = NewEditGroup()
	}
	s.open.Push(e, before)
}

// CommitGroup closes the open group, recording the selection it produced.
//
// Called on every boundary the user can perceive: a newline, a cursor move, a
// save, a focus change. Without it the open group swallows an entire typing
// session and one Ctrl+Z erases all of it.
func (s *UndoStack) CommitGroup(after Selection) {
	if s.open == nil {
		return
	}
	if !s.open.IsEmpty() {
		end := after
		s.open.after = &end
		s.undo = append(s.undo, s.open)
	}
	s.open = nil
}

// Undo reverses the last group and reports the selection to restore.
//
// The returned selection is the one the group started with, which is exactly
// where the caret was before the step — not a clamped version of where it
// happens to be now.
func (s *UndoStack) Undo(b *buffer.Buffer) (bool, *Selection, error) {
	if len(s.undo) == 0 {
		return false, nil, nil
	}
	group := s.undo[len(s.undo)-1]
	s.undo = s.undo[:len(s.undo)-1]
	if err := group.applyInverse(b); err != nil {
		return false, nil, err
	}
	s.redo = append(s.redo, group)
	return true, group.before, nil
}

// Redo reapplies the last undone group and reports the selection to restore.
func (s *UndoStack) Redo(b *buffer.Buffer) (bool, *Selection, error) {
	if len(s.redo) == 0 {
		return false, nil, nil
	}
	group := s.redo[len(s.redo)-1]
	s.redo = s.redo[:len(s.redo)-1]
	if err := group.applyForward(b); err != nil {
		return false, nil, err
	}
	s.undo = append(s.undo, group)
	return true, group.after, nil
}

// CanUndo reports whether there is anything to undo, counting the open group.
func (s *UndoStack) CanUndo() bool {
	return (s.open != nil && !s.open.IsEmpty()) || len(s.undo) > 0
}

// CanRedo reports whether there is anything to redo.
func (s *UndoStack) CanRedo() bool { return len(s.redo) > 0 }

// UndoLabels returns the history labels, oldest first.
//
// The open group is labelled distinctly and last, because it is the step an
// undo would take next — showing it as if it were already committed would
// misnumber the history the user sees.
func (s *UndoStack) UndoLabels() []string {
	out := make([]string, 0, len(s.undo)+1)
	for _, group := range s.undo {
		out = append(out, group.Summary())
	}
	if s.open != nil && !s.open.IsEmpty() {
		out = append(out, "(aberto) "+s.open.Summary())
	}
	return out
}

// RedoLabels returns the redo labels.
func (s *UndoStack) RedoLabels() []string {
	out := make([]string, 0, len(s.redo))
	for _, group := range s.redo {
		out = append(out, group.Summary())
	}
	return out
}

// Len returns the number of committed groups plus the open one.
func (s *UndoStack) Len() int {
	extra := 0
	if s.open != nil && !s.open.IsEmpty() {
		extra = 1
	}
	return len(s.undo) + extra
}

func truncateRunes(text string, limit int) string {
	if utf8.RuneCountInString(text) <= limit {
		return text
	}
	count := 0
	for index := range text {
		if count == limit {
			return text[:index]
		}
		count++
	}
	return text
}

func countLines(text string) int {
	if text == "" {
		return 0
	}
	lines := strings.Count(text, "\n")
	// A trailing newline does not start a line for this label; the reference
	// counts the lines the text occupies.
	if !strings.HasSuffix(text, "\n") {
		lines++
	}
	return lines
}
