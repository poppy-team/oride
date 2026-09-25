package tui

import (
	"strings"
	"testing"

	"github.com/ori-team/oride/internal/tui/layout"
	"github.com/ori-team/oride/internal/tui/theme"
)

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
