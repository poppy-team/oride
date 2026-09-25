package layout

import (
	"strings"
	"testing"
)

// TestClassifyBoundaries pins the edges. An off-by-one here moves every surface
// on screen, and the boundaries are the only part of the classification that can
// be wrong.
func TestClassifyBoundaries(t *testing.T) {
	cases := map[int]WidthClass{
		0: Compact, 1: Compact, 79: Compact,
		80: Standard, 119: Standard,
		120: Wide, 200: Wide,
	}
	for width, want := range cases {
		if got := Classify(width); got != want {
			t.Errorf("Classify(%d) = %v, esperado %v", width, got, want)
		}
	}
}

// TestComputeStacksEverySurface covers the vertical order the layout document
// promises: menu bar, tabs, body, status bar — top to bottom, no gaps.
func TestComputeStacksEverySurface(t *testing.T) {
	regions := Compute(Size{Width: 100, Height: 30}, Wants{})

	if regions.MenuBar.Y != 0 {
		t.Errorf("menu bar em y=%d, esperado 0", regions.MenuBar.Y)
	}
	if regions.Tabs.Y != regions.MenuBar.Y+regions.MenuBar.Height {
		t.Errorf("tabs em y=%d, logo abaixo do menu bar", regions.Tabs.Y)
	}
	if regions.Editor.Y != regions.Tabs.Y+regions.Tabs.Height {
		t.Errorf("editor em y=%d, logo abaixo das abas", regions.Editor.Y)
	}
	if regions.StatusBar.Y != 29 {
		t.Errorf("statusbar em y=%d, esperado 29 (última linha)", regions.StatusBar.Y)
	}

	// Sem sobreposição e sem buraco: a soma das alturas cobre o terminal.
	covered := regions.MenuBar.Height + regions.Tabs.Height + regions.Editor.Height + regions.StatusBar.Height
	if covered != 30 {
		t.Errorf("alturas somam %d, esperado 30", covered)
	}
}

// TestComputeNeverOverlapsHorizontally: a tree drawn on top of the editor's first
// columns would hide text with no sign that it is there.
func TestComputeNeverOverlapsHorizontally(t *testing.T) {
	for _, width := range []int{60, 80, 100, 120, 160} {
		regions := Compute(Size{Width: width, Height: 30}, Wants{ShowTree: true, TreeWidth: 30})

		if regions.TreeOverlaid {
			continue // sobreposição é o comportamento declarado em compact
		}
		if regions.Tree.X+regions.Tree.Width != regions.Editor.X {
			t.Errorf("largura %d: árvore termina em %d e editor começa em %d",
				width, regions.Tree.X+regions.Tree.Width, regions.Editor.X)
		}
	}
}

// TestComputeOverlaysTheTreeWhenCompact is the rule of the layout document: in
// compact the tree floats instead of taking a column, so the editor keeps its
// width.
func TestComputeOverlaysTheTreeWhenCompact(t *testing.T) {
	regions := Compute(Size{Width: 60, Height: 30}, Wants{ShowTree: true, TreeWidth: 30})

	if !regions.TreeOverlaid {
		t.Fatal("em compact a árvore deveria estar sobreposta")
	}
	if regions.Editor.Width != 60 {
		t.Errorf("o editor ficou com %d colunas, esperado 60 (a largura toda)", regions.Editor.Width)
	}
	if regions.Tree.Width > regions.Editor.Width {
		t.Error("a árvore flutuante ficou maior que o editor")
	}
}

// TestComputeProtectsTheEditor: a tree that squeezes the editor below the minimum
// has stopped being useful, so the tree gives way.
func TestComputeProtectsTheEditor(t *testing.T) {
	regions := Compute(Size{Width: 90, Height: 30}, Wants{ShowTree: true, TreeWidth: 70})

	if !regions.TreeOverlaid {
		t.Errorf("o editor ficou com %d colunas; a árvore deveria ter cedido", regions.Editor.Width)
	}
	if regions.Editor.Width < 40 {
		t.Errorf("o editor ficou com %d colunas, abaixo do mínimo", regions.Editor.Width)
	}
}

