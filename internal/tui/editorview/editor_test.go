package editorview

import (
	"strings"
	"testing"

	"github.com/ori-team/oride/internal/tui/layout"
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
