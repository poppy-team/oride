// Package buffer holds editable text with efficient line and byte access.
//
// The Rust implementation delegated this to `ropey`. Go has no maintained rope,
// so the index arithmetic is owned here — and that arithmetic is the most
// expensive class of bug in an editor. A position has four meanings that look
// alike and are not:
//
//   - a byte offset, which is what this package's API speaks;
//   - a rune (scalar) offset, which is what a caret column counts;
//   - a grapheme cluster, which is what a user perceives as one character;
//   - a visual column, which accounts for tab width and wide characters.
//
// Only the first two are modelled here, because they are the two the product's
// observable state uses. Grapheme and visual columns belong to the view layer
// and must not be computed anywhere else but there.
package buffer

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

// Offset is a byte offset from the start of the buffer.
//
// A named type because confusing a byte offset with a column is the bug this
// package exists to prevent; the compiler can help when the two are not both
// plain integers.
type Offset int

// Caret is a 0-based position: a line, and a column counted in Unicode scalars.
//
// Not bytes, and not visual columns. A caret at the newline of a line reports
// the column just past its content.
type Caret struct {
	Line   int
	Column int
}

// Errors returned by Buffer. They are values rather than formatted strings so a
// caller can react to the class of failure, and so a test can assert on it.
var (
	ErrOffsetOutOfBounds = errors.New("byte offset out of bounds")
	ErrNotCharBoundary   = errors.New("byte offset is not on a char boundary")
	ErrLineOutOfBounds   = errors.New("line out of bounds")
)

// OutOfBoundsError names the offset and the length that rejected it.
type OutOfBoundsError struct {
	Offset int
	Len    int
}

func (e OutOfBoundsError) Error() string {
	return fmt.Sprintf("%v: %d (len %d)", ErrOffsetOutOfBounds, e.Offset, e.Len)
}
func (e OutOfBoundsError) Unwrap() error { return ErrOffsetOutOfBounds }

// NotCharBoundaryError names the offset that split a rune.
type NotCharBoundaryError struct {
	Offset int
}

func (e NotCharBoundaryError) Error() string {
	return fmt.Sprintf("%v: %d", ErrNotCharBoundary, e.Offset)
}
func (e NotCharBoundaryError) Unwrap() error { return ErrNotCharBoundary }

// LineOutOfBoundsError names the line and the count that rejected it.
type LineOutOfBoundsError struct {
	Line      int
	LineCount int
}

func (e LineOutOfBoundsError) Error() string {
	return fmt.Sprintf("%v: %d (%d lines)", ErrLineOutOfBounds, e.Line, e.LineCount)
}
func (e LineOutOfBoundsError) Unwrap() error { return ErrLineOutOfBounds }

// Buffer is a byte slice plus the byte offset where each line begins.
//
// Editing is O(n) in the size of the buffer, which is a deliberate trade: the
// structure stays small enough to be obviously correct, and the operations the
// product performs are bounded by what a person can type. Large-file mode is a
// separate capability, not a reason to make every path harder to verify.
type Buffer struct {
	data []byte
	// starts[i] is the byte offset where line i begins. starts[0] is 0 and
	// len(starts) equals LineCount.
	starts []int
}

// New returns an empty buffer, which has one empty line.
func New() *Buffer {
	return &Buffer{starts: []int{0}}
}

// FromText returns a buffer holding text.
func FromText(text string) *Buffer {
	b := &Buffer{data: []byte(text)}
	b.reindex()
	return b
}

// LenBytes returns the size of the buffer in bytes.
func (b *Buffer) LenBytes() int { return len(b.data) }

// IsEmpty reports whether the buffer holds no bytes.
func (b *Buffer) IsEmpty() bool { return len(b.data) == 0 }

// LineCount returns the number of lines.
//
// An empty buffer has one line, and a trailing line break starts a new empty
// line: "abc\n" has two. Line breaks are "\n", "\r\n" and "\r", plus the
// Unicode separators U+0085, U+2028 and U+2029 — the same set the reference
// implementation counts, verified by a conformance case rather than assumed.
func (b *Buffer) LineCount() int { return len(b.starts) }

// String returns the whole buffer.
func (b *Buffer) String() string { return string(b.data) }

// CharCount returns the number of Unicode scalars.
func (b *Buffer) CharCount() int { return utf8.RuneCount(b.data) }

// LineText returns the text of a line without its line break.
func (b *Buffer) LineText(line int) (string, error) {
	if err := b.ensureLine(line); err != nil {
		return "", err
	}
	return string(b.lineContent(line)), nil
}

