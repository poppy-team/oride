package editor

import (
	"strings"
	"testing"

	"github.com/ori-team/oride/internal/buffer"
)

func newDoc(text string) *Document {
	return NewDocumentFromText(1, "", text)
}

func mustInsert(t *testing.T, d *Document, text string) {
	t.Helper()
	if err := d.InsertText(text); err != nil {
		t.Fatalf("InsertText(%q): %v", text, err)
	}
}

func mustDo(t *testing.T, d *Document, what string, action func() (bool, error)) bool {
	t.Helper()
	changed, err := action()
	if err != nil {
		t.Fatalf("%s: %v", what, err)
	}
	return changed
}

// TestUndoLabelWording pins the exact strings the history shows.
//
// They are observable state: a case comparing undo labels against the oracle
// fails on a wording difference, including the typographic quotes.
func TestUndoLabelWording(t *testing.T) {
	cases := []struct {
		name   string
		action func(t *testing.T, d *Document)
		want   string
	}{
		{
			name:   "single short insert",
			action: func(t *testing.T, d *Document) { mustInsert(t, d, "abc") },
			want:   "insert “abc”",
		},
		{
			name: "insert crossing lines",
			action: func(t *testing.T, d *Document) {
				mustInsert(t, d, "um\ndois\ntres")
			},
			want: "insert 3 lines",
		},
		{
			name: "delete",
			action: func(t *testing.T, d *Document) {
				if err := d.DeleteForward(); err != nil {
					t.Fatalf("DeleteForward: %v", err)
				}
			},
			want: "delete “a”",
		},
		{
			name: "coalesced inserts",
			action: func(t *testing.T, d *Document) {
				mustInsert(t, d, "a")
				mustInsert(t, d, "b")
				mustInsert(t, d, "c")
			},
			want: "insert “a…” (+2 edits)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := newDoc("abc")
			if tc.name == "delete" {
				d = newDoc("a")
			}
			tc.action(t, d)

			// Committed first: the open group is labelled distinctly on purpose,
			// and TestOpenGroupIsLabelledDistinctly covers that marker.
			d.CommitEditGroup()

			labels := d.UndoHistoryLabels()
			if len(labels) != 1 {
				t.Fatalf("rótulos = %q, esperado exatamente um", labels)
			}
			if labels[0] != tc.want {
				t.Errorf("rótulo = %q, esperado %q", labels[0], tc.want)
			}
		})
	}
}

// TestOpenGroupIsLabelledDistinctly records that the uncommitted group is shown
// last with a marker, because it is the step an undo would take next.
func TestOpenGroupIsLabelledDistinctly(t *testing.T) {
	d := newDoc("")
	mustInsert(t, d, "abc")
	// Nothing has closed the group yet.
	if labels := d.UndoHistoryLabels(); len(labels) != 1 || !strings.HasPrefix(labels[0], "(aberto) ") {
		t.Fatalf("rótulos = %q, esperado um marcado como aberto", labels)
	}

	d.CommitEditGroup()
	labels := d.UndoHistoryLabels()
	if len(labels) != 1 || strings.HasPrefix(labels[0], "(aberto) ") {
		t.Fatalf("após commit, rótulos = %q", labels)
	}
}

// TestEnterClosesTheUndoGroup is ledger item B1: the reference coalesced abc,
// a newline and xyz into one group, so a single undo erased all of it.
//
// Two groups, not three: the newline closes the group it belongs to, because it
// completes the line the user just typed. The point is that it does not swallow
// what comes after.
func TestEnterClosesTheUndoGroup(t *testing.T) {
	d := newDoc("")
	mustInsert(t, d, "abc")
	mustInsert(t, d, "\n")
	mustInsert(t, d, "xyz")

	if labels := d.UndoHistoryLabels(); len(labels) != 2 {
		t.Fatalf("rótulos = %q, esperado dois grupos", labels)
	}

	mustDo(t, d, "undo", d.Undo)
	if got := d.Buffer().String(); got != "abc\n" {
		t.Fatalf("após um undo = %q, esperado \"abc\\n\" — o newline ficou no primeiro grupo", got)
	}
	mustDo(t, d, "undo", d.Undo)
	if got := d.Buffer().String(); got != "" {
		t.Fatalf("após dois undos = %q, esperado vazio", got)
	}
}

// TestMovementClosesTheUndoGroup: typing, moving and typing again are two
// thoughts, and one undo must not erase both.
func TestMovementClosesTheUndoGroup(t *testing.T) {
	d := newDoc("alpha\n")
	if err := d.MoveBufferEnd(false); err != nil {
		t.Fatalf("MoveBufferEnd: %v", err)
	}
	mustInsert(t, d, "AAA")
	if err := d.MoveLeft(false); err != nil {
		t.Fatalf("MoveLeft: %v", err)
	}
	mustInsert(t, d, "BBB")

	mustDo(t, d, "undo", d.Undo)
	if got := d.Buffer().String(); got != "alpha\nAAA" {
		t.Fatalf("após um undo = %q, esperado \"alpha\\nAAA\"", got)
	}
}

