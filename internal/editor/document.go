package editor

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ori-team/oride/internal/buffer"
)

// ID identifies an open document.
type ID uint64

// Document is one open buffer with its editing state.
//
// The buffer holds bytes; the selection holds byte offsets; the caret is a
// derived view for display. That division is what makes editing correct even
// where the caret representation is lossy — a caret cannot name a position
// inside a line break, but a byte offset can.
type Document struct {
	id        ID
	path      string
	hasPath   bool
	buffer    *buffer.Buffer
	selection Selection
	// extraCarets are the secondary cursors, in insertion order.
	extraCarets []buffer.Offset
	// preferredColumn carries a column across vertical movement, so moving down
	// through short lines returns to the column the user started in.
	preferredColumn *int
	undo            *UndoStack
	dirty           bool
	version         uint64
}

// NewDocument returns an empty document with no path.
func NewDocument(id ID) *Document {
	return &Document{
		id:        id,
		buffer:    buffer.New(),
		selection: CaretAt(0),
		undo:      NewUndoStack(),
	}
}

// NewDocumentFromText returns a document holding text.
func NewDocumentFromText(id ID, path, text string) *Document {
	return &Document{
		id:        id,
		path:      path,
		hasPath:   path != "",
		buffer:    buffer.FromText(text),
		selection: CaretAt(0),
		undo:      NewUndoStack(),
	}
}

// ID returns the document identifier.
func (d *Document) ID() ID { return d.id }

// Version counts modifications.
func (d *Document) Version() uint64 { return d.version }

// Path returns the file path, and whether there is one.
func (d *Document) Path() (string, bool) { return d.path, d.hasPath }

// SetPath sets the file path and marks the document saved.
func (d *Document) SetPath(path string) {
	d.path = path
	d.hasPath = path != ""
}

// IsDirty reports whether the document has unsaved changes.
func (d *Document) IsDirty() bool { return d.dirty }

// Buffer returns the underlying text buffer.
func (d *Document) Buffer() *buffer.Buffer { return d.buffer }

// Selection returns the current selection.
func (d *Document) Selection() Selection { return d.selection }

// Caret returns the caret position derived from the selection head.
func (d *Document) Caret() (buffer.Caret, error) {
	return d.buffer.ByteToCaret(d.selection.Head)
}

// ExtraCarets returns the secondary cursor offsets.
func (d *Document) ExtraCarets() []buffer.Offset {
	out := make([]buffer.Offset, len(d.extraCarets))
	copy(out, d.extraCarets)
	return out
}

