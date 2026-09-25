package tui

import (
	"strconv"
	"strings"
	"testing"

	"github.com/ori-team/oride/internal/tui/layout"
	"github.com/ori-team/oride/internal/tui/menubar"
	"github.com/ori-team/oride/internal/tui/statusbar"
	"github.com/ori-team/oride/internal/tui/tabs"
	"github.com/ori-team/oride/internal/tui/theme"
	"github.com/ori-team/oride/internal/tui/tree"
)

// exactWidth is the invariant every one-line surface shares: a row that comes
// back short or long shifts everything beside it.
func assertExactWidth(t *testing.T, name string, row string, width int) {
	t.Helper()
	if got := layout.Width(row); got != width {
		t.Errorf("%s: %d células, esperado %d em %q", name, got, width, row)
	}
}

func TestTabBarHasTheExactWidth(t *testing.T) {
	views := []tabs.View{
		{},
		{Tabs: []tabs.Tab{{Title: "a.go", Active: true}}},
		{Tabs: []tabs.Tab{{Title: "a.go", Active: true}, {Title: "b.go", Dirty: true}}},
		{Tabs: []tabs.Tab{
			{Title: strings.Repeat("longo", 30)}, {Title: "curto", Active: true},
		}},
	}
	for index, view := range views {
		for _, width := range []int{10, 40, 80, 200} {
			assertExactWidth(t, "tabs "+strconv.Itoa(index), tabs.Render(width, view), width)
		}
	}
}

// TestTabBarMarksDirtyAndActiveInText: neither state may depend on colour.
func TestTabBarMarksDirtyAndActiveInText(t *testing.T) {
	row := tabs.Render(60, tabs.View{Tabs: []tabs.Tab{{Title: "a.go", Dirty: true, Active: true}}})

	if !strings.Contains(row, "*") {
		t.Errorf("o estado sujo não está escrito: %q", row)
	}
	if !strings.Contains(row, "[a.go*]") {
		t.Errorf("a aba ativa não está marcada: %q", row)
	}
}

// TestTabBarKeepsTheActiveTabWhenTheyDoNotFit: hiding which document is being
// edited is worse than hiding the others.
func TestTabBarKeepsTheActiveTabWhenTheyDoNotFit(t *testing.T) {
	view := tabs.View{Tabs: []tabs.Tab{
		{Title: strings.Repeat("x", 40)},
		{Title: "ativo.go", Active: true},
	}}

	if row := tabs.Render(20, view); !strings.Contains(row, "ativo") {
		t.Errorf("a aba ativa sumiu: %q", row)
	}
}

func TestStatusBarHasTheExactWidth(t *testing.T) {
	views := []statusbar.View{
		{},
		{File: "a.go", Dirty: true, Focus: "editor", CaretLine: 12, CaretCol: 4, HasCaret: true},
		{File: strings.Repeat("pasta/", 20) + "a.go", Message: strings.Repeat("mensagem ", 10), HasCaret: true},
	}
	for index, view := range views {
		for _, width := range []int{10, 40, 80, 200} {
			assertExactWidth(t, "status "+strconv.Itoa(index), statusbar.Render(width, view), width)
		}
	}
}

// TestStatusBarKeepsTheCaretPosition: the position is the field read by location,
// so it must survive truncation of everything else.
func TestStatusBarKeepsTheCaretPosition(t *testing.T) {
	view := statusbar.View{
		File:    strings.Repeat("muito/longo/", 20) + "arquivo.go",
		Message: strings.Repeat("aviso ", 20), Focus: "editor",
		CaretLine: 7, CaretCol: 3, HasCaret: true,
	}

	row := statusbar.Render(60, view)
	if !strings.Contains(row, "Ln 7, Col 3") {
		t.Errorf("a posição do cursor não sobreviveu: %q", row)
	}
}

// TestStatusBarMarksDirtyInText.
func TestStatusBarMarksDirtyInText(t *testing.T) {
	dirty := statusbar.Render(40, statusbar.View{File: "a.go", Dirty: true})
	clean := statusbar.Render(40, statusbar.View{File: "a.go"})

	if !strings.Contains(dirty, "*") || strings.Contains(clean, "*") {
		t.Errorf("sujo=%q limpo=%q", dirty, clean)
	}
}

func TestMenuBarHasTheExactWidth(t *testing.T) {
	views := []menubar.View{
		{},
		{Labels: []string{"Arquivo", "Editar", "Exibir", "Ir", "Git", "Ajuda"}, Open: -1},
		{Labels: []string{"Arquivo", "Editar"}, Open: 1},
	}
	for index, view := range views {
		for _, width := range []int{5, 20, 80, 200} {
			assertExactWidth(t, "menu "+strconv.Itoa(index), menubar.Render(width, view), width)
		}
	}
}

