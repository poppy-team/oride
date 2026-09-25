package menubar

import (
	"strings"
	"testing"

	"github.com/ori-team/oride/internal/tui/layout"
)

func TestExactWidth(t *testing.T) {
	views := []View{
		{},
		{Labels: []string{"Arquivo", "Editar", "Exibir", "Ir", "Git", "Ajuda"}, Open: -1},
		{Labels: []string{"Arquivo", "Editar"}, Open: 1},
	}
	for index, view := range views {
		for _, width := range []int{5, 20, 80, 200} {
			row := Render(width, view)
			if got := layout.Width(row); got != width {
				t.Errorf("caso %d largura %d: %d células", index, width, got)
			}
		}
	}
}

// TestMarksTheOpenMenu: the open menu is spelled out, not only tinted.
func TestMarksTheOpenMenu(t *testing.T) {
	row := Render(60, View{Labels: []string{"Arquivo", "Editar"}, Open: 1})

	if !strings.Contains(row, "[Editar]") {
		t.Errorf("o menu aberto não está marcado: %q", row)
	}
	if strings.Contains(row, "[Arquivo]") {
		t.Errorf("o menu fechado foi marcado: %q", row)
	}
}

// TestNoOpenMenuMarksNothing.
func TestNoOpenMenuMarksNothing(t *testing.T) {
	if row := Render(60, View{Labels: []string{"Arquivo"}, Open: -1}); strings.Contains(row, "[") {
		t.Errorf("marcou um menu com nenhum aberto: %q", row)
	}
}

// TestLabelsForKeepsTheDocumentedOrder: a translation supplies the words, and the
// bar decides the arrangement.
func TestLabelsForKeepsTheDocumentedOrder(t *testing.T) {
	labels := LabelsFor("F", "E", "V", "G", "S", "A")

	want := []string{"F", "E", "V", "G", "S", "A"}
	for index := range want {
		if labels[index] != want[index] {
			t.Fatalf("posição %d = %q, esperado %q", index, labels[index], want[index])
		}
	}
}
