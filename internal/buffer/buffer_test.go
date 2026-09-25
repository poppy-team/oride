package buffer

import (
	"errors"
	"testing"
	"unicode/utf8"
)

func TestEmptyBufferHasOneLine(t *testing.T) {
	b := New()
	if got := b.LineCount(); got != 1 {
		t.Errorf("LineCount() = %d, esperado 1", got)
	}
	if !b.IsEmpty() {
		t.Error("buffer novo não está vazio")
	}
	if got := b.String(); got != "" {
		t.Errorf("String() = %q", got)
	}
}

// TestTrailingBreakStartsANewLine pins the rule that decides every line-count
// assertion: a trailing break opens an empty final line, so "abc\n" has two
// lines and not one.
func TestTrailingBreakStartsANewLine(t *testing.T) {
	cases := []struct {
		text  string
		lines int
	}{
		{"", 1},
		{"abc", 1},
		{"abc\n", 2},
		{"abc\ndef", 2},
		{"abc\ndef\n", 3},
		{"\n", 2},
		{"\n\n", 3},
		{"a\r\nb\r\n", 3},
		{"a\rb", 2},
	}
	for _, tc := range cases {
		if got := FromText(tc.text).LineCount(); got != tc.lines {
			t.Errorf("%q: LineCount() = %d, esperado %d", tc.text, got, tc.lines)
		}
	}
}

func TestLineTextStripsTheBreak(t *testing.T) {
	b := FromText("um\r\ndois\ntres")
	want := []string{"um", "dois", "tres"}
	got := b.Lines()
	if len(got) != len(want) {
		t.Fatalf("Lines() = %q, esperado %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("linha %d = %q, esperado %q", i, got[i], want[i])
		}
	}
}

func TestInsertAndReadLines(t *testing.T) {
	b := New()
	if err := b.Insert(0, "hello\nworld"); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	for line, want := range map[int]string{0: "hello", 1: "world"} {
		got, err := b.LineText(line)
		if err != nil {
			t.Fatalf("LineText(%d): %v", line, err)
		}
		if got != want {
			t.Errorf("linha %d = %q, esperado %q", line, got, want)
		}
	}
	if b.String() != "hello\nworld" {
		t.Errorf("String() = %q", b.String())
	}
}

func TestDeleteRangeMiddle(t *testing.T) {
	b := FromText("abcdef")
	removed, err := b.DeleteRange(2, 4)
	if err != nil {
		t.Fatalf("DeleteRange: %v", err)
	}
	if removed != "cd" {
		t.Errorf("removido = %q, esperado \"cd\"", removed)
	}
	if b.String() != "abef" {
		t.Errorf("restou %q, esperado \"abef\"", b.String())
	}
}

func TestCaretRoundtripASCII(t *testing.T) {
	b := FromText("ab\ncd")
	offset, err := b.CaretToByte(Caret{Line: 1, Column: 1})
	if err != nil {
		t.Fatalf("CaretToByte: %v", err)
	}
	if offset != 4 {
		t.Errorf("CaretToByte = %d, esperado 4", offset)
	}
	caret, err := b.ByteToCaret(offset)
	if err != nil {
		t.Fatalf("ByteToCaret: %v", err)
	}
	if caret != (Caret{Line: 1, Column: 1}) {
		t.Errorf("ByteToCaret = %+v, esperado {1 1}", caret)
	}
}

func TestRejectsOffsetPastEnd(t *testing.T) {
	b := FromText("hi")
	_, err := b.ByteToCaret(99)
	if err == nil {
		t.Fatal("offset além do fim foi aceito")
	}
	if !errors.Is(err, ErrOffsetOutOfBounds) {
		t.Errorf("erro = %v, esperado ErrOffsetOutOfBounds", err)
	}
}

