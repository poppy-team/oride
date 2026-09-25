package statusbar

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

func TestExactWidth(t *testing.T) {
	views := []View{
		{},
		{File: "a.go", Dirty: true, Focus: "editor", CaretLine: 12, CaretCol: 4, HasCaret: true},
		{File: strings.Repeat("pasta/", 20) + "a.go", Message: strings.Repeat("mensagem ", 10), HasCaret: true},
	}
	for index, view := range views {
		for _, width := range []int{10, 40, 80, 200} {
			exact(t, "caso "+strconv.Itoa(index), Render(width, view), width)
		}
	}
}

// TestKeepsTheCaretPosition: the position is the field read by location, so it
// survives truncation of everything else.
func TestKeepsTheCaretPosition(t *testing.T) {
	view := View{
		File:    strings.Repeat("muito/longo/", 20) + "arquivo.go",
		Message: strings.Repeat("aviso ", 20), Focus: "editor",
		CaretLine: 7, CaretCol: 3, HasCaret: true,
	}

	if row := Render(60, view); !strings.Contains(row, "Ln 7, Col 3") {
		t.Errorf("a posição do cursor não sobreviveu: %q", row)
	}
}

// TestMarksDirtyInText: no state is carried by colour alone.
func TestMarksDirtyInText(t *testing.T) {
	dirty := Render(40, View{File: "a.go", Dirty: true})
	clean := Render(40, View{File: "a.go"})

	if !strings.Contains(dirty, dirtyMark) || strings.Contains(clean, dirtyMark) {
		t.Errorf("sujo=%q limpo=%q", dirty, clean)
	}
}

// TestNoDocumentSaysUntitled.
func TestNoDocumentSaysUntitled(t *testing.T) {
	if row := Render(40, View{}); !strings.Contains(row, untitled) {
		t.Errorf("sem documento sem aviso: %q", row)
	}
}

// TestNoCaretOmitsThePosition: a bar claiming "Ln 1, Col 1" with nothing open
// would be describing a cursor that does not exist.
func TestNoCaretOmitsThePosition(t *testing.T) {
	if row := Render(60, View{File: "a.go"}); strings.Contains(row, "Ln") {
		t.Errorf("posição exibida sem cursor: %q", row)
	}
}