// LineToByte returns the byte offset where a line begins.
func (b *Buffer) LineToByte(line int) (Offset, error) {
	if err := b.ensureLine(line); err != nil {
		return 0, err
	}
	return Offset(b.starts[line]), nil
}

// ByteToCaret converts a byte offset into a line and scalar column.
//
// An offset inside a line's break reports the column just past the content:
// there is no caret position between the last character and the newline.
func (b *Buffer) ByteToCaret(offset Offset) (Caret, error) {
	if err := b.ensureOffsetBoundary(offset); err != nil {
		return Caret{}, err
	}
	line := b.lineAt(int(offset))
	content := b.lineContent(line)

	columnBytes := int(offset) - b.starts[line]
	if columnBytes > len(content) {
		columnBytes = len(content)
	}
	// Walk back to a rune boundary: an offset in the middle of a rune cannot
	// produce a column, and rounding down is the only answer that keeps the
	// caret inside the text. The end of the content is itself a boundary, so the
	// loop must not index past it.
	for columnBytes > 0 && columnBytes < len(content) && !utf8.RuneStart(content[columnBytes]) {
		columnBytes--
	}
	return Caret{Line: line, Column: utf8.RuneCount(content[:columnBytes])}, nil
}

// CaretToByte converts a line and scalar column into a byte offset.
//
// A column past the end of the line clamps to the end of its content, so a
// caret that outlived an edit still resolves.
func (b *Buffer) CaretToByte(caret Caret) (Offset, error) {
	if err := b.ensureLine(caret.Line); err != nil {
		return 0, err
	}
	content := b.lineContent(caret.Line)

	column := caret.Column
	if column < 0 {
		column = 0
	}

	byteInLine := 0
	counted := 0
	for index := 0; index < len(content) && counted < column; {
		_, size := utf8.DecodeRune(content[index:])
		index += size
		byteInLine = index
		counted++
	}

	return Offset(b.starts[caret.Line] + byteInLine), nil
}

// PrevCharOffset returns the offset where the rune before offset begins, or
// offset itself at the start of the buffer.
func (b *Buffer) PrevCharOffset(offset Offset) (Offset, error) {
	if err := b.ensureOffsetBoundary(offset); err != nil {
		return 0, err
	}
	if offset == 0 {
		return 0, nil
	}
	index := int(offset) - 1
	for index > 0 && !utf8.RuneStart(b.data[index]) {
		index--
	}
	return Offset(index), nil
}

// NextCharOffset returns the offset just past the rune at offset, or the end of
// the buffer.
func (b *Buffer) NextCharOffset(offset Offset) (Offset, error) {
	if err := b.ensureOffsetBoundary(offset); err != nil {
		return 0, err
	}
	if int(offset) >= len(b.data) {
		return offset, nil
	}
	_, size := utf8.DecodeRune(b.data[offset:])
	return offset + Offset(size), nil
}

// Insert puts text at at.
func (b *Buffer) Insert(at Offset, text string) error {
	if err := b.ensureOffsetBoundary(at); err != nil {
		return err
	}
	index := int(at)
	grown := make([]byte, 0, len(b.data)+len(text))
	grown = append(grown, b.data[:index]...)
	grown = append(grown, text...)
	grown = append(grown, b.data[index:]...)
	b.data = grown
	b.reindex()
	return nil
}

// TextRange returns the text in the half-open byte range [start, end).
func (b *Buffer) TextRange(start, end Offset) (string, error) {
	if start > end {
		return "", OutOfBoundsError{Offset: int(start), Len: len(b.data)}
	}
	if err := b.ensureOffsetBoundary(start); err != nil {
		return "", err
	}
	if err := b.ensureOffsetBoundary(end); err != nil {
		return "", err
	}
	return string(b.data[start:end]), nil
}

// DeleteRange removes the half-open byte range [start, end) and returns what it
// removed.
func (b *Buffer) DeleteRange(start, end Offset) (string, error) {
	if start > end {
		return "", OutOfBoundsError{Offset: int(start), Len: len(b.data)}
	}
	if err := b.ensureOffsetBoundary(start); err != nil {
		return "", err
	}
	if err := b.ensureOffsetBoundary(end); err != nil {
		return "", err
	}
	removed := string(b.data[start:end])

	remaining := make([]byte, 0, len(b.data)-(int(end)-int(start)))
	remaining = append(remaining, b.data[:start]...)
	remaining = append(remaining, b.data[end:]...)
	b.data = remaining
	b.reindex()
	return removed, nil
}

