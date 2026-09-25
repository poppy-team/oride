package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ori-team/oride/internal/action"
	"github.com/ori-team/oride/internal/keymap"
	"github.com/ori-team/oride/internal/tui/layout"
	"github.com/ori-team/oride/internal/tui/overlay"
)

// chordFor finds the chord bound to an action, so a test does not hard-code a
// keybinding that a remap would break.
func chordFor(t *testing.T, keys *keymap.Map, id action.Action) string {
	t.Helper()

	for _, binding := range keys.Bindings() {
		if binding.Action == id {
			return binding.Chord.String()
		}
	}
	t.Fatalf("nenhum binding para %s", id)
	return ""
}

// key builds a keystroke the model understands.
func key(text string) tea.KeyPressMsg { return tea.KeyPressMsg(tea.Key{Text: text}) }

// pressChord opens whatever a chord resolves to.
//
// It calls the model's own resolution rather than reconstructing a keystroke from
// the chord: Bubble Tea builds the chord from the key's fields, and a test that
// rebuilt it would be testing that encoding rather than this code.
func pressChord(t *testing.T, model Model, chord string) Model {
	t.Helper()

	if !model.openOverlayFor(chord) {
		t.Fatalf("o chord %q não abriu nenhuma sobreposição", chord)
	}
	return model
}

// TestOverlayCapturesEveryKeystroke is the first rule of the focus graph, and the
// most important test in this file: a key typed into a filter that also edited the
// buffer would be data loss, not a nuisance.
func TestOverlayCapturesEveryKeystroke(t *testing.T) {
	model := sized(t, newModel(t, nil), 100, 30)
	model.application.Store.OpenEmpty()

	// Open the palette through the keymap, so the test exercises the real path.
	chord := chordFor(t, model.keys, action.CommandPalette)
	model = pressChord(t, model, chord)
	if !model.overlay.Active() {
		t.Fatalf("a palette não abriu com %q", chord)
	}

	// Type characters that would edit the buffer if the capture rule leaked.
	before := bufferText(t, model)
	for _, r := range "abc" {
		updated, _ := model.Update(key(string(r)))
		model = updated.(Model)
	}

	if got := bufferText(t, model); got != before {
		t.Errorf("o buffer mudou com a palette aberta: %q → %q", before, got)
	}
	if model.filter != "abc" {
		t.Errorf("filtro = %q, esperado \"abc\"", model.filter)
	}
}

// TestClosingTheOverlayReturnsInputToSurfaces is the other direction: the capture
// must not outlive the overlay.
func TestClosingTheOverlayReturnsInputToSurfaces(t *testing.T) {
	model := sized(t, newModel(t, nil), 100, 30)
	model.application.Store.OpenEmpty()
	model.overlay = overlay.Palette

	updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	model = updated.(Model)
	if model.overlay.Active() {
		t.Fatal("Esc não fechou a sobreposição")
	}

	updated, _ = model.Update(key("z"))
	model = updated.(Model)
	if got := bufferText(t, model); got != "z" {
		t.Errorf("depois de fechar, digitar não alcançou o buffer: %q", got)
	}
}

func bufferText(t *testing.T, model Model) string {
	t.Helper()

	document, err := model.application.Store.Active()
	if err != nil {
		return ""
	}
	return document.Buffer().String()
}

// TestTheFilterNarrowsTheList.
func TestTheFilterNarrowsTheList(t *testing.T) {
	model := sized(t, newModel(t, nil), 100, 30)
	model.overlay = overlay.Palette

	all := len(model.overlayView().Items)
	if all == 0 {
		t.Fatal("a palette não ofereceu nenhum comando")
	}

	model.filter = "undo"
	narrowed := len(model.overlayView().Items)
	if narrowed >= all {
		t.Errorf("o filtro não reduziu: %d de %d", narrowed, all)
	}
	if narrowed == 0 {
		t.Error("o filtro não achou o comando undo, que existe")
	}
}

// TestBackspaceRemovesARuneNotAByte: an accented letter takes two bytes, and a
// byte-wise delete would leave an invalid fragment behind.
func TestBackspaceRemovesARuneNotAByte(t *testing.T) {
	model := sized(t, newModel(t, nil), 100, 30)
	model.overlay = overlay.Palette
	model.filter = "aç"

	updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyBackspace}))
	model = updated.(Model)

	if model.filter != "a" {
		t.Errorf("filtro = %q, esperado \"a\"", model.filter)
	}
}

// TestSelectionIsClampedToTheFilteredList: moving down past the end would leave
// the highlight on a row that is not drawn.
func TestSelectionIsClampedToTheFilteredList(t *testing.T) {
	model := sized(t, newModel(t, nil), 100, 30)
	model.overlay = overlay.WhichKey

	total := len(model.overlayView().Items)
	for range total + 3 {
		updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
		model = updated.(Model)
	}
	if model.selected != total-1 {
		t.Errorf("seleção = %d, esperado %d (o fim da lista)", model.selected, total-1)
	}

	for range total + 3 {
		updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))
		model = updated.(Model)
	}
	if model.selected != 0 {
		t.Errorf("seleção = %d, esperado 0", model.selected)
	}
}

// TestOverlayFramesKeepTheExactSize: an overlay that changed the frame's
// dimensions would leave the previous frame's remnants around it.
func TestOverlayFramesKeepTheExactSize(t *testing.T) {
	for _, kind := range []overlay.Kind{overlay.Palette, overlay.WhichKey, overlay.Help} {
		for _, size := range []struct{ width, height int }{{60, 24}, {80, 30}, {140, 40}} {
			model := sized(t, newModel(t, nil), size.width, size.height)
			model.overlay = kind

			for index, row := range splitFrame(model.Frame()) {
				if got := layout.Width(row); got != size.width {
					t.Errorf("%v %dx%d: linha %d com %d células", kind, size.width, size.height, index, got)
				}
			}
		}
	}
}

// TestOverlayKeepsTheChromeVisible: the menu bar, tabs and status bar stay
// readable, because the reader keeps needing to know what is open and where the
// cursor is.
func TestOverlayKeepsTheChromeVisible(t *testing.T) {
	model := sized(t, newModel(t, nil), 100, 30)
	model.overlay = overlay.Palette

	rows := splitFrame(model.Frame())
	if len(rows) != 30 {
		t.Fatalf("linhas = %d, esperado 30", len(rows))
	}
	if !strings.Contains(rows[0], model.catalog.Menu.File) {
		t.Errorf("o menu bar sumiu: %q", rows[0])
	}
	if !strings.Contains(rows[29], string(model.surface)) {
		t.Errorf("a statusbar sumiu: %q", rows[29])
	}
}
