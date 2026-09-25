package app

import (
	"testing"

	"github.com/ori-team/oride/internal/action"
	"github.com/ori-team/oride/internal/config"
	"github.com/ori-team/oride/internal/editor"
	"github.com/ori-team/oride/internal/keymap"
	"github.com/ori-team/oride/internal/search"
)

func searchable(t *testing.T, text string) *App {
	t.Helper()

	application := New(editor.NewStore(), config.Default(), keymap.New())
	application.Store.OpenEmpty()

	document, err := application.Store.Active()
	if err != nil {
		t.Fatalf("documento: %v", err)
	}
	if err := document.InsertText(text); err != nil {
		t.Fatalf("inserindo: %v", err)
	}
	document.MarkSaved()

	application.Find.Query = "alfa"
	application.Find.Options = search.Options{IgnoreAccents: true}
	application.Find.Recompute(document.Buffer().String())
	return application
}

// TestFindNextSelectsTheMatch: the match has to be visible, and the selection is
// what a replacement acts on.
func TestFindNextSelectsTheMatch(t *testing.T) {
	application := searchable(t, "alfa beta\nbeta alfa\n")

	if err := application.Apply(action.FindNext); err != nil {
		t.Fatalf("find_next: %v", err)
	}

	document, _ := application.Store.Active()
	selected := document.SelectedText()
	if selected != "alfa" {
		t.Errorf("seleção = %q, esperado o casamento", selected)
	}
}

// TestFindMovesForwardThenBackward.
func TestFindMovesForwardThenBackward(t *testing.T) {
	application := searchable(t, "alfa um\nalfa dois\nalfa três\n")
	document, _ := application.Store.Active()

	_ = application.Apply(action.FindNext)
	first := document.Selection().Start()
	if err := application.Apply(action.FindNext); err != nil {
		t.Fatalf("segundo find_next: %v", err)
	}
	if document.Selection().Start() <= first {
		t.Errorf("find_next não avançou: %d depois de %d", document.Selection().Start(), first)
	}

	before := document.Selection().Start()
	if err := application.Apply(action.FindPrev); err != nil {
		t.Fatalf("find_prev: %v", err)
	}
	if document.Selection().Start() >= before {
		t.Errorf("find_prev não recuou: %d depois de %d", document.Selection().Start(), before)
	}
}

// TestFindOpensTheOverlayWithoutTheReplaceField.
func TestFindOpensTheOverlayWithoutTheReplaceField(t *testing.T) {
	application := searchable(t, "alfa\n")

	if err := application.Apply(action.Find); err != nil {
		t.Fatalf("find: %v", err)
	}
	if application.Overlay != "find" {
		t.Errorf("overlay = %q, esperado \"find\"", application.Overlay)
	}
	if application.Find.ShowReplace {
		t.Error("Find não deveria revelar o campo de substituição")
	}
}

// TestReplaceRevealsTheFieldAndDoesNotSubstitute pins a behaviour the oracle
// corrected.
//
// An earlier version performed the substitution in this action, and the
// conformance case showed the Rust leaving the document untouched with the
// selection still on the match. Substituting is a separate flow, and a test is
// worth more than a comment here: this is exactly the kind of thing a reader
// would "fix" back into a bug.
func TestReplaceRevealsTheFieldAndDoesNotSubstitute(t *testing.T) {
	application := searchable(t, "alfa beta\n")
	document, _ := application.Store.Active()
	before := document.Buffer().String()

	_ = application.Apply(action.FindNext)
	application.Find.Replace = "ALFA"
	if err := application.Apply(action.Replace); err != nil {
		t.Fatalf("replace: %v", err)
	}

	if !application.Find.ShowReplace {
		t.Error("Replace deveria revelar o campo de substituição")
	}
	if got := document.Buffer().String(); got != before {
		t.Errorf("Replace alterou o documento: %q", got)
	}
}

// TestFindNextWithoutMatchesIsNotAnError: a query that matches nothing is a
// normal state, not a failure.
func TestFindNextWithoutMatchesIsNotAnError(t *testing.T) {
	application := searchable(t, "nada aqui\n")
	application.Find.Query = "inexistente"
	document, _ := application.Store.Active()
	application.Find.Recompute(document.Buffer().String())

	if err := application.Apply(action.FindNext); err != nil {
		t.Errorf("find_next sem casamentos: %v", err)
	}
	if err := application.Apply(action.FindPrev); err != nil {
		t.Errorf("find_prev sem casamentos: %v", err)
	}
	if err := application.Apply(action.Replace); err != nil {
		t.Errorf("replace sem casamentos: %v", err)
	}
}
