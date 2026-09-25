package component

import (
	"testing"

	"github.com/ori-team/oride/internal/tui/layout"
)

// TestWindowKeepsTheSelectionVisible is the whole point of the helper.
func TestWindowKeepsTheSelectionVisible(t *testing.T) {
	start, end := Window(100, 10, 55, 0)

	if 55 < start || 55 >= end {
		t.Errorf("janela [%d,%d) não contém a seleção 55", start, end)
	}
}

// TestWindowDoesNotJumpWhileTheSelectionIsInside: a list that recentres on every
// arrow key is exhausting to read.
func TestWindowDoesNotJumpWhileTheSelectionIsInside(t *testing.T) {
	_, end := Window(100, 10, 3, 0)

	if start, _ := Window(100, 10, 7, 0); start != 0 || end != 10 {
		t.Errorf("a janela pulou para [%d,%d) com a seleção ainda dentro", start, end)
	}
}

// TestWindowFollowsTheSelectionPastTheEdge.
func TestWindowFollowsTheSelectionPastTheEdge(t *testing.T) {
	start, end := Window(100, 10, 12, 0)

	if 12 < start || 12 >= end {
		t.Errorf("janela [%d,%d) perdeu a seleção 12", start, end)
	}
}

// TestWindowNeverRunsPastTheEnd: following the selection can push the window
// beyond the last item, which would leave blank rows below it.
func TestWindowNeverRunsPastTheEnd(t *testing.T) {
	start, end := Window(30, 10, 29, 0)

	if end != 30 {
		t.Errorf("fim = %d, esperado 30", end)
	}
	if start < 0 || start > end {
		t.Errorf("janela inválida [%d,%d)", start, end)
	}
}

func TestWindowHandlesEmptyAndZeroHeightLists(t *testing.T) {
	for _, tc := range []struct{ total, height int }{{0, 10}, {10, 0}, {-1, 5}, {5, -1}} {
		start, end := Window(tc.total, tc.height, 0, 0)
		if start != 0 || end != 0 {
			t.Errorf("Window(%d, %d) = [%d,%d), esperado vazio", tc.total, tc.height, start, end)
		}
	}
}

func TestWindowClampsANegativeScroll(t *testing.T) {
	start, end := Window(100, 10, 5, -20)

	if start < 0 {
		t.Errorf("início = %d, negativo", start)
	}
	if 5 < start || 5 >= end {
		t.Errorf("janela [%d,%d) perdeu a seleção 5", start, end)
	}
}

// TestRowReservesTheCursorColumnWhetherSelectedOrNot: a list whose text shifts
// when the selection moves is exhausting to read.
func TestRowReservesTheCursorColumnWhetherSelectedOrNot(t *testing.T) {
	selected := Row("arquivo.go", 0, true, 30)
	plain := Row("arquivo.go", 0, false, 30)

	if layout.Width(selected) != 30 || layout.Width(plain) != 30 {
		t.Fatalf("larguras = %d e %d", layout.Width(selected), layout.Width(plain))
	}
	// A coluna em CÉLULAS, não em bytes: o cursor é um caractere de vários bytes,
	// e comparar índices de byte acusa uma diferença que não existe na tela.
	selectedColumn := cellColumn(selected, "arquivo")
	plainColumn := cellColumn(plain, "arquivo")
	if selectedColumn != plainColumn {
		t.Errorf("o texto mudou de coluna: %d vs %d", selectedColumn, plainColumn)
	}
	if indexOf(selected, Mark) < 0 {
		t.Error("a linha selecionada não traz o cursor")
	}
	if indexOf(plain, Mark) >= 0 {
		t.Error("a linha não selecionada traz o cursor")
	}
}

// cellColumn is the cell offset where a substring starts, or -1.
//
// The layout measures in cells and Go indexes in bytes, and mixing the two is how
// a width bug hides behind a passing test.
func cellColumn(text, needle string) int {
	index := indexOf(text, needle)
	if index < 0 {
		return -1
	}
	return layout.Width(text[:index])
}

func TestRowIndentsByDepth(t *testing.T) {
	shallow := cellColumn(Row("x.go", 0, false, 30), "x.go")
	deep := cellColumn(Row("x.go", 2, false, 30), "x.go")

	if deep <= shallow {
		t.Errorf("profundidade 2 não recuou: %d vs %d", deep, shallow)
	}
}

// indexOf locates a substring, treating a missing one as -1.
func indexOf(text, needle string) int {
	for index := 0; index+len(needle) <= len(text); index++ {
		if text[index:index+len(needle)] == needle {
			return index
		}
	}
	return -1
}