// AllCaretOffsets returns every caret, sorted and deduplicated.
func (d *Document) AllCaretOffsets() []buffer.Offset {
	out := make([]buffer.Offset, 0, len(d.extraCarets)+1)
	out = append(out, d.selection.Head)
	out = append(out, d.extraCarets...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return dedupe(out)
}

// ClearExtraCarets removes every secondary cursor.
func (d *Document) ClearExtraCarets() { d.extraCarets = nil }

// CollapseSelection turns a range into a caret at its head.
func (d *Document) CollapseSelection() {
	d.selection = CaretAt(d.selection.Head)
}

// AddCursorAt adds a secondary cursor.
func (d *Document) AddCursorAt(offset buffer.Offset) { d.pushExtraCaret(offset) }

// TabTitle returns the label for the tab.
func (d *Document) TabTitle() string {
	if !d.hasPath {
		return "untitled"
	}
	if name := filepath.Base(d.path); name != "" && name != "." {
		return name
	}
	return d.path
}

// SelectedText returns the text the selection covers.
func (d *Document) SelectedText() string {
	if d.selection.IsEmpty() {
		return ""
	}
	text, err := d.buffer.TextRange(d.selection.Start(), d.selection.End())
	if err != nil {
		return ""
	}
	return text
}

// SetSelection replaces the selection and closes the open undo group.
func (d *Document) SetSelection(selection Selection) {
	d.selection = selection
	d.extraCarets = nil
	d.preferredColumn = nil
	d.undo.CommitGroup(d.selection)
}

// SetSelectionLive updates the selection during a drag.
//
// No group commit: a drag produces dozens of updates per second, and treating
// each as a boundary would make every mouse movement an undo step.
func (d *Document) SetSelectionLive(anchor, head buffer.Offset) {
	d.selection = NewSelection(anchor, head)
}

// JumpToByte moves the caret and closes the open undo group.
func (d *Document) JumpToByte(offset buffer.Offset) {
	d.selection = CaretAt(offset)
	d.extraCarets = nil
	d.preferredColumn = nil
	d.undo.CommitGroup(d.selection)
}

// SelectByteRange selects a range and closes the open undo group.
func (d *Document) SelectByteRange(start, end buffer.Offset) {
	d.selection = NewSelection(start, end)
	d.extraCarets = nil
	d.preferredColumn = nil
	d.undo.CommitGroup(d.selection)
}

// InsertText inserts at the caret, replacing the selection.
func (d *Document) InsertText(text string) error {
	if len(d.extraCarets) > 0 && d.selection.IsEmpty() {
		return d.insertTextMulti(text)
	}
	if !d.selection.IsEmpty() {
		if err := d.DeleteSelection(); err != nil {
			return err
		}
	}

	at := d.selection.Head
	if err := d.buffer.Insert(at, text); err != nil {
		return err
	}
	d.undo.PushApplied(Edit{At: at, Text: text}, d.selection)
	d.selection = CaretAt(at + buffer.Offset(len(text)))
	d.markModified()
	d.commitNewlineBoundary(text)
	return nil
}

// insertTextMulti inserts at every caret, from the highest offset down.
//
// Downwards because a prior edit would otherwise invalidate the offsets of the
// edits still to come.
func (d *Document) insertTextMulti(text string) error {
	carets := d.AllCaretOffsets()
	step := len(text)

	newOffsets := make([]buffer.Offset, 0, len(carets))
	for i, at := range carets {
		newOffsets = append(newOffsets, at+buffer.Offset((i+1)*step))
	}

	for i := len(carets) - 1; i >= 0; i-- {
		at := carets[i]
		if err := d.buffer.Insert(at, text); err != nil {
			return err
		}
		d.undo.PushApplied(Edit{At: at, Text: text}, d.selection)
	}
	d.setCarets(newOffsets)
	d.preferredColumn = nil
	d.markModified()
	d.commitNewlineBoundary(text)
	return nil
}

// commitNewlineBoundary closes the open undo group after a newline.
//
// Without it a whole typing session across several lines collapses into one
// group and a single undo erases all of it. The reference implementation had
// that defect; this one does not, and the divergence is recorded as ledger B1.
func (d *Document) commitNewlineBoundary(inserted string) {
	if strings.Contains(inserted, "\n") {
		d.undo.CommitGroup(d.selection)
	}
}

// Backspace deletes the selection, or the character before the caret.
func (d *Document) Backspace() error {
	if !d.selection.IsEmpty() {
		return d.DeleteSelection()
	}
	if len(d.extraCarets) > 0 {
		return d.backspaceMulti()
	}

	end := d.selection.Head
	if end == 0 {
		return nil
	}
	start, err := d.buffer.PrevCharOffset(end)
	if err != nil {
		return err
	}
	removed, err := d.buffer.DeleteRange(start, end)
	if err != nil {
		return err
	}
	d.undo.PushApplied(Edit{At: start, Removed: true, Text: removed}, d.selection)
	d.selection = CaretAt(start)
	d.preferredColumn = nil
	d.markModified()
	return nil
}

func (d *Document) backspaceMulti() error {
	carets := d.AllCaretOffsets()

	type operation struct {
		start, end buffer.Offset
		length     int
	}
	ops := make([]operation, 0, len(carets))
	for _, end := range carets {
		if end == 0 {
			ops = append(ops, operation{end, end, 0})
			continue
		}
		start, err := d.buffer.PrevCharOffset(end)
		if err != nil {
			return err
		}
		ops = append(ops, operation{start, end, int(end - start)})
	}

	newOffsets := make([]buffer.Offset, 0, len(ops))
	cumulative := 0
	for _, op := range ops {
		newOffsets = append(newOffsets, op.start-buffer.Offset(cumulative))
		cumulative += op.length
	}

	for i := len(ops) - 1; i >= 0; i-- {
		op := ops[i]
		if op.length == 0 {
			continue
		}
		removed, err := d.buffer.DeleteRange(op.start, op.end)
		if err != nil {
			return err
		}
		d.undo.PushApplied(Edit{At: op.start, Removed: true, Text: removed}, d.selection)
	}

	d.setCarets(newOffsets)
	d.preferredColumn = nil
	d.markModified()
	return nil
}

// DeleteForward deletes the selection, or the character after the caret.
func (d *Document) DeleteForward() error {
	if !d.selection.IsEmpty() {
		return d.DeleteSelection()
	}
	if len(d.extraCarets) > 0 {
		return d.deleteForwardMulti()
	}

	start := d.selection.Head
	if int(start) >= d.buffer.LenBytes() {
		return nil
	}
	end, err := d.buffer.NextCharOffset(start)
	if err != nil {
		return err
	}
	removed, err := d.buffer.DeleteRange(start, end)
	if err != nil {
		return err
	}
	d.undo.PushApplied(Edit{At: start, Removed: true, Text: removed}, d.selection)
	d.preferredColumn = nil
	d.markModified()
	return nil
}

func (d *Document) deleteForwardMulti() error {
	carets := d.AllCaretOffsets()
	bufferLen := buffer.Offset(d.buffer.LenBytes())

	type operation struct {
		start, end buffer.Offset
		length     int
	}
	ops := make([]operation, 0, len(carets))
	for _, start := range carets {
		if start >= bufferLen {
			ops = append(ops, operation{start, start, 0})
			continue
		}
		end, err := d.buffer.NextCharOffset(start)
		if err != nil {
			return err
		}
		ops = append(ops, operation{start, end, int(end - start)})
	}

	newOffsets := make([]buffer.Offset, 0, len(ops))
	cumulative := 0
	for _, op := range ops {
		newOffsets = append(newOffsets, op.start-buffer.Offset(cumulative))
		cumulative += op.length
	}

	for i := len(ops) - 1; i >= 0; i-- {
		op := ops[i]
		if op.length == 0 {
			continue
		}
		removed, err := d.buffer.DeleteRange(op.start, op.end)
		if err != nil {
			return err
		}
		d.undo.PushApplied(Edit{At: op.start, Removed: true, Text: removed}, d.selection)
	}

	d.setCarets(newOffsets)
	d.preferredColumn = nil
	d.markModified()
	return nil
}

// DeleteSelection removes the selected range.
func (d *Document) DeleteSelection() error {
	if d.selection.IsEmpty() {
		return nil
	}
	start, end := d.selection.Start(), d.selection.End()
	removed, err := d.buffer.DeleteRange(start, end)
	if err != nil {
		return err
	}
	d.undo.PushApplied(Edit{At: start, Removed: true, Text: removed}, d.selection)
	d.selection = CaretAt(start)
	d.preferredColumn = nil
	d.markModified()
	return nil
}

// Undo reverses the last group and restores the selection it started from.
func (d *Document) Undo() (bool, error) {
	d.undo.CommitGroup(d.selection)

	changed, restore, err := d.undo.Undo(d.buffer)
	if err != nil {
		return false, err
	}
	if !changed {
		return false, nil
	}
	d.restoreSelection(restore)
	d.markModified()
	return true, nil
}

// Redo reapplies the last undone group and restores the selection it produced.
func (d *Document) Redo() (bool, error) {
	d.undo.CommitGroup(d.selection)

	changed, restore, err := d.undo.Redo(d.buffer)
	if err != nil {
		return false, err
	}
	if !changed {
		return false, nil
	}
	d.restoreSelection(restore)
	d.markModified()
	return true, nil
}

// restoreSelection puts the caret where the history step says it belongs.
//
// A group records the selection it started and ended with, so undo returns the
// caret to before the step and redo to after it. Falling back to clamping the
// current head is what left the caret at the wrong end of restored text in the
// reference implementation — ledger item B2.
func (d *Document) restoreSelection(restore *Selection) {
	length := buffer.Offset(d.buffer.LenBytes())

	if restore == nil {
		// A group from before selections were recorded, or one built by hand.
		d.selection = CaretAt(clampOffset(d.selection.Head, length))
	} else {
		d.selection = Selection{
			Anchor: clampOffset(restore.Anchor, length),
			Head:   clampOffset(restore.Head, length),
		}
	}
	d.clampExtraCarets(length)
	d.preferredColumn = nil
}

// clampExtraCarets drops secondary cursors the buffer can no longer hold.
//
// Without it a cursor survives an undo pointing past the end of the text, and the
// next insert fails — ledger item B4.
func (d *Document) clampExtraCarets(length buffer.Offset) {
	kept := d.extraCarets[:0]
	for _, caret := range d.extraCarets {
		if caret <= length {
			kept = append(kept, caret)
		}
	}
	d.extraCarets = kept
}

func clampOffset(offset, length buffer.Offset) buffer.Offset {
	if offset > length {
		return length
	}
	return offset
}

// CommitEditGroup closes the open undo group.
func (d *Document) CommitEditGroup() { d.undo.CommitGroup(d.selection) }

// MarkSaved clears the dirty flag and closes the open group.
func (d *Document) MarkSaved() {
	d.undo.CommitGroup(d.selection)
	d.dirty = false
}

// UndoHistoryLabels returns the undo history labels.
func (d *Document) UndoHistoryLabels() []string { return d.undo.UndoLabels() }

// RedoHistoryLabels returns the redo history labels.
func (d *Document) RedoHistoryLabels() []string { return d.undo.RedoLabels() }

func (d *Document) markModified() {
	d.dirty = true
	d.version++
}

// pushExtraCaret adds a secondary cursor, ignoring a duplicate or a position
// that is already the primary caret.
func (d *Document) pushExtraCaret(offset buffer.Offset) {
	if offset == d.selection.Head {
		return
	}
	for _, existing := range d.extraCarets {
		if existing == offset {
			return
		}
	}
	d.extraCarets = append(d.extraCarets, offset)
}

// setCarets replaces every caret from a set of offsets, sorted and deduplicated.
//
// It does not close the undo group. The edit paths call it while their group is
// still open, and closing there would split an edit from itself: three cursors
// typing one character would become three undo steps.
func (d *Document) setCarets(offsets []buffer.Offset) {
	offsets = dedupeSorted(offsets)
	if len(offsets) == 0 {
		return
	}
	d.selection = CaretAt(offsets[0])
	d.extraCarets = append([]buffer.Offset(nil), offsets[1:]...)
}

// setCaretsAtBoundary replaces the carets and closes the open undo group.
//
// Movement is a boundary; an edit is not. This is the difference between "I
// typed here" and "I went there".
func (d *Document) setCaretsAtBoundary(offsets []buffer.Offset) {
	d.setCarets(offsets)
	d.undo.CommitGroup(d.selection)
}

// dedupeSorted sorts and removes repeated offsets.
func dedupeSorted(offsets []buffer.Offset) []buffer.Offset {
	sort.Slice(offsets, func(i, j int) bool { return offsets[i] < offsets[j] })
	return dedupe(offsets)
}

// moveHeadTo moves the caret and closes the open undo group.
//
// Movement is an undo boundary: typing, moving and typing again are two
// thoughts, and one undo must not erase both.
func (d *Document) moveHeadTo(head buffer.Offset, extend bool) {
	d.selection = d.selection.MoveHead(head, extend)
	if !extend {
		d.extraCarets = nil
	}
	d.undo.CommitGroup(d.selection)
}

func dedupe(sorted []buffer.Offset) []buffer.Offset {
	if len(sorted) < 2 {
		return sorted
	}
	out := sorted[:1]
	for _, offset := range sorted[1:] {
		if offset != out[len(out)-1] {
			out = append(out, offset)
		}
	}
	return out
}

// DocumentError wraps a buffer failure with the document it happened in.
type DocumentError struct {
	ID    ID
	Inner error
}

func (e DocumentError) Error() string {
	return fmt.Sprintf("documento %d: %v", e.ID, e.Inner)
}

func (e DocumentError) Unwrap() error { return e.Inner }

// ErrDocumentNotFound is returned when a document id is not open.
var ErrDocumentNotFound = errors.New("document not open")
