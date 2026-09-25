package theme

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/ori-team/oride/internal/config"
	"github.com/ori-team/oride/internal/tui/layout"
)

func sampleConfig() (config.UIConfig, config.SyntaxColors) {
	return config.UIConfig{
			Background: "#282a36", CursorBG: "#44475a", CursorFG: "#f8f8f2",
			LineNumber: "#6272a4", StatusBG: "#282a36", StatusFG: "#f8f8f2",
			StatusDirty: "#ff5555",
		}, config.SyntaxColors{
			Comment: "#6272a4",
		}
}

// TestNoColorAppliesNothing is the whole no-colour guarantee, and it is asserted
// over every role rather than one: a theme that suppresses colour in one place and
// paints in another is the error the rule names.
func TestNoColorAppliesNothing(t *testing.T) {
	ui, syntax := sampleConfig()
	plain := New(ui, syntax, NoColor)

	roles := map[string]string{
		"selection": plain.Selection().Render("texto"),
		"gutter":    plain.Gutter().Render("12"),
		"status":    plain.Status().Render("status"),
		"dirty":     plain.StatusDirty().Render("sujo"),
		"comment":   plain.Comment().Render("// nota"),
	}
	for role, rendered := range roles {
		if strings.ContainsRune(rendered, 0x1b) {
			t.Errorf("%s: o perfil sem cor emitiu sequência de escape: %q", role, rendered)
		}
		if rendered != strings.TrimSpace(rendered) && strings.Contains(rendered, "\x1b") {
			t.Errorf("%s trouxe escape", role)
		}
	}
}

// TestNoColorRendersTheTextItself: suppressing colour must not suppress content.
func TestNoColorRendersTheTextItself(t *testing.T) {
	ui, syntax := sampleConfig()
	plain := New(ui, syntax, NoColor)

	if got := plain.Selection().Render("texto"); got != "texto" {
		t.Errorf("renderizou %q, esperado o texto sem alteração", got)
	}
}

// TestColorProfilesEmitSequences proves the other direction: a rule that only ever
// suppressed colour would pass the tests above while showing a monochrome editor.
func TestColorProfilesEmitSequences(t *testing.T) {
	for _, profile := range []Profile{ANSI16, ANSI256, TrueColor} {
		ui, syntax := sampleConfig()
		themed := New(ui, syntax, profile)

		if got := themed.Selection().Render("texto"); !strings.ContainsRune(got, 0x1b) {
			t.Errorf("%v: nenhuma sequência emitida: %q", profile, got)
		}
	}
}

// TestStyledTextMeasuresTheSameWidth is the invariant that keeps a styled frame
// aligned: escapes occupy no cells.
func TestStyledTextMeasuresTheSameWidth(t *testing.T) {
	ui, syntax := sampleConfig()
	themed := New(ui, syntax, TrueColor)

	for _, text := range []string{"texto", "日本語", "com ç e ã", ""} {
		plain := layout.Width(text)
		styled := layout.Width(themed.Selection().Render(text))

		if styled != plain {
			t.Errorf("%q: sem estilo %d células, com estilo %d", text, plain, styled)
		}
	}
}

// TestAStyledRowPadsToTheExactWidth is what a frame actually does, and where a
// width bug would show up as a diagonal layout.
func TestAStyledRowPadsToTheExactWidth(t *testing.T) {
	ui, syntax := sampleConfig()
	themed := New(ui, syntax, TrueColor)

	for _, width := range []int{10, 40, 79, 80} {
		row := layout.Pad(themed.Selection().Render("um rótulo"), width)
		if got := layout.Width(row); got != width {
			t.Errorf("largura %d: linha com %d células", width, got)
		}
	}
}

// TestEmptyAndResetMeanNoColour: both are normal in the shipped themes, and both
// are values the style library cannot resolve.
func TestEmptyAndResetMeanNoColour(t *testing.T) {
	for _, value := range []string{"", "reset", "none"} {
		if got := parse(value); got != nil {
			t.Errorf("parse(%q) = %v, esperado nada", value, got)
		}
	}
}

// TestParseAcceptsHexAndNames.
func TestParseAcceptsHexAndNames(t *testing.T) {
	for _, value := range []string{"#282a36", "#fff", "red", "brightblack"} {
		if parse(value) == nil {
			t.Errorf("parse(%q) não resolveu", value)
		}
	}
}

// TestNamedColoursActuallyRender is the regression guard for a bug that every
// other test in this file missed.
//
// They used hex, the shipped configuration uses names, and an unresolved name is a
// silent no-op — so the package passed its tests while the product rendered
// monochrome. This asserts the rendered output, not the parsed value.
func TestNamedColoursActuallyRender(t *testing.T) {
	for _, name := range []string{"darkgray", "red", "brightblack"} {
		rendered := lipgloss.NewStyle().Foreground(parse(name)).Render("x")
		if !strings.ContainsRune(rendered, 0x1b) {
			t.Errorf("a cor %q não produziu estilo: %q", name, rendered)
		}
	}
}

// TestShippedDefaultRendersWithColour uses the configuration the product ships,
// which is what the previous test could not cover.
func TestShippedDefaultRendersWithColour(t *testing.T) {
	shipped := config.Default()
	built := New(shipped.UI, shipped.Syntax, TrueColor)

	for role, rendered := range map[string]string{
		"status": built.Status().Render("x"),
		"gutter": built.Gutter().Render("1"),
	} {
		if !strings.ContainsRune(rendered, 0x1b) {
			t.Errorf("%s: a configuração padrão não produz cor: %q", role, rendered)
		}
	}
}

// TestProfileOrdering lets a caller ask "at least this capable" without a switch.
func TestProfileOrdering(t *testing.T) {
	if !(NoColor < ANSI16 && ANSI16 < ANSI256 && ANSI256 < TrueColor) {
		t.Error("os perfis não estão ordenados por capacidade")
	}
}
