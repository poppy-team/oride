package app

import (
	"github.com/ori-team/oride/internal/action"
	"github.com/ori-team/oride/internal/editor"
)

// viewCommands change what is on screen and where the focus is.
//
// They do not touch a document, so they keep working with an empty store — which
// is what lets someone toggle the tree or quit before opening a file.
func viewCommands() map[action.Action]Handler {
	return map[action.Action]Handler{
		action.ToggleTree: modelHandler(func(a *App, _ *editor.Document) error {
			a.ShowTree = !a.ShowTree
			return nil
		}),
		action.FocusTree: modelHandler(func(a *App, _ *editor.Document) error {
			a.Focus = FocusTree
			return nil
		}),
		action.FocusEditor: modelHandler(func(a *App, _ *editor.Document) error {
			a.Focus = FocusEditor
			return nil
		}),
	}
}