func TestRejectsOffsetInsideARune(t *testing.T) {
	// "é" is two bytes; offset 1 splits it.
	b := FromText("é")
	if _, err := b.ByteToCaret(1); err == nil {
		t.Fatal("offset no meio de um rune foi aceito")
	} else if !errors.Is(err, ErrNotCharBoundary) {
		t.Errorf("erro = %v, esperado ErrNotCharBoundary", err)
	}
	// The end of the buffer is a boundary even though it is not a rune start.
	if _, err := b.ByteToCaret(2); err != nil {
		t.Errorf("offset no fim do buffer foi rejeitado: %v", err)
	}
}

// TestCaretsCountScalarsNotBytes is the distinction the whole package exists to
// preserve: a caret column counts Unicode scalars, a byte offset counts bytes.
func TestCaretsCountScalarsNotBytes(t *testing.T) {
	cases := []struct {
		text   string
		column int
		offset Offset
	}{
		{"abc", 2, 2},     // one byte each
		{"ção", 2, 4},     // ç and ã are two bytes each, o is one
		{"日本", 1, 3},      // each ideograph is three bytes
		{"a🙂b", 2, 5},     // a is one byte, the emoji four
		{"e\u0301", 1, 1}, // combining acute: two scalars, three bytes
	}
	for _, tc := range cases {
		b := FromText(tc.text)
		got, err := b.CaretToByte(Caret{Line: 0, Column: tc.column})
		if err != nil {
			t.Fatalf("%q: CaretToByte: %v", tc.text, err)
		}
		if got != tc.offset {
			t.Errorf("%q coluna %d → offset %d, esperado %d", tc.text, tc.column, got, tc.offset)
		}
		back, err := b.ByteToCaret(tc.offset)
		if err != nil {
			t.Fatalf("%q: ByteToCaret(%d): %v", tc.text, tc.offset, err)
		}
		if back != (Caret{Line: 0, Column: tc.column}) {
			t.Errorf("%q offset %d → %+v, esperado {0 %d}", tc.text, tc.offset, back, tc.column)
		}
	}
}

// TestCaretRoundtripsAtEveryBoundary asserts the invariant that actually holds.
//
// The naive property — offset → caret → offset is the identity — is false, and
// cannot be true: a caret has no way to name a position inside a line break.
// What holds, and is what the rest of the editor depends on, is that a caret is
// a canonical representation: caret → offset → caret returns the same caret. A
// byte offset inside a break collapses to the end of the line's content and
// stays there.
func TestCaretRoundtripsAtEveryBoundary(t *testing.T) {
	texts := []string{
		"",
		"a",
		"abc",
		"ab\ncd\n",
		"ção\n日本語\n",
		"a🙂b\nc\u0301d\n",
		"linha um\r\nlinha dois\r\n",
		"fim sem quebra",
		"\n\n\n",
		"🙂\n🙂",
	}
	for _, text := range texts {
		b := FromText(text)
		for offset := Offset(0); offset <= Offset(len(text)); offset++ {
			// The end-of-buffer check comes first: indexing an empty string to
			// ask whether it is a rune start panics.
			if int(offset) != len(text) && !utf8.RuneStart(text[offset]) {
				continue
			}
			caret, err := b.ByteToCaret(offset)
			if err != nil {
				t.Errorf("%q offset %d: ByteToCaret: %v", text, offset, err)
				continue
			}
			back, err := b.CaretToByte(caret)
			if err != nil {
				t.Errorf("%q offset %d: CaretToByte(%+v): %v", text, offset, caret, err)
				continue
			}
			again, err := b.ByteToCaret(back)
			if err != nil {
				t.Errorf("%q offset %d: ByteToCaret(%d): %v", text, offset, back, err)
				continue
			}
			if again != caret {
				t.Errorf("%q: caret %+v → offset %d → caret %+v", text, caret, back, again)
			}
		}
	}
}

