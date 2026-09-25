package app

import (
	"github.com/ori-team/oride/internal/action"
	"github.com/ori-team/oride/internal/editor"
)

// editCommands change the document's text or its history.
func editCommands() map[action.Action]Handler {
	return map[action.Action]Handler{
		action.Undo: documentHandler(func(_ *App, document *editor.Document) error {
			_, err := document.Undo()
			return err
		}),
		action.Redo: documentHandler(func(_ *App, document *editor.Document) error {
			_, err := document.Redo()
			return err
		}),

		action.InsertNewline: documentHandler(func(_ *App, document *editor.Document) error {
			return document.InsertText("\n")
		}),
		action.InsertTab: documentHandler(func(a *App, document *editor.Document) error {
			return a.indent(document)
		}),

		action.Backspace: documentHandler(func(_ *App, document *editor.Document) error {
			return document.Backspace()
		}),
		action.Delete: documentHandler(func(_ *App, document *editor.Document) error {
			return document.DeleteForward()
		}),
	}
}