// TestRedoRestoresTheCaret is ledger item B2: the reference left the caret where
// the undo clamped it instead of following the restored text.
func TestRedoRestoresTheCaret(t *testing.T) {
	d := newDoc("")
	mustInsert(t, d, "abc")
	mustInsert(t, d, "\n")
	mustInsert(t, d, "xyz")

	before, err := d.Caret()
	if err != nil {
		t.Fatalf("Caret: %v", err)
	}

	mustDo(t, d, "undo", d.Undo)
	mustDo(t, d, "redo", d.Redo)

	after, err := d.Caret()
	if err != nil {
		t.Fatalf("Caret: %v", err)
	}
	if after != before {
		t.Errorf("caret após redo = %+v, esperado %+v (fim do texto restaurado)", after, before)
	}
	if got := d.Buffer().String(); got != "abc\nxyz" {
		t.Errorf("texto após redo = %q", got)
	}
}

// TestUndoDropsCursorsPastTheEndOfTheBuffer is ledger item B4: the reference kept
// secondary cursors that the shrinking buffer had invalidated, so the next
// insert failed.
func TestUndoDropsCursorsPastTheEndOfTheBuffer(t *testing.T) {
	d := newDoc("um\ndois\ntres")
	if err := d.MoveBufferEnd(false); err != nil {
		t.Fatalf("MoveBufferEnd: %v", err)
	}
	if err := d.AddCursorAbove(); err != nil {
		t.Fatalf("AddCursorAbove: %v", err)
	}
	mustInsert(t, d, "XXXXXXXX")

	for _, caret := range d.AllCaretOffsets() {
		if int(caret) > d.Buffer().LenBytes() {
			t.Fatalf("caret %d além do fim do buffer (%d) antes do undo", caret, d.Buffer().LenBytes())
		}
	}

	mustDo(t, d, "undo", d.Undo)

	for _, caret := range d.AllCaretOffsets() {
		if int(caret) > d.Buffer().LenBytes() {
			t.Errorf("caret %d além do fim do buffer (%d) após o undo",
				caret, d.Buffer().LenBytes())
		}
	}
	// And the next edit must still work rather than fail on a stale offset.
	mustInsert(t, d, "ok")
}

func TestMultiCursorInsert(t *testing.T) {
	d := newDoc("um\ndois\ntres")
	if err := d.MoveBufferEnd(false); err != nil {
		t.Fatalf("MoveBufferEnd: %v", err)
	}
	if err := d.AddCursorAbove(); err != nil {
		t.Fatalf("AddCursorAbove: %v", err)
	}
	if err := d.AddCursorAbove(); err != nil {
		t.Fatalf("AddCursorAbove: %v", err)
	}
	if got := len(d.AllCaretOffsets()); got != 3 {
		t.Fatalf("cursores = %d, esperado 3", got)
	}

	mustInsert(t, d, "!")

	want := "um!\ndois!\ntres!"
	if got := d.Buffer().String(); got != want {
		t.Errorf("texto = %q, esperado %q", got, want)
	}
}

func TestMultiCursorBackspace(t *testing.T) {
	d := newDoc("um\ndois\ntres")
	if err := d.MoveBufferEnd(false); err != nil {
		t.Fatalf("MoveBufferEnd: %v", err)
	}
	if err := d.AddCursorAbove(); err != nil {
		t.Fatalf("AddCursorAbove: %v", err)
	}
	if err := d.AddCursorAbove(); err != nil {
		t.Fatalf("AddCursorAbove: %v", err)
	}

	if err := d.Backspace(); err != nil {
		t.Fatalf("Backspace: %v", err)
	}
	// Each cursor removes the last character of its line.
	want := "u\ndoi\ntre"
	if got := d.Buffer().String(); got != want {
		t.Errorf("texto = %q, esperado %q", got, want)
	}
}

func TestCursorAtEndOfBufferSurvivesBackspace(t *testing.T) {
	d := newDoc("")
	if err := d.Backspace(); err != nil {
		t.Fatalf("Backspace em buffer vazio: %v", err)
	}
	if err := d.DeleteForward(); err != nil {
		t.Fatalf("DeleteForward em buffer vazio: %v", err)
	}
	if d.IsDirty() {
		t.Error("backspace num buffer vazio marcou o documento como sujo")
	}
}