// TestComputeTerminalTakesFromTheBody: the terminal is carved out of the body and
// the editor keeps what is left, with the status bar untouched.
func TestComputeTerminalTakesFromTheBody(t *testing.T) {
	regions := Compute(Size{Width: 100, Height: 40}, Wants{ShowTerminal: true, TerminalHeight: 10})

	if regions.Terminal.Height != 10 {
		t.Fatalf("terminal com %d linhas, esperado 10", regions.Terminal.Height)
	}
	if regions.Terminal.Y+regions.Terminal.Height != regions.StatusBar.Y {
		t.Errorf("o terminal não termina onde a statusbar começa: %d+%d vs %d",
			regions.Terminal.Y, regions.Terminal.Height, regions.StatusBar.Y)
	}
	if regions.Editor.Y+regions.Editor.Height != regions.Terminal.Y {
		t.Errorf("o editor deveria terminar onde o terminal começa")
	}
}

// TestComputeRefusesToEatTheEditor: on a short terminal a tall terminal panel
// would leave nothing to edit, so it is dropped rather than shown empty.
func TestComputeRefusesToEatTheEditor(t *testing.T) {
	regions := Compute(Size{Width: 100, Height: 12}, Wants{ShowTerminal: true, TerminalHeight: 20})

	if !regions.Terminal.Empty() {
		t.Errorf("o terminal tomou %d linhas de um corpo de %d", regions.Terminal.Height, regions.Editor.Height)
	}
	if regions.Editor.Empty() {
		t.Error("o editor ficou sem área")
	}
}

// TestComputeHandlesAnUnmeasuredTerminal: zero is a valid size. A terminal that
// has not been measured yet must produce empty regions, not a panic or a negative
// width that renders as garbage.
func TestComputeHandlesAnUnmeasuredTerminal(t *testing.T) {
	for _, size := range []Size{{}, {Width: 0, Height: 0}, {Width: -5, Height: 10}, {Width: 80, Height: -1}, {Width: 3, Height: 3}} {
		regions := Compute(size, Wants{ShowTree: true, TreeWidth: 30, ShowTerminal: true, TerminalHeight: 10})

		for name, region := range map[string]Region{
			"editor": regions.Editor, "tree": regions.Tree,
			"terminal": regions.Terminal, "statusbar": regions.StatusBar,
		} {
			if region.Width < 0 || region.Height < 0 {
				t.Errorf("%v: %s com dimensão negativa: %+v", size, name, region)
			}
		}
	}
}

// TestTruncateMarksTheCut: a label that vanishes with no sign is
// indistinguishable from one that never existed.
//
// The expected strings encode the marker's real width — two cells, because an
// ambiguous-width character counts as wide here. An earlier version of this test
// assumed one cell and agreed with a constant that was wrong.
func TestTruncateMarksTheCut(t *testing.T) {
	if got := Truncate("abcdef", 4); got != "ab…" {
		t.Errorf("Truncate = %q, esperado \"ab…\" (o marcador ocupa 2 células)", got)
	}
	if got := Truncate("abc", 10); got != "abc" {
		t.Errorf("texto que cabe foi alterado: %q", got)
	}
	// Em uma coluna o marcador não cabe; um ponto cabe.
	if got := Truncate("abcdef", 1); got != "." {
		t.Errorf("Truncate com uma coluna = %q, esperado \".\"", got)
	}
	if got := Truncate("abcdef", 0); got != "" {
		t.Errorf("Truncate com zero = %q", got)
	}
}

// TestTruncateNeverOverflows is the invariant behind the three bugs this had:
// every result must fit the width it was given, for every width.
func TestTruncateNeverOverflows(t *testing.T) {
	texts := []string{"abcdef", "日本語のテキスト", "uma frase longa", "→→→", "a", ""}
	for _, text := range texts {
		for width := 0; width <= 20; width++ {
			if got := Width(Truncate(text, width)); got > width {
				t.Errorf("Truncate(%q, %d) = %q com %d células", text, width, Truncate(text, width), got)
			}
		}
	}
}

// TestTruncateCountsCellsNotRunes: a wide character occupies two columns, and
// counting runes overflows the row.
func TestTruncateCountsCellsNotRunes(t *testing.T) {
	// 日本語 = 3 runas, 6 células.
	if got := Width("日本語"); got != 6 {
		t.Fatalf("Width(日本語) = %d, esperado 6", got)
	}
	if got := Truncate("日本語", 5); Width(got) > 5 {
		t.Errorf("Truncate = %q com %d células, acima de 5", got, Width(got))
	}
}

