package app

import (
	"github.com/ori-team/oride/internal/action"
	"github.com/ori-team/oride/internal/editor"
)

// selectionCommands change what is selected and how many carets exist.
func selectionCommands() map[action.Action]Handler {
	return map[action.Action]Handler{
		action.SelectAll: documentHandler(func(_ *App, document *editor.Document) error {
			document.SelectAll()
			return nil
		}),

		action.AddCursorAbove: documentHandler(func(_ *App, document *editor.Document) error {
			return document.AddCursorAbove()
		}),
		action.AddCursorBelow: documentHandler(func(_ *App, document *editor.Document) error {
			return document.AddCursorBelow()
		}),
		action.ClearExtraCursors: documentHandler(func(_ *App, document *editor.Document) error {
			document.ClearExtraCarets()
			return nil
		}),
	}
}