func TestBackspaceJoinsLines(t *testing.T) {
	d := newDoc("ab\ncd")
	if err := d.MoveBufferEnd(false); err != nil {
		t.Fatalf("MoveBufferEnd: %v", err)
	}
	for range 5 {
		if err := d.Backspace(); err != nil {
			t.Fatalf("Backspace: %v", err)
		}
	}
	if got := d.Buffer().String(); got != "" {
		t.Errorf("texto = %q, esperado vazio", got)
	}
}

// TestPreferredColumnSurvivesAShortLine is the behaviour the column exists for:
// moving through a line shorter than the caret's column must come back out.
func TestPreferredColumnSurvivesAShortLine(t *testing.T) {
	d := newDoc("abcdef\nab\nabcdef")
	// Caret at column 5 of line 0.
	offset, err := d.Buffer().CaretToByte(buffer.Caret{Line: 0, Column: 5})
	if err != nil {
		t.Fatalf("CaretToByte: %v", err)
	}
	d.JumpToByte(offset)

	if err := d.MoveDown(false); err != nil {
		t.Fatalf("MoveDown: %v", err)
	}
	// Line 1 has two characters, so the caret clamps to column 2.
	if caret, _ := d.Caret(); caret.Column != 2 {
		t.Fatalf("na linha curta, coluna = %d, esperado 2", caret.Column)
	}
	if err := d.MoveDown(false); err != nil {
		t.Fatalf("MoveDown: %v", err)
	}
	// Back on a long line, the original column returns.
	if caret, _ := d.Caret(); caret.Column != 5 {
		t.Errorf("de volta à linha longa, coluna = %d, esperado 5", caret.Column)
	}
}

func TestSelectAllAndDeleteSelection(t *testing.T) {
	d := newDoc("abc\ndef")
	d.SelectAll()

	if got := d.SelectedText(); got != "abc\ndef" {
		t.Fatalf("SelectedText = %q", got)
	}
	if err := d.DeleteSelection(); err != nil {
		t.Fatalf("DeleteSelection: %v", err)
	}
	if got := d.Buffer().String(); got != "" {
		t.Errorf("texto = %q, esperado vazio", got)
	}
	if !d.IsDirty() {
		t.Error("apagar a seleção não marcou o documento como sujo")
	}

	mustDo(t, d, "undo", d.Undo)
	if got := d.Buffer().String(); got != "abc\ndef" {
		t.Errorf("após undo = %q", got)
	}
}

func TestVersionCountsModifications(t *testing.T) {
	d := newDoc("")
	if d.Version() != 0 {
		t.Fatalf("versão inicial = %d", d.Version())
	}
	mustInsert(t, d, "a")
	if d.Version() != 1 {
		t.Errorf("após uma edição, versão = %d, esperado 1", d.Version())
	}
	mustInsert(t, d, "b")
	if d.Version() != 2 {
		t.Errorf("após duas edições, versão = %d, esperado 2", d.Version())
	}
}

func TestTabTitle(t *testing.T) {
	if got := NewDocument(1).TabTitle(); got != "untitled" {
		t.Errorf("documento sem path: título = %q, esperado \"untitled\"", got)
	}
	if got := NewDocumentFromText(2, "/tmp/projeto/notas.txt", "").TabTitle(); got != "notas.txt" {
		t.Errorf("título = %q, esperado \"notas.txt\"", got)
	}
}

func TestStoreOpensAndCyclesTabs(t *testing.T) {
	s := NewStore()
	first := s.OpenEmpty()
	second := s.OpenEmpty()

	if got, _ := s.ActiveID(); got != second {
		t.Fatalf("ativo = %d, esperado %d", got, second)
	}
	if _, ok := s.ActivateNextTab(); !ok {
		t.Fatal("ActivateNextTab falhou")
	}
	if got, _ := s.ActiveID(); got != first {
		t.Errorf("após ciclo, ativo = %d, esperado %d", got, first)
	}
	if s.Len() != 2 {
		t.Errorf("Len = %d, esperado 2", s.Len())
	}
}

func TestStoreCloseSelectsANeighbour(t *testing.T) {
	s := NewStore()
	first := s.OpenEmpty()
	second := s.OpenEmpty()

	remaining, err := s.Close(second)
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
	if !remaining {
		t.Fatal("fechar uma de duas abas não deveria esvaziar o store")
	}
	if got, _ := s.ActiveID(); got != first {
		t.Errorf("ativo após fechar = %d, esperado %d", got, first)
	}

	remaining, err = s.Close(first)
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
	if remaining {
		t.Error("fechar a última aba deveria esvaziar o store")
	}
	if _, ok := s.ActiveID(); ok {
		t.Error("store vazio ainda tem aba ativa")
	}
}
