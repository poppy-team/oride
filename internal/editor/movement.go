package editor

import "github.com/ori-team/oride/internal/buffer"

// MoveLeft moves every caret one character left.
func (d *Document) MoveLeft(extend bool) error {
	if len(d.extraCarets) > 0 && !extend {
		d.moveEachCaret(func(offset buffer.Offset) buffer.Offset {
			previous, err := d.buffer.PrevCharOffset(offset)
			if err != nil {
				return offset
			}
			return previous
		})
		d.preferredColumn = nil
		return nil
	}
	head, err := d.buffer.PrevCharOffset(d.selection.Head)
	if err != nil {
		return err
	}
	d.preferredColumn = nil
	d.moveHeadTo(head, extend)
	return nil
}

// MoveRight moves every caret one character right.
func (d *Document) MoveRight(extend bool) error {
	if len(d.extraCarets) > 0 && !extend {
		d.moveEachCaret(func(offset buffer.Offset) buffer.Offset {
			next, err := d.buffer.NextCharOffset(offset)
			if err != nil {
				return offset
			}
			return next
		})
		d.preferredColumn = nil
		return nil
	}
	head, err := d.buffer.NextCharOffset(d.selection.Head)
	if err != nil {
		return err
	}
	d.preferredColumn = nil
	d.moveHeadTo(head, extend)
	return nil
}

// MoveUp moves every caret one line up, preserving the preferred column.
func (d *Document) MoveUp(extend bool) error {
	if len(d.extraCarets) > 0 && !extend {
		d.moveEachCaret(d.shiftLine(-1))
		return nil
	}
	caret, err := d.buffer.ByteToCaret(d.selection.Head)
	if err != nil {
		return err
	}
	if caret.Line == 0 {
		return nil
	}
	column := d.columnFor(caret)
	target, err := d.buffer.CaretToByte(buffer.Caret{Line: caret.Line - 1, Column: column})
	if err != nil {
		return err
	}
	d.moveHeadTo(target, extend)
	return nil
}

// MoveDown moves every caret one line down, preserving the preferred column.
func (d *Document) MoveDown(extend bool) error {
	if len(d.extraCarets) > 0 && !extend {
		d.moveEachCaret(d.shiftLine(1))
		return nil
	}
	caret, err := d.buffer.ByteToCaret(d.selection.Head)
	if err != nil {
		return err
	}
	if caret.Line >= d.buffer.LineCount()-1 {
		return nil
	}
	column := d.columnFor(caret)
	target, err := d.buffer.CaretToByte(buffer.Caret{Line: caret.Line + 1, Column: column})
	if err != nil {
		return err
	}
	d.moveHeadTo(target, extend)
	return nil
}

// MoveLineStart moves every caret to the start of its line.
func (d *Document) MoveLineStart(extend bool) error {
	if len(d.extraCarets) > 0 && !extend {
		d.moveEachCaret(func(offset buffer.Offset) buffer.Offset {
			caret, err := d.buffer.ByteToCaret(offset)
			if err != nil {
				return offset
			}
			start, err := d.buffer.CaretToByte(buffer.Caret{Line: caret.Line, Column: 0})
			if err != nil {
				return offset
			}
			return start
		})
		d.preferredColumn = intPointer(0)
		return nil
	}
	caret, err := d.buffer.ByteToCaret(d.selection.Head)
	if err != nil {
		return err
	}
	head, err := d.buffer.CaretToByte(buffer.Caret{Line: caret.Line, Column: 0})
	if err != nil {
		return err
	}
	d.preferredColumn = intPointer(0)
	d.moveHeadTo(head, extend)
	return nil
}

// MoveLineEnd moves every caret to the end of its line's content.
func (d *Document) MoveLineEnd(extend bool) error {
	if len(d.extraCarets) > 0 && !extend {
		d.moveEachCaret(func(offset buffer.Offset) buffer.Offset {
			caret, err := d.buffer.ByteToCaret(offset)
			if err != nil {
				return offset
			}
			line, err := d.buffer.LineText(caret.Line)
			if err != nil {
				return offset
			}
			end, err := d.buffer.CaretToByte(buffer.Caret{
				Line:   caret.Line,
				Column: runeCount(line),
			})
			if err != nil {
				return offset
			}
			return end
		})
		return nil
	}
	caret, err := d.buffer.ByteToCaret(d.selection.Head)
	if err != nil {
		return err
	}
	line, err := d.buffer.LineText(caret.Line)
	if err != nil {
		return err
	}
	column := runeCount(line)
	head, err := d.buffer.CaretToByte(buffer.Caret{Line: caret.Line, Column: column})
	if err != nil {
		return err
	}
	d.preferredColumn = intPointer(column)
	d.moveHeadTo(head, extend)
	return nil
}

