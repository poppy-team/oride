package tabs

import (
	"strconv"
	"strings"
	"testing"

	"github.com/ori-team/oride/internal/tui/layout"
)

func exact(t *testing.T, name, row string, width int) {
	t.Helper()
	if got := layout.Width(row); got != width {
		t.Errorf("%s: %d células, esperado %d em %q", name, got, width, row)
	}
}

// TestExactWidth: a row that comes back short or long shifts everything beside it.
func TestExactWidth(t *testing.T) {
	views := []View{
		{},
		{Tabs: []Tab{{Title: "a.go", Active: true}}},
		{Tabs: []Tab{{Title: "a.go", Active: true}, {Title: "b.go", Dirty: true}}},
		{Tabs: []Tab{{Title: strings.Repeat("longo", 30)}, {Title: "curto", Active: true}}},
	}
	for index, view := range views {
		for _, width := range []int{10, 40, 80, 200} {
			exact(t, "caso "+strconv.Itoa(index), Render(width, view), width)
		}
	}
}

// TestMarksDirtyAndActiveInText: neither state may depend on colour.
func TestMarksDirtyAndActiveInText(t *testing.T) {
	row := Render(60, View{Tabs: []Tab{{Title: "a.go", Dirty: true, Active: true}}})

	if !strings.Contains(row, dirtySuffix) {
		t.Errorf("o estado sujo não está escrito: %q", row)
	}
	if !strings.Contains(row, "[a.go*]") {
		t.Errorf("a aba ativa não está marcada: %q", row)
	}
}

// TestKeepsTheActiveTabWhenTheyDoNotFit: hiding which document is being edited is
// worse than hiding the others.
func TestKeepsTheActiveTabWhenTheyDoNotFit(t *testing.T) {
	view := View{Tabs: []Tab{
		{Title: strings.Repeat("x", 40)},
		{Title: "ativo.go", Active: true},
	}}

	if row := Render(20, view); !strings.Contains(row, "ativo") {
		t.Errorf("a aba ativa sumiu: %q", row)
	}
}

// TestAnEmptyBarSaysSo: a blank bar is indistinguishable from one that lost its
// documents.
func TestAnEmptyBarSaysSo(t *testing.T) {
	if row := Render(40, View{}); !strings.Contains(row, "sem documentos") {
		t.Errorf("barra sem abas sem aviso: %q", row)
	}
}

func TestZeroWidthRendersNothing(t *testing.T) {
	if got := Render(0, View{Tabs: []Tab{{Title: "a.go"}}}); got != "" {
		t.Errorf("largura zero devolveu %q", got)
	}
}
