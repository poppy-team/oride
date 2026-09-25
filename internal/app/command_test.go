package app

import (
	"errors"
	"testing"

	"github.com/ori-team/oride/internal/action"
	"github.com/ori-team/oride/internal/config"
	"github.com/ori-team/oride/internal/editor"
	"github.com/ori-team/oride/internal/keymap"
)

// dispatchedBefore is every action the old switch handled.
//
// A table can lose an entry without failing anything: the lookup simply stops
// finding it and the action returns ErrNotImplemented, which looks exactly like
// an action that was never implemented. This list is the regression guard for
// the refactor that introduced the registry.
var dispatchedBefore = []action.Action{
	action.Quit, action.Save, action.SaveAs, action.SaveAll,
	action.Undo, action.Redo,
	action.InsertNewline, action.InsertTab, action.Backspace, action.Delete,
	action.MoveLeftPlain, action.MoveLeftExtend,
	action.MoveRightPlain, action.MoveRightExtend,
	action.MoveUpPlain, action.MoveUpExtend,
	action.MoveDownPlain, action.MoveDownExtend,
	action.MoveLineStartPlain, action.MoveLineStartExtend,
	action.MoveLineEndPlain, action.MoveLineEndExtend,
	action.MoveDocStartPlain, action.MoveDocStartExtend,
	action.MoveDocEndPlain, action.MoveDocEndExtend,
	action.SelectAll,
	action.AddCursorAbove, action.AddCursorBelow, action.ClearExtraCursors,
	action.ToggleTree, action.FocusTree, action.FocusEditor,
}

func TestRegistryKeepsEveryActionThatWasDispatched(t *testing.T) {
	for _, id := range dispatchedBefore {
		if _, present := commands[id]; !present {
			t.Errorf("%s perdeu o registro", id)
		}
	}
	if got := CommandCount(); got != len(dispatchedBefore) {
		t.Errorf("a tabela tem %d ações e a lista de referência tem %d", got, len(dispatchedBefore))
	}
}

// TestRegistryKeysAreCanonical catches a typo in a registration: an id that is
// not in the canonical table would register a command nothing can ever invoke.
func TestRegistryKeysAreCanonical(t *testing.T) {
	for id := range commands {
		parsed, err := action.Parse(id.String())
		if err != nil {
			t.Errorf("id registrado não é canônico: %q: %v", id, err)
			continue
		}
		if parsed != id {
			t.Errorf("id %q volta como %q", id, parsed)
		}
	}
}

// TestNeedsDocumentIsHonoured is the behavioural half of the flag.
//
// It asserts the two directions: an action declared as needing a document fails
// with an empty store, and one declared as not needing it still runs — because a
// handler that quietly required a document would make quitting with nothing open
// impossible.
func TestNeedsDocumentIsHonoured(t *testing.T) {
	empty := New(editor.NewStore(), config.Default(), keymap.New())

	if err := empty.Apply(action.ToggleTree); err != nil {
		t.Errorf("ToggleTree sem documento aberto: %v", err)
	}
	if err := empty.Apply(action.Quit); err != nil {
		t.Errorf("Quit sem documento aberto: %v", err)
	}

	for _, id := range []action.Action{action.Undo, action.MoveLeftPlain, action.SelectAll} {
		if err := empty.Apply(id); err == nil {
			t.Errorf("%s deveria exigir um documento aberto", id)
		}
	}
}

// TestUnknownActionStillReportsNotImplemented keeps the sentinel in place: the
// harness and the palette both distinguish "not implemented yet" from a real
// failure.
func TestUnknownActionStillReportsNotImplemented(t *testing.T) {
	model := New(editor.NewStore(), config.Default(), keymap.New())

	err := model.Apply(action.HealthCheck)
	if err == nil {
		t.Fatal("uma ação fora da tabela foi aceita")
	}
	if !errors.Is(err, ErrNotImplemented) {
		t.Errorf("erro = %v, esperado ErrNotImplemented", err)
	}
}
