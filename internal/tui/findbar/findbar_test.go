package findbar

import (
	"strings"
	"testing"

	"github.com/ori-team/oride/internal/config"
	"github.com/ori-team/oride/internal/tui/layout"
	"github.com/ori-team/oride/internal/tui/theme"
)

func sample() View {
	return View{Query: "alfa", IgnoreAccents: true, Current: 2, Total: 5}
}

// TestEveryRowHasTheExactWidth keeps the bar from shifting the frame.
func TestEveryRowHasTheExactWidth(t *testing.T) {
	for _, showReplace := range []bool{false, true} {
		view := sample()
		view.ShowReplace = showReplace

		for _, width := range []int{20, 40, 79, 80, 200} {
			rows := Render(width, view)
			if len(rows) != view.Height() {
				t.Errorf("revelar=%v: %d linhas, esperado %d", showReplace, len(rows), view.Height())
			}
			for index, row := range rows {
				if got := layout.Width(row); got != width {
					t.Errorf("revelar=%v largura=%d: linha %d com %d células", showReplace, width, index, got)
				}
			}
		}
	}
}

// TestHeightGrowsOnlyForTheReplaceField: the layout is told before it divides the
// screen, so the height has to be right before anything is drawn.
func TestHeightGrowsOnlyForTheReplaceField(t *testing.T) {
	view := sample()
	if got := view.Height(); got != 1 {
		t.Errorf("altura = %d, esperado 1", got)
	}
	view.ShowReplace = true
	if got := view.Height(); got != 2 {
		t.Errorf("altura = %d, esperado 2", got)
	}
}

// TestFlagsAreSpelledOut: a flag carried only by colour is a flag the reader
// misreads a result over.
func TestFlagsAreSpelledOut(t *testing.T) {
	view := sample()
	view.CaseSensitive = true
	view.UseRegex = true

	row := Render(120, view)[0]
	for _, want := range []string{"Aa", "acentos", "regex"} {
		if !strings.Contains(row, want) {
			t.Errorf("o modificador %q não está escrito: %q", want, row)
		}
	}
}

// TestNoMatchSaysSo: a bar showing "0/0" leaves the reader unsure whether the
// search ran.
func TestNoMatchSaysSo(t *testing.T) {
	view := sample()
	view.Total = 0

	if row := Render(120, view)[0]; !strings.Contains(row, noMatch) {
		t.Errorf("sem casamentos sem aviso: %q", row)
	}
}

// TestARegularExpressionErrorReplacesTheCount: a pattern that does not compile
// matched nothing, so a count beside the error would be a useful lie.
func TestARegularExpressionErrorReplacesTheCount(t *testing.T) {
	view := sample()
	view.RegexError = "error parsing regexp: missing closing ]"
	view.Total = 0

	row := Render(200, view)[0]
	if !strings.Contains(row, "padrão inválido") {
		t.Errorf("o erro não apareceu: %q", row)
	}
	if strings.Contains(row, noMatch) {
		t.Errorf("a contagem apareceu junto do erro: %q", row)
	}
}

// TestALongQueryShowsItsTail: someone typing needs to see where they are typing.
func TestALongQueryShowsItsTail(t *testing.T) {
	view := sample()
	view.Query = strings.Repeat("x", 40) + "FIM"

	row := Render(120, view)[0]
	if !strings.Contains(row, "FIM") {
		t.Errorf("o fim da consulta não aparece: %q", row)
	}
}

// TestAnEmptyQueryShowsAPlaceholder: an empty field is indistinguishable from a
// field that lost its content.
func TestAnEmptyQueryShowsAPlaceholder(t *testing.T) {
	view := sample()
	view.Query = ""

	if row := Render(120, view)[0]; !strings.Contains(row, "digite…") {
		t.Errorf("campo vazio sem indicação: %q", row)
	}
}

// TestNoColourRendersNothingExtra is the bar's share of the no-colour guarantee.
func TestNoColourRendersNothingExtra(t *testing.T) {
	view := sample()
	view.Theme = plain()

	for _, row := range Render(80, view) {
		if strings.ContainsRune(row, 0x1b) {
			t.Errorf("o perfil sem cor emitiu escape: %q", row)
		}
	}
}

func TestZeroWidthRendersNothing(t *testing.T) {
	if rows := Render(0, sample()); rows != nil {
		t.Errorf("largura zero devolveu %d linhas", len(rows))
	}
}

// plain is the theme with colour suppressed, built from the product's own
// configuration so the test cannot drift from the values the bar receives.
func plain() theme.Theme {
	shipped := config.Default()
	return theme.New(shipped.UI, shipped.Syntax, theme.NoColor)
}
