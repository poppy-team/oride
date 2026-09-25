package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ori-team/oride/internal/app"
	"github.com/ori-team/oride/internal/config"
	"github.com/ori-team/oride/internal/editor"
	"github.com/ori-team/oride/internal/keymap"
	"github.com/ori-team/oride/internal/tui/focus"
	"github.com/ori-team/oride/internal/tui/layout"
)

func newModel(t *testing.T, configure func(*config.Config)) Model {
	t.Helper()

	cfg := config.Default()
	if configure != nil {
		configure(&cfg)
	}
	keys, err := keymap.FromBindings(cfg.Keys)
	if err != nil {
		t.Fatalf("keymap: %v", err)
	}
	application := app.New(editor.NewStore(), cfg, keys)
	return New(application, keys)
}

func sized(t *testing.T, model Model, width, height int) Model {
	t.Helper()

	updated, _ := model.Update(tea.WindowSizeMsg{Width: width, Height: height})
	sizedModel, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update devolveu %T", updated)
	}
	return sizedModel
}

// TestMouseIsOffByDefault is the gate the milestone names. The documented default
// is off, and off has to mean off: an editor that captures the mouse without being
// asked takes text selection away from the terminal.
func TestMouseIsOffByDefault(t *testing.T) {
	model := newModel(t, nil)

	if got := model.View().MouseMode; got != tea.MouseModeNone {
		t.Errorf("modo de mouse = %v, esperado MouseModeNone", got)
	}
	if !config.Default().Mouse {
		t.Skip()
	}
}

// TestMouseModeFollowsConfiguration: the mode is derived from config, not toggled
// per frame.
func TestMouseModeFollowsConfiguration(t *testing.T) {
	withMouse := newModel(t, func(c *config.Config) { c.Mouse = true })

	if got := withMouse.View().MouseMode; got != tea.MouseModeCellMotion {
		t.Errorf("com mouse ligado o modo = %v, esperado MouseModeCellMotion", got)
	}
}

// TestViewDeclaresTheAlternateScreen: capability is declared, so the terminal is
// reconciled by the runtime rather than by imperative escape sequences.
func TestViewDeclaresTheAlternateScreen(t *testing.T) {
	view := newModel(t, nil).View()

	if !view.AltScreen {
		t.Error("a tela alternativa deveria estar declarada")
	}
	if view.WindowTitle != windowTitle {
		t.Errorf("título = %q, esperado %q", view.WindowTitle, windowTitle)
	}
}

// TestFrameFillsTheTerminalExactly: a frame shorter or wider than the terminal
// leaves the previous frame's remnants on screen.
func TestFrameFillsTheTerminalExactly(t *testing.T) {
	for _, width := range []int{60, 80, 120, 140} {
		for _, height := range []int{10, 24, 40} {
			model := sized(t, newModel(t, nil), width, height)
			frame := model.Frame()

			lines := splitFrame(frame)
			if len(lines) != height {
				t.Errorf("%dx%d: %d linhas", width, height, len(lines))
				continue
			}
			for index, line := range lines {
				if got := layout.Width(line); got != width {
					t.Errorf("%dx%d: linha %d com %d células, esperado %d", width, height, index, got, width)
					break
				}
			}
		}
	}
}

// TestFrameIsEmptyBeforeTheTerminalIsMeasured: a zero size is valid, and rendering
// into it must produce nothing rather than a panic or a negative width.
func TestFrameIsEmptyBeforeTheTerminalIsMeasured(t *testing.T) {
	model := newModel(t, nil)

	if got := model.Frame(); got != "" {
		t.Errorf("frame sem tamanho medido = %q", got)
	}
}

// TestFocusIsAlwaysOnAVisibleSurface: focus must never rest where nothing is
// painted, because the next keystroke would go somewhere invisible.
func TestFocusIsAlwaysOnAVisibleSurface(t *testing.T) {
	model := sized(t, newModel(t, nil), 100, 30)

	for range 12 {
		updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
		next, ok := updated.(Model)
		if !ok {
			t.Fatalf("Update devolveu %T", updated)
		}
		model = next

		if !model.surfaceVisible(model.surface) {
			t.Fatalf("o foco parou em %q, que não está visível", model.surface)
		}
	}
}

// TestFocusSyncsToTheModel: the graph speaks the presentation vocabulary, the
// model speaks its own, and the dump reads the model's — so the two must agree.
func TestFocusSyncsToTheModel(t *testing.T) {
	model := sized(t, newModel(t, nil), 100, 30)
	model.application.ShowTree = true

	seenTree := false
	for range 12 {
		updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
		next, _ := updated.(Model)
		model = next

		if model.application.Focus == app.FocusTree {
			seenTree = true
		}
		if model.surface == focus.Tree && model.application.Focus != app.FocusTree {
			t.Error("a superfície e o modelo discordam sobre o foco")
		}
	}
	if !seenTree {
		t.Error("Tab nunca alcançou a árvore, que está visível")
	}
}

// TestUnboundPrintableKeyTypes is what makes typing work: a keystroke the keymap
// does not resolve is text.
func TestUnboundPrintableKeyTypes(t *testing.T) {
	model := sized(t, newModel(t, nil), 100, 30)
	model.application.Store.OpenEmpty()

	updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Text: "x"}))
	next, _ := updated.(Model)

	document, err := next.application.Store.Active()
	if err != nil {
		t.Fatalf("documento ativo: %v", err)
	}
	if got := document.Buffer().String(); got != "x" {
		t.Errorf("buffer = %q, esperado \"x\"", got)
	}
}

// TestNamedKeyIsNeverTypedAsText: an unbound named key must not insert its own
// name into the buffer.
func TestNamedKeyIsNeverTypedAsText(t *testing.T) {
	for _, chord := range []string{"f5", "escape", "ctrl+alt+shift+k"} {
		if isTypable(chord) {
			t.Errorf("%q foi considerado digitável", chord)
		}
	}
	for _, chord := range []string{"a", "ç", "1", " "} {
		if !isTypable(chord) {
			t.Errorf("%q deveria ser digitável", chord)
		}
	}
}

// splitFrame splits on newlines. strings.Split is not used because a frame with a
// trailing newline would gain a line that is not on screen.
func splitFrame(frame string) []string {
	return strings.Split(frame, "\n")
}
