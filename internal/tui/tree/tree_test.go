package tree

import (
	"strings"
	"testing"

	"github.com/ori-team/oride/internal/tui/layout"
)

func sample() View {
	return View{
		Rows:     []Row{{Name: "src", IsDir: true, Expanded: true}, {Name: "a.go", Depth: 1}},
		Selected: 1,
	}
}

func TestExactWidthAndHeight(t *testing.T) {
	for _, width := range []int{10, 30, 80} {
		rows := Render(width, 5, sample())
		if len(rows) != 5 {
			t.Fatalf("largura %d: %d linhas", width, len(rows))
		}
		for index, row := range rows {
			if got := layout.Width(row); got != width {
				t.Errorf("largura %d linha %d: %d células", width, index, got)
			}
		}
	}
}

// TestMarksDirectoriesInText: the affordance is what tells a reader the entry can
// be opened, so it cannot be colour or an ambiguous-width glyph.
func TestMarksDirectoriesInText(t *testing.T) {
	view := View{Rows: []Row{
		{Name: "aberto", IsDir: true, Expanded: true},
		{Name: "fechado", IsDir: true},
		{Name: "arquivo.go"},
	}}

	rows := Render(30, 3, view)
	if !strings.Contains(rows[0], expandedDir) {
		t.Errorf("diretório aberto sem marcador: %q", rows[0])
	}
	if !strings.Contains(rows[1], collapsedDir) {
		t.Errorf("diretório fechado sem marcador: %q", rows[1])
	}
	if strings.Contains(rows[2], expandedDir) || strings.Contains(rows[2], collapsedDir) {
		t.Errorf("arquivo marcado como diretório: %q", rows[2])
	}
}

// TestSaysEmptyInsteadOfBlank: a blank rectangle does not distinguish an empty
// panel from one that failed to load.
func TestSaysEmptyInsteadOfBlank(t *testing.T) {
	if rows := Render(30, 3, View{}); !strings.Contains(rows[0], emptyMessage) {
		t.Errorf("painel vazio sem aviso: %q", rows[0])
	}
}

// TestTheSelectionStaysVisible: the window follows the selection, which is what
// the panel is for.
func TestTheSelectionStaysVisible(t *testing.T) {
	rows := make([]Row, 40)
	for index := range rows {
		rows[index] = Row{Name: "arquivo"}
	}
	view := View{Rows: rows, Selected: 35}

	rendered := Render(30, 10, view)
	marked := -1
	for index, row := range rendered {
		if strings.Contains(row, "▶") {
			marked = index
		}
	}
	if marked < 0 {
		t.Fatal("a linha selecionada não apareceu")
	}
}

func TestZeroSizeRendersNothing(t *testing.T) {
	if rows := Render(0, 10, sample()); rows != nil {
		t.Errorf("largura zero devolveu %d linhas", len(rows))
	}
	if rows := Render(10, 0, sample()); rows != nil {
		t.Errorf("altura zero devolveu %d linhas", len(rows))
	}
}
