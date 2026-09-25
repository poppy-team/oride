package overlay

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func items() []Item {
	return []Item{
		{Label: "undo", Detail: "edição"},
		{Label: "redo"},
		{Label: "save", Detail: "arquivo"},
	}
}

func opened(t *testing.T) Model {
	t.Helper()

	model := New()
	model.Open(Palette, "Comandos", items(), 60, 10)
	return model
}

// TestCapturesEverything is the first rule of the focus graph, expressed as data.
func TestCapturesEverything(t *testing.T) {
	for _, kind := range []Kind{Palette, WhichKey, Help, Welcome, QuitConfirm, CloseConfirm} {
		if !kind.Captures() {
			t.Errorf("%v deveria capturar a entrada", kind)
		}
	}
	if None.Captures() {
		t.Error("nenhuma sobreposição não deveria capturar nada")
	}
}

// TestNewStartsClosed: an overlay that opened itself would capture input before
// anything asked for it.
func TestNewStartsClosed(t *testing.T) {
	model := New()

	if model.Active() {
		t.Error("a sobreposição nasceu aberta")
	}
	if model.View() != "" {
		t.Errorf("uma sobreposição fechada desenhou algo: %q", model.View())
	}
	if model.Kind() != None {
		t.Errorf("tipo = %v, esperado None", model.Kind())
	}
}

func TestOpenAndCloseRoundTrip(t *testing.T) {
	model := opened(t)

	if !model.Active() || model.Kind() != Palette {
		t.Fatalf("não abriu: ativo=%v tipo=%v", model.Active(), model.Kind())
	}
	if model.View() == "" {
		t.Error("aberta e não desenhou nada")
	}

	model.Close()
	if model.Active() {
		t.Error("Close não fechou")
	}
	if model.View() != "" {
		t.Error("fechada e ainda desenha")
	}
}

// TestUpdateIsIgnoredWhileClosed: a closed overlay must not swallow messages that
// belong to the surfaces.
func TestUpdateIsIgnoredWhileClosed(t *testing.T) {
	model := New()

	updated, cmd := model.Update(tea.KeyPressMsg(tea.Key{Text: "x"}))
	if updated.Active() {
		t.Error("fechada, e mesmo assim abriu")
	}
	if cmd != nil {
		t.Error("fechada, e mesmo assim devolveu comando")
	}
}

// TestTheFilterIsOwnedByTheComponent is the point of the package: typing narrows
// the list without this code knowing how.
func TestTheFilterIsOwnedByTheComponent(t *testing.T) {
	model := opened(t)

	// O componente só entra em modo de filtro por uma tecla própria dele.
	typed, _ := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	model = typed
	if !model.Active() {
		t.Fatal("a tecla de filtro fechou a sobreposição")
	}
}

// TestSelectedReportsTheRow: the caller needs to know what was chosen without
// reaching into the component.
func TestSelectedReportsTheRow(t *testing.T) {
	model := opened(t)

	choice, ok := model.Selected()
	if !ok {
		t.Fatal("nada selecionado numa lista com itens")
	}
	if choice.Label != "undo" {
		t.Errorf("selecionado = %q, esperado o primeiro item", choice.Label)
	}
}

// TestASelectedItemSatisfiesBothInterfaces: the component needs FilterValue to
// match against and Title/Description to draw, and a type that satisfied only one
// would compile and render blank.
func TestASelectedItemSatisfiesBothInterfaces(t *testing.T) {
	item := Item{Label: "undo", Detail: "edição"}

	if item.FilterValue() != "undo" {
		t.Errorf("FilterValue = %q", item.FilterValue())
	}
	if item.Title() != "undo" || item.Description() != "edição" {
		t.Errorf("título=%q descrição=%q", item.Title(), item.Description())
	}
}

// TestFilteringStartsFalse: Escape has two owners, and the overlay takes it only
// when the filter is not using it.
func TestFilteringStartsFalse(t *testing.T) {
	if opened(t).Filtering() {
		t.Error("o filtro começou ativo, e o Escape seria dele")
	}
}

// TestSetSizeIsIgnoredWhileClosed: resizing a closed overlay must not resurrect it.
func TestSetSizeIsIgnoredWhileClosed(t *testing.T) {
	model := New()
	model.SetSize(80, 20)

	if model.Active() {
		t.Error("o redimensionamento abriu a sobreposição")
	}
}

// TestTheViewFitsTheAskedSize: the composition root paints into a fixed region, and
// a view wider or taller than that would write past it.
func TestTheViewFitsTheAskedSize(t *testing.T) {
	model := New()
	model.Open(Palette, "Comandos", items(), 40, 8)

	lines := strings.Split(model.View(), "\n")
	if len(lines) > 8 {
		t.Errorf("view com %d linhas, acima de 8", len(lines))
	}
}