// MoveBufferStart moves every caret to offset zero.
func (d *Document) MoveBufferStart(extend bool) error {
	d.preferredColumn = intPointer(0)
	d.moveHeadTo(0, extend)
	return nil
}

// MoveBufferEnd moves every caret to the end of the buffer.
func (d *Document) MoveBufferEnd(extend bool) error {
	d.preferredColumn = nil
	d.moveHeadTo(buffer.Offset(d.buffer.LenBytes()), extend)
	return nil
}

// SelectAll selects the whole document.
func (d *Document) SelectAll() {
	d.selection = NewSelection(0, buffer.Offset(d.buffer.LenBytes()))
	d.extraCarets = nil
	d.preferredColumn = nil
	d.undo.CommitGroup(d.selection)
}

// AddCursorAbove adds a secondary cursor one line above the topmost caret.
func (d *Document) AddCursorAbove() error {
	carets := d.AllCaretOffsets()
	topmost := d.selection.Head
	if len(carets) > 0 {
		topmost = carets[0]
	}
	caret, err := d.buffer.ByteToCaret(topmost)
	if err != nil {
		return err
	}
	if caret.Line == 0 {
		return nil
	}
	column := d.columnFor(caret)
	target, err := d.buffer.CaretToByte(buffer.Caret{Line: caret.Line - 1, Column: column})
	if err != nil {
		return err
	}
	d.pushExtraCaret(target)
	return nil
}

// AddCursorBelow adds a secondary cursor one line below the bottommost caret.
func (d *Document) AddCursorBelow() error {
	carets := d.AllCaretOffsets()
	bottommost := d.selection.Head
	if len(carets) > 0 {
		bottommost = carets[len(carets)-1]
	}
	caret, err := d.buffer.ByteToCaret(bottommost)
	if err != nil {
		return err
	}
	if caret.Line >= d.buffer.LineCount()-1 {
		return nil
	}
	column := d.columnFor(caret)
	target, err := d.buffer.CaretToByte(buffer.Caret{Line: caret.Line + 1, Column: column})
	if err != nil {
		return err
	}
	d.pushExtraCaret(target)
	return nil
}

// moveEachCaret applies a movement to every caret at once.
//
// A caret that cannot move keeps its position rather than being dropped: losing
// a cursor because it reached the start of the buffer would be a surprise.
func (d *Document) moveEachCaret(move func(buffer.Offset) buffer.Offset) {
	carets := d.AllCaretOffsets()
	next := make([]buffer.Offset, 0, len(carets))
	for _, offset := range carets {
		next = append(next, move(offset))
	}
	d.setCaretsAtBoundary(next)
}

// shiftLine returns a movement that moves a caret one line in a direction,
// remembering the preferred column.
func (d *Document) shiftLine(delta int) func(buffer.Offset) buffer.Offset {
	return func(offset buffer.Offset) buffer.Offset {
		caret, err := d.buffer.ByteToCaret(offset)
		if err != nil {
			return offset
		}
		target := caret.Line + delta
		if target < 0 || target >= d.buffer.LineCount() {
			return offset
		}
		// Read-only: the multi-caret path must not record a preferred column,
		// because each cursor would overwrite it for the next one.
		column := d.preferredOr(caret)
		moved, err := d.buffer.CaretToByte(buffer.Caret{Line: target, Column: column})
		if err != nil {
			return offset
		}
		return moved
	}
}

// preferredOr returns the column a vertical move should aim for, without
// recording it.
func (d *Document) preferredOr(caret buffer.Caret) int {
	if d.preferredColumn != nil {
		return *d.preferredColumn
	}
	return caret.Column
}

// columnFor returns the column a vertical move should aim for, and records it.
//
// Recording is what carries the column across a pass through a short line:
// moving down onto a line with three characters and then further down returns to
// the original column, not to column three.
func (d *Document) columnFor(caret buffer.Caret) int {
	column := d.preferredOr(caret)
	d.preferredColumn = intPointer(column)
	return column
}

func intPointer(value int) *int { return &value }

func runeCount(text string) int {
	count := 0
	for range text {
		count++
	}
	return count
}