// TestOffsetInsideALineBreakCollapsesToEndOfContent records the lossiness
// explicitly, because it is the kind of thing a port silently changes.
//
// It is safe for editing — the selection carries exact byte offsets and the
// caret is a derived view for display — but it means the caret alone cannot
// distinguish "before the CR" from "before the LF".
func TestOffsetInsideALineBreakCollapsesToEndOfContent(t *testing.T) {
	b := FromText("linha um\r\nlinha dois\r\n")
	// Offsets 8 and 9 are the \r and the \n of the first line.
	for offset := Offset(8); offset <= 9; offset++ {
		caret, err := b.ByteToCaret(offset)
		if err != nil {
			t.Fatalf("ByteToCaret(%d): %v", offset, err)
		}
		if caret != (Caret{Line: 0, Column: 8}) {
			t.Errorf("offset %d → %+v, esperado {0 8}", offset, caret)
		}
		back, err := b.CaretToByte(caret)
		if err != nil {
			t.Fatalf("CaretToByte(%+v): %v", caret, err)
		}
		if back != 8 {
			t.Errorf("caret %+v → offset %d, esperado 8", caret, back)
		}
	}
}

// TestOffsetInTheBreakMapsToEndOfContent records what happens at the boundary
// between a line's text and its break: there is no caret between the last
// character and the newline, so the column is the end of the content.
func TestOffsetInTheBreakMapsToEndOfContent(t *testing.T) {
	b := FromText("ab\ncd")
	// Offset 2 is the newline itself.
	caret, err := b.ByteToCaret(2)
	if err != nil {
		t.Fatalf("ByteToCaret(2): %v", err)
	}
	if caret != (Caret{Line: 0, Column: 2}) {
		t.Errorf("offset da quebra → %+v, esperado {0 2}", caret)
	}
}

func TestPrevAndNextCharOffset(t *testing.T) {
	b := FromText("aé")
	// a = 0..1, é = 1..3
	next, err := b.NextCharOffset(0)
	if err != nil {
		t.Fatalf("NextCharOffset(0): %v", err)
	}
	if next != 1 {
		t.Errorf("NextCharOffset(0) = %d, esperado 1", next)
	}
	next, err = b.NextCharOffset(1)
	if err != nil {
		t.Fatalf("NextCharOffset(1): %v", err)
	}
	if next != 3 {
		t.Errorf("NextCharOffset(1) = %d, esperado 3", next)
	}
	// Past the end it is the identity, not an error.
	next, err = b.NextCharOffset(3)
	if err != nil {
		t.Fatalf("NextCharOffset(3): %v", err)
	}
	if next != 3 {
		t.Errorf("NextCharOffset no fim = %d, esperado 3", next)
	}

	prev, err := b.PrevCharOffset(3)
	if err != nil {
		t.Fatalf("PrevCharOffset(3): %v", err)
	}
	if prev != 1 {
		t.Errorf("PrevCharOffset(3) = %d, esperado 1", prev)
	}
	prev, err = b.PrevCharOffset(0)
	if err != nil {
		t.Fatalf("PrevCharOffset(0): %v", err)
	}
	if prev != 0 {
		t.Errorf("PrevCharOffset(0) = %d, esperado 0", prev)
	}
}

func TestLineToByte(t *testing.T) {
	b := FromText("ab\ncd\n")
	cases := map[int]Offset{0: 0, 1: 3, 2: 6}
	for line, want := range cases {
		got, err := b.LineToByte(line)
		if err != nil {
			t.Fatalf("LineToByte(%d): %v", line, err)
		}
		if got != want {
			t.Errorf("LineToByte(%d) = %d, esperado %d", line, got, want)
		}
	}
	if _, err := b.LineToByte(3); err == nil {
		t.Error("linha além do fim foi aceita")
	}
}