// TestWidthIsConservative treats ambiguous-width characters as wide. Erring that
// way costs a column; erring the other way misaligns the whole row.
func TestWidthIsConservative(t *testing.T) {
	if got := Width("→"); got != 2 {
		t.Errorf("Width(→) = %d; a largura ambígua deveria contar como 2", got)
	}
}

// TestPadFitsExactly: a row that comes back short lets whatever is to its right
// shift left, which turns a two-column layout into a diagonal one.
func TestPadFitsExactly(t *testing.T) {
	cases := []struct {
		text  string
		width int
	}{
		{"abc", 10}, {"", 5}, {"exatamente", 10}, {"uma frase bem mais longa", 12},
		{"日本語", 8}, {"日本語", 4},
	}
	for _, tc := range cases {
		got := Pad(tc.text, tc.width)
		if Width(got) != tc.width {
			t.Errorf("Pad(%q, %d) = %q com %d células", tc.text, tc.width, got, Width(got))
		}
	}
}

// TestPadWithNonPositiveWidth returns nothing rather than padding to a negative
// width.
func TestPadWithNonPositiveWidth(t *testing.T) {
	for _, width := range []int{0, -1} {
		if got := Pad("abc", width); got != "" {
			t.Errorf("Pad(abc, %d) = %q", width, got)
		}
	}
}

// TestEveryWidthProducesASaneFrame walks the widths the gate names, plus the
// boundaries, and asserts the frame is always usable.
func TestEveryWidthProducesASaneFrame(t *testing.T) {
	for _, width := range []int{20, 60, 79, 80, 100, 119, 120, 140, 200} {
		regions := Compute(Size{Width: width, Height: 40}, Wants{ShowTree: true, TreeWidth: 30})

		if regions.Editor.Empty() {
			t.Errorf("largura %d: o editor ficou vazio", width)
			continue
		}
		if regions.MenuBar.Width != width || regions.StatusBar.Width != width {
			t.Errorf("largura %d: as barras não cobrem o terminal", width)
		}
		if regions.Editor.X+regions.Editor.Width > width {
			t.Errorf("largura %d: o editor passa de %d", regions.Editor.X+regions.Editor.Width, width)
		}
	}
}

// TestOverlaidTreeNeverExceedsTheBody keeps the floating panel inside its parent.
func TestOverlaidTreeNeverExceedsTheBody(t *testing.T) {
	for _, height := range []int{3, 10, 30, 80} {
		regions := Compute(Size{Width: 60, Height: height}, Wants{ShowTree: true, TreeWidth: 40})

		if regions.Tree.Y+regions.Tree.Height > height {
			t.Errorf("altura %d: a árvore flutuante passa do terminal", height)
		}
		if regions.Tree.Width > 60 {
			t.Errorf("altura %d: a árvore flutuante passa da largura", height)
		}
	}
}

// TestNoSurfaceIsEverNegative is the cheap invariant that catches an arithmetic
// mistake anywhere in Compute.
func TestNoSurfaceIsEverNegative(t *testing.T) {
	for width := 0; width <= 200; width += 7 {
		for height := 0; height <= 60; height += 5 {
			regions := Compute(Size{Width: width, Height: height},
				Wants{ShowTree: true, TreeWidth: 30, ShowTerminal: true, TerminalHeight: 12})

			all := []Region{regions.MenuBar, regions.Tabs, regions.Tree,
				regions.Editor, regions.Terminal, regions.StatusBar}
			for index, region := range all {
				if region.Width < 0 || region.Height < 0 || region.X < 0 || region.Y < 0 {
					t.Fatalf("%dx%d: região %d negativa: %+v", width, height, index, region)
				}
			}
		}
	}
}

// TestTruncateIsStableUnderRepetition: truncating an already-truncated string must
// not keep eating it. A surface that re-truncates its own output every frame would
// otherwise shrink the text one marker at a time.
func TestTruncateIsStableUnderRepetition(t *testing.T) {
	text := "uma linha bem longa que não cabe"
	once := Truncate(text, 12)
	twice := Truncate(once, 12)

	if once != twice {
		t.Errorf("truncar de novo mudou o texto: %q → %q", once, twice)
	}
	if strings.HasSuffix(once, "……") {
		t.Errorf("o marcador se acumulou: %q", once)
	}
}
