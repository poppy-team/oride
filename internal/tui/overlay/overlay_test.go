package overlay

import (
	"strings"
	"testing"

	"github.com/ori-team/oride/internal/tui/layout"
)

func sample() ListView {
	return ListView{
		Title: "Comandos",
		Hint:  "digite para filtrar",
		Items: []Item{
			{Label: "undo", Selected: true},
			{Label: "redo"},
			{Label: "save", Detail: "arquivo"},
		},
	}
}

// TestEveryRowHasTheExactWidth keeps the overlay from shifting the frame it floats
// over.
func TestEveryRowHasTheExactWidth(t *testing.T) {
	for _, width := range []int{20, 40, 76, 120} {
		for index, row := range List(width, 8, sample()) {
			if got := layout.Width(row); got != listInnerWidth(width) {
				t.Errorf("largura %d linha %d: %d células", width, index, got)
			}
		}
	}
}

// TestHeightIsRespected: the composition root paints into a fixed region, and a
// list that came back taller would write past it.
func TestHeightIsRespected(t *testing.T) {
	for _, height := range []int{1, 3, 8, 20} {
		if got := len(List(60, height, sample())); got != height {
			t.Errorf("altura %d devolveu %d linhas", height, got)
		}
	}
}

// TestCapturesEverything is the first rule of the focus graph, expressed as data.
func TestCapturesEverything(t *testing.T) {
	for _, kind := range []Kind{Palette, WhichKey, Help, Welcome, QuitConfirm, CloseConfirm} {
		if !kind.Captures() || !kind.Active() {
			t.Errorf("%v deveria capturar a entrada", kind)
		}
	}
	if None.Captures() || None.Active() {
		t.Error("nenhuma sobreposição não deveria capturar nada")
	}
}

// TestTheHintIsAlwaysTheLastRow: a hint that moves is a hint nobody finds twice.
func TestTheHintIsAlwaysTheLastRow(t *testing.T) {
	for _, count := range []int{0, 1, 5, 30} {
		view := sample()
		view.Items = view.Items[:min(count, len(view.Items))]
		for count > len(view.Items) {
			view.Items = append(view.Items, Item{Label: "x"})
		}

		rows := List(60, 8, view)
		if got := strings.TrimSpace(rows[len(rows)-1]); got != view.Hint {
			t.Errorf("%d itens: última linha = %q, esperado a dica", count, got)
		}
	}
}

// TestEmptySaysSomething: a blank box is indistinguishable from one that failed to
// load.
func TestEmptySaysSomething(t *testing.T) {
	view := ListView{Title: "Comandos", Empty: "(nenhum resultado)"}

	rows := List(60, 5, view)
	if !strings.Contains(strings.Join(rows, "\n"), "(nenhum resultado)") {
		t.Errorf("lista vazia sem aviso: %q", rows)
	}
}

// TestTheSelectedRowIsMarkedInText.
func TestTheSelectedRowIsMarkedInText(t *testing.T) {
	rows := List(60, 5, sample())

	if !strings.Contains(rows[1], marker) {
		t.Errorf("a linha selecionada não traz o cursor: %q", rows[1])
	}
	for _, index := range []int{2, 3} {
		if strings.Contains(rows[index], marker) {
			t.Errorf("a linha %d trouxe o cursor sem estar selecionada: %q", index, rows[index])
		}
	}
}

// TestTheBoxLeavesMargin: an overlay that covers everything is not an overlay, it
// is another screen — and the reader loses sight of what is behind it.
func TestTheBoxLeavesMargin(t *testing.T) {
	area := layout.Region{X: 0, Y: 0, Width: 120, Height: 40}
	region := Region(area, 100, 20)

	if region.Width >= area.Width {
		t.Errorf("a caixa ocupou a largura toda: %d de %d", region.Width, area.Width)
	}
	if region.X <= area.X || region.Y <= area.Y {
		t.Errorf("a caixa encostou na borda: %+v", region)
	}
	if region.X+region.Width > area.Width {
		t.Errorf("a caixa passou da área: %+v", region)
	}
}

// TestRegionHandlesASmallArea: zero and tiny sizes are valid, and an overlay must
// not produce a negative region.
func TestRegionHandlesASmallArea(t *testing.T) {
	for _, area := range []layout.Region{
		{}, {Width: 10, Height: 3}, {Width: 5, Height: 1},
	} {
		region := Region(area, 20, 10)
		if region.X < 0 || region.Y < 0 || region.Width < 0 || region.Height < 0 {
			t.Errorf("área %+v produziu região negativa: %+v", area, region)
		}
	}
}

func TestZeroSizeRendersNothing(t *testing.T) {
	if rows := List(0, 10, sample()); rows != nil {
		t.Errorf("largura zero devolveu %d linhas", len(rows))
	}
	if rows := List(60, 0, sample()); rows != nil {
		t.Errorf("altura zero devolveu %d linhas", len(rows))
	}
}