// TestLineIndexSurvivesEdits is the invariant the cached line starts depend on:
// every mutation must leave them consistent, or carets land on the wrong line
// after an edit.
func TestLineIndexSurvivesEdits(t *testing.T) {
	b := FromText("um\ndois\ntres")

	if err := b.Insert(2, "X\nY"); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	if got, want := b.String(), "umX\nY\ndois\ntres"; got != want {
		t.Fatalf("após insert = %q, esperado %q", got, want)
	}
	if got, want := b.LineCount(), 4; got != want {
		t.Fatalf("LineCount = %d, esperado %d", got, want)
	}

	if _, err := b.DeleteRange(0, 2); err != nil {
		t.Fatalf("DeleteRange: %v", err)
	}
	if got, want := b.String(), "X\nY\ndois\ntres"; got != want {
		t.Fatalf("após delete = %q, esperado %q", got, want)
	}
	if got, want := b.LineCount(), 4; got != want {
		t.Fatalf("LineCount = %d, esperado %d", got, want)
	}

	// Every line must still resolve, and the offsets must be monotonic.
	previous := Offset(-1)
	for line := 0; line < b.LineCount(); line++ {
		offset, err := b.LineToByte(line)
		if err != nil {
			t.Fatalf("LineToByte(%d): %v", line, err)
		}
		if offset <= previous {
			t.Errorf("LineToByte(%d) = %d, não é maior que o anterior %d", line, offset, previous)
		}
		previous = offset
	}
}

func TestInsertRejectsBadOffsets(t *testing.T) {
	b := FromText("é")
	if err := b.Insert(1, "x"); err == nil {
		t.Error("insert no meio de um rune foi aceito")
	}
	if err := b.Insert(99, "x"); err == nil {
		t.Error("insert além do fim foi aceito")
	}
	if err := b.Insert(2, "x"); err != nil {
		t.Errorf("insert no fim foi rejeitado: %v", err)
	}
}

func TestDeleteRangeRejectsReversedRange(t *testing.T) {
	b := FromText("abc")
	if _, err := b.DeleteRange(2, 1); err == nil {
		t.Error("intervalo invertido foi aceito")
	}
}

func TestTextRange(t *testing.T) {
	b := FromText("abcdef")
	got, err := b.TextRange(1, 4)
	if err != nil {
		t.Fatalf("TextRange: %v", err)
	}
	if got != "bcd" {
		t.Errorf("TextRange(1,4) = %q, esperado \"bcd\"", got)
	}
}

func TestClampToCharBoundary(t *testing.T) {
	b := FromText("ação")
	// "a"=0..1, "ç"=1..3
	if got := b.ClampToCharBoundary(2); got != 1 {
		t.Errorf("ClampToCharBoundary(2) = %d, esperado 1", got)
	}
	if got := b.ClampToCharBoundary(0); got != 0 {
		t.Errorf("ClampToCharBoundary(0) = %d, esperado 0", got)
	}
	if got := b.ClampToCharBoundary(99); got != Offset(len("ação")) {
		t.Errorf("ClampToCharBoundary(99) = %d, esperado o fim", got)
	}
}

func TestCharCountCountsScalars(t *testing.T) {
	b := FromText("aé日🙂")
	if got := b.CharCount(); got != 4 {
		t.Errorf("CharCount() = %d, esperado 4", got)
	}
	if got := b.LenBytes(); got != 1+2+3+4 {
		t.Errorf("LenBytes() = %d, esperado 10", got)
	}
}

// TestByteToCaretAcceptsExactlyTheBoundaries asserts the contract directly: an
// offset on a rune boundary (or at the end) converts, and anything else is
// refused. The boundary walks index into the content, and an off-by-one there is
// a panic in a text editor rather than a wrong answer.
func TestByteToCaretAcceptsExactlyTheBoundaries(t *testing.T) {
	texts := []string{"", "a", "ab\n", "\n", "é", "é\né", "🙂\n", "a\r\nb\r\n", "a\u2028b"}
	for _, text := range texts {
		b := FromText(text)
		for offset := Offset(0); offset <= Offset(len(text)); offset++ {
			isBoundary := int(offset) == len(text) || utf8.RuneStart(text[offset])

			caret, err := b.ByteToCaret(offset)
			switch {
			case isBoundary && err != nil:
				t.Errorf("%q offset %d é fronteira e falhou: %v", text, offset, err)
			case !isBoundary && err == nil:
				t.Errorf("%q offset %d não é fronteira e foi aceito como %+v", text, offset, caret)
			}
		}
	}
}