// ClampToCharBoundary moves offset back to the nearest rune boundary, never past
// the end of the buffer.
func (b *Buffer) ClampToCharBoundary(offset Offset) Offset {
	target := int(offset)
	if target > len(b.data) {
		target = len(b.data)
	}
	// The end of the buffer is a boundary even though it is not a rune start.
	for target > 0 && target < len(b.data) && !utf8.RuneStart(b.data[target]) {
		target--
	}
	return Offset(target)
}

// IsCharBoundary reports whether offset splits no rune. The end of the buffer
// is always a boundary.
func (b *Buffer) IsCharBoundary(offset Offset) bool {
	index := int(offset)
	if index < 0 || index > len(b.data) {
		return false
	}
	return index == len(b.data) || utf8.RuneStart(b.data[index])
}

// ---------------------------------------------------------------------------
// internals
// ---------------------------------------------------------------------------

// lineContent returns the line's bytes without its break.
func (b *Buffer) lineContent(line int) []byte {
	start := b.starts[line]
	end := len(b.data)
	if line+1 < len(b.starts) {
		end = b.starts[line+1]
	}
	content := b.data[start:end]
	return trimLineBreak(content)
}

// trimLineBreak removes one trailing line break, longest form first so CRLF is
// not mistaken for a lone CR.
func trimLineBreak(line []byte) []byte {
	if n := breakLengthAtEnd(line); n > 0 {
		return line[:len(line)-n]
	}
	return line
}

// breakLengthAtEnd returns the length of the line break ending line, or 0.
func breakLengthAtEnd(line []byte) int {
	if len(line) == 0 {
		return 0
	}
	if line[len(line)-1] == '\n' {
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			return 2
		}
		return 1
	}
	switch {
	case line[len(line)-1] == '\r':
		return 1
	case hasSuffixRune(line, '\u0085'), hasSuffixRune(line, '\u2028'), hasSuffixRune(line, '\u2029'):
		_, size := utf8.DecodeLastRune(line)
		return size
	}
	return 0
}

func hasSuffixRune(line []byte, want rune) bool {
	r, _ := utf8.DecodeLastRune(line)
	return r == want
}

// reindex recomputes the line starts.
func (b *Buffer) reindex() {
	b.starts = b.starts[:0]
	b.starts = append(b.starts, 0)
	for index := 0; index < len(b.data); {
		size := breakLengthAt(index, b.data)
		if size > 0 {
			index += size
			b.starts = append(b.starts, index)
			continue
		}
		_, runeSize := utf8.DecodeRune(b.data[index:])
		index += runeSize
	}
}

// breakLengthAt returns the length of the line break starting at index, or 0.
func breakLengthAt(index int, data []byte) int {
	if index >= len(data) {
		return 0
	}
	switch data[index] {
	case '\n':
		return 1
	case '\r':
		if index+1 < len(data) && data[index+1] == '\n' {
			return 2
		}
		return 1
	}
	r, size := utf8.DecodeRune(data[index:])
	if r == '\u0085' || r == '\u2028' || r == '\u2029' {
		return size
	}
	return 0
}

// lineAt returns the line containing a byte offset.
func (b *Buffer) lineAt(offset int) int {
	// The starts slice is sorted, so the line is the last start not past offset.
	low, high := 0, len(b.starts)-1
	for low < high {
		mid := (low + high + 1) / 2
		if b.starts[mid] <= offset {
			low = mid
		} else {
			high = mid - 1
		}
	}
	return low
}

func (b *Buffer) ensureOffsetBoundary(offset Offset) error {
	o := int(offset)
	if o < 0 {
		return OutOfBoundsError{Offset: o, Len: len(b.data)}
	}
	if o > len(b.data) {
		return OutOfBoundsError{Offset: o, Len: len(b.data)}
	}
	// The end of the buffer is always a valid position, even though it is not
	// the start of a rune.
	if o < len(b.data) && !utf8.RuneStart(b.data[o]) {
		return NotCharBoundaryError{Offset: o}
	}
	return nil
}

func (b *Buffer) ensureLine(line int) error {
	if line < 0 || line >= len(b.starts) {
		return LineOutOfBoundsError{Line: line, LineCount: len(b.starts)}
	}
	return nil
}

// Lines returns the text of every line without breaks, for tests and dumps.
func (b *Buffer) Lines() []string {
	out := make([]string, 0, len(b.starts))
	for i := range b.starts {
		out = append(out, string(b.lineContent(i)))
	}
	return out
}

// Equal reports whether two buffers hold the same bytes.
func (b *Buffer) Equal(other *Buffer) bool {
	if b == nil || other == nil {
		return b == other
	}
	return string(b.data) == string(other.data)
}
