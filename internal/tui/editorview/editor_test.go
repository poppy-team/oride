package editorview

import (
	"strings"
	"testing"

	"github.com/ori-team/oride/internal/config"
	"github.com/ori-team/oride/internal/tui/layout"
	"github.com/ori-team/oride/internal/tui/theme"
)

func lines(count int) []string {
	out := make([]string, count)
	for index := range out {
		out[index] = "linha " + string(rune('a'+index%26))
	}
	return out
}

// TestEveryRowHasTheExactWidth is the invariant that keeps a frame from shifting.
func TestEveryRowHasTheExactWidth(t *testing.T) {
	for _, gutter := range []bool{false, true} {
		for _, width := range []int{20, 40, 79, 80} {
			view := View{Lines: lines(10), Offset: 20, Gutter: gutter, HasCaret: true, Caret: Position{Line: 30}}
			for index, row := range Render(width, 10, view) {
				if got := layout.Width(row); got != width {
					t.Errorf("gutter=%v largura=%d: linha %d com %d células", gutter, width, index, got)
				}
			}
		}
	}
}

func TestRenderFillsTheRequestedHeight(t *testing.T) {
	view := View{Lines: lines(3), Gutter: true}

	rows := Render(40, 12, view)
	if len(rows) != 12 {
		t.Fatalf("linhas = %d, esperado 12", len(rows))
	}
}

func TestRenderHandlesZeroSize(t *testing.T) {
	for _, tc := range []struct{ width, height int }{{0, 10}, {10, 0}, {-1, 5}} {
		if rows := Render(tc.width, tc.height, View{Lines: lines(5)}); rows != nil {
			t.Errorf("%dx%d devolveu %d linhas", tc.width, tc.height, len(rows))
		}
	}
}

// TestTheCaretLineIsMarked: the position must be readable without colour.
func TestTheCaretLineIsMarked(t *testing.T) {
	view := View{Lines: lines(10), Offset: 2, Gutter: true, HasCaret: true, Caret: Position{Line: 4}}

	rows := Render(40, 10, view)
	// A linha 4 do documento é o índice 2 da fatia, porque a fatia começa na 2.
	if !strings.Contains(rows[2], caretMarker) {
		t.Errorf("a linha do caret não traz o marcador: %q", rows[2])
	}
	for _, index := range []int{0, 1, 3, 4} {
		if strings.Contains(rows[index], caretMarker) {
			t.Errorf("a linha %d trouxe o marcador sem ser a do caret: %q", index, rows[index])
		}
	}
}

// TestTheCaretIsMarkedWhereverItIs: the model extracts the slice, so whichever
// lines it hands over are the lines shown — and the caret must be marked among
// them.
func TestTheCaretIsMarkedWhereverItIs(t *testing.T) {
	for _, offset := range []int{0, 40, 90} {
		view := View{Lines: lines(10), Offset: offset, Gutter: true, HasCaret: true, Caret: Position{Line: offset + 3}}

		rows := Render(40, 10, view)
		if !strings.Contains(rows[3], caretMarker) {
			t.Errorf("offset %d: a linha do caret não trouxe o marcador: %q", offset, rows[3])
		}
	}
}

// TestGutterGivesWayBeforeTheText: a column of numbers with no text beside it is
// not a smaller editor, it is a broken one.
func TestGutterGivesWayBeforeTheText(t *testing.T) {
	view := View{Lines: lines(100), Gutter: true}

	rows := Render(4, 3, view)
	for _, row := range rows {
		if layout.Width(row) != 4 {
			t.Errorf("linha com %d células: %q", layout.Width(row), row)
		}
	}
}

// TestGutterWidthFollowsTheDocumentLength: numbering that re-widened while
// scrolling would shift every line of text sideways.
func TestGutterWidthFollowsTheDocumentLength(t *testing.T) {
	short := Render(40, 1, View{Lines: lines(1), Gutter: true})[0]
	long := Render(40, 1, View{Lines: lines(1), Offset: 1999, Gutter: true})[0]

	if layout.Width(short) != 40 || layout.Width(long) != 40 {
		t.Fatalf("larguras = %d e %d", layout.Width(short), layout.Width(long))
	}
	if short == long {
		t.Error("a gutter não mudou com o número de linhas")
	}
}

// TestWideTextIsTruncatedWithAMarker: a line that silently loses its tail is
// indistinguishable from one that ends there.
func TestWideTextIsTruncatedWithAMarker(t *testing.T) {
	view := View{Lines: []string{strings.Repeat("x", 500)}}

	row := Render(20, 1, view)[0]
	if layout.Width(row) != 20 {
		t.Fatalf("linha com %d células", layout.Width(row))
	}
	if !strings.Contains(row, "…") {
		t.Errorf("o corte não foi marcado: %q", row)
	}
}