// TestMenuBarMarksTheOpenMenu.
func TestMenuBarMarksTheOpenMenu(t *testing.T) {
	row := menubar.Render(60, menubar.View{Labels: []string{"Arquivo", "Editar"}, Open: 1})

	if !strings.Contains(row, "[Editar]") {
		t.Errorf("o menu aberto não está marcado: %q", row)
	}
	if strings.Contains(row, "[Arquivo]") {
		t.Errorf("o menu fechado foi marcado: %q", row)
	}
}

func TestTreeHasTheExactWidthAndHeight(t *testing.T) {
	view := tree.View{
		Rows:     []tree.Row{{Name: "src", IsDir: true, Expanded: true}, {Name: "a.go", Depth: 1}},
		Selected: 1,
	}
	for _, width := range []int{10, 30, 80} {
		rows := tree.Render(width, 5, view)
		if len(rows) != 5 {
			t.Fatalf("largura %d: %d linhas", width, len(rows))
		}
		for index, row := range rows {
			assertExactWidth(t, "tree "+strconv.Itoa(index), row, width)
		}
	}
}

// TestTreeMarksDirectoriesInText: the affordance is what tells a reader the entry
// can be opened, so it cannot be colour or an ambiguous-width glyph.
func TestTreeMarksDirectoriesInText(t *testing.T) {
	view := tree.View{Rows: []tree.Row{
		{Name: "aberto", IsDir: true, Expanded: true},
		{Name: "fechado", IsDir: true},
		{Name: "arquivo.go"},
	}}

	rows := tree.Render(30, 3, view)
	if !strings.Contains(rows[0], "v ") {
		t.Errorf("diretório aberto sem marcador: %q", rows[0])
	}
	if !strings.Contains(rows[1], "> ") {
		t.Errorf("diretório fechado sem marcador: %q", rows[1])
	}
	if strings.Contains(rows[2], "v ") || strings.Contains(rows[2], "> ") {
		t.Errorf("arquivo marcado como diretório: %q", rows[2])
	}
}

// TestTreeSaysEmptyInsteadOfBlank: a blank rectangle does not distinguish an
// empty panel from one that failed to load.
func TestTreeSaysEmptyInsteadOfBlank(t *testing.T) {
	rows := tree.Render(30, 3, tree.View{})

	if !strings.Contains(rows[0], "(vazio)") {
		t.Errorf("painel vazio sem aviso: %q", rows[0])
	}
}

// TestNoColourFramesCarryNoEscapeSequences is the no-colour guarantee, asserted
// over every case rather than one.
//
// The profile is injected rather than read from the environment, so the test
// cannot leak into another one through shared state.
func TestNoColourFramesCarryNoEscapeSequences(t *testing.T) {
	for _, testCase := range frameCases() {
		model := sized(t, newModel(t, nil), testCase.width, testCase.height).WithProfile(theme.NoColor)
		if testCase.arrange != nil {
			testCase.arrange(&model)
		}

		if frame := model.Frame(); strings.ContainsRune(frame, 0x1b) {
			t.Errorf("%s: o perfil sem cor emitiu escape", testCase.name)
		}
	}
}

// TestColourFramesDoEmitEscapeSequences is the other direction, and it is not
// redundant: a rule that only ever suppressed colour would pass the test above
// while shipping a monochrome editor.
func TestColourFramesDoEmitEscapeSequences(t *testing.T) {
	for _, profile := range []theme.Profile{theme.ANSI256, theme.TrueColor} {
		model := sized(t, newModel(t, nil), 100, 30).WithProfile(profile)
		if !strings.ContainsRune(model.Frame(), 0x1b) {
			t.Errorf("%v: nenhum estilo foi aplicado ao frame", profile)
		}
	}
}

// TestStyledFramesKeepTheExactSize: escapes occupy no cells, so a coloured frame
// must measure the same as a plain one. This is where a width bug in the theme
// would show, and it would show as a diagonal layout.
func TestStyledFramesKeepTheExactSize(t *testing.T) {
	for _, profile := range []theme.Profile{theme.NoColor, theme.ANSI256, theme.TrueColor} {
		model := sized(t, newModel(t, nil), 100, 30).WithProfile(profile)
		model.application.Store.OpenEmpty()

		for index, row := range splitFrame(model.Frame()) {
			if got := layout.Width(row); got != 100 {
				t.Errorf("%v: linha %d com %d células", profile, index, got)
			}
		}
	}
}