// TestAnEmptyDocumentStillRenders: a file with no lines is a real state, and it
// must not produce an empty frame.
func TestAnEmptyDocumentStillRenders(t *testing.T) {
	rows := Render(30, 5, View{Gutter: true})

	if len(rows) != 5 {
		t.Fatalf("linhas = %d, esperado 5", len(rows))
	}
	for _, row := range rows {
		if layout.Width(row) != 30 {
			t.Errorf("linha com %d células", layout.Width(row))
		}
	}
}

// selection builds a view with a selection over the given document range.
func selection(startLine, startColumn, endLine, endColumn int) Selection {
	return Selection{
		Start: Position{Line: startLine, Column: startColumn},
		End:   Position{Line: endLine, Column: endColumn},
	}
}

// TestSelectionIsPaintedOnlyInsideItsRange is the rule that keeps the tint honest:
// a background across the whole row would claim the selection reaches the edge of
// the screen, which it does not.
func TestSelectionIsPaintedOnlyInsideItsRange(t *testing.T) {
	view := View{Lines: []string{"abcdefghij"}, Selection: selection(0, 2, 0, 5), Theme: colourTheme()}

	row := Render(20, 1, view)[0]
	if !strings.ContainsRune(row, 0x1b) {
		t.Fatalf("nada foi pintado: %q", row)
	}
	// As células fora da seleção continuam sem estilo: o prefixo e o sufixo não
	// podem carregar o realce.
	if strings.HasPrefix(row, "\x1b") {
		t.Errorf("o prefixo foi tingido junto: %q", row)
	}
	if got := layout.Width(row); got != 20 {
		t.Errorf("linha com %d células", got)
	}
}

// TestSelectionOnTheFirstAndLastLineIsPartial: every line between them is selected
// whole, and the two ends are not.
func TestSelectionOnTheFirstAndLastLineIsPartial(t *testing.T) {
	lines := []string{"primeira", "meio", "ultima"}
	view := View{Lines: lines, Selection: selection(0, 3, 2, 2), Theme: colourTheme()}

	rows := Render(20, 3, view)
	for index, row := range rows {
		if !strings.ContainsRune(row, 0x1b) {
			t.Errorf("linha %d não foi pintada: %q", index, row)
		}
		if got := layout.Width(row); got != 20 {
			t.Errorf("linha %d com %d células", index, got)
		}
	}
}

// TestAnEmptySelectionPaintsNothing: a collapsed selection is a caret, and tinting
// the character under the cursor would misrepresent it.
func TestAnEmptySelectionPaintsNothing(t *testing.T) {
	view := View{Lines: []string{"abcdef"}, Selection: Selection{Empty: true}, Theme: colourTheme()}

	if row := Render(20, 1, view)[0]; strings.ContainsRune(row, 0x1b) {
		t.Errorf("seleção vazia foi pintada: %q", row)
	}
}

// TestASelectionPastTheEndOfTheLineDoesNotBreakTheWidth: the arithmetic would go
// negative, and a negative width renders as garbage.
func TestASelectionPastTheEndOfTheLineDoesNotBreakTheWidth(t *testing.T) {
	cases := []Selection{
		selection(0, 50, 0, 90),
		selection(0, 0, 0, 500),
		selection(0, 3, 0, 1),
	}
	for _, sel := range cases {
		for _, width := range []int{10, 20, 40} {
			row := Render(width, 1, View{Lines: []string{"curta"}, Selection: sel, Theme: colourTheme()})[0]
			if got := layout.Width(row); got != width {
				t.Errorf("%+v largura %d: linha com %d células", sel, width, got)
			}
		}
	}
}

// TestSelectionSurvivesWideCharacters: the range is in cells, and slicing by byte
// would land in the middle of a wide character.
func TestSelectionSurvivesWideCharacters(t *testing.T) {
	view := View{Lines: []string{"日本語のテキスト"}, Selection: selection(0, 2, 0, 8), Theme: colourTheme()}

	for _, width := range []int{20, 30} {
		row := Render(width, 1, view)[0]
		if got := layout.Width(row); got != width {
			t.Errorf("largura %d: linha com %d células", width, got)
		}
		if !strings.ContainsRune(row, 0x1b) {
			t.Errorf("largura %d: nada foi pintado", width)
		}
	}
}

// TestSelectionFollowsTheSliceOffset: the viewport draws a slice, so a selection
// on document line 40 must paint when that line is the slice's first row.
func TestSelectionFollowsTheSliceOffset(t *testing.T) {
	view := View{
		Lines:     []string{"alvo"},
		Offset:    40,
		Selection: selection(40, 0, 40, 4),
		Theme:     colourTheme(),
	}

	if row := Render(20, 1, view)[0]; !strings.ContainsRune(row, 0x1b) {
		t.Errorf("a seleção não seguiu o deslocamento da fatia: %q", row)
	}
}

// colourTheme is the shipped palette at full colour, so the tests exercise the
// painting path rather than only the plain one.
func colourTheme() theme.Theme {
	shipped := config.Default()
	return theme.New(shipped.UI, shipped.Syntax, theme.TrueColor)
}
