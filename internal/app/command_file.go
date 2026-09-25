package app

import (
	"github.com/ori-team/oride/internal/action"
	"github.com/ori-team/oride/internal/editor"
)

// fileCommands are the actions about the session and the documents in it.
func fileCommands() map[action.Action]Handler {
	return map[action.Action]Handler{
		action.Quit: modelHandler(func(a *App, _ *editor.Document) error {
			a.Quit = true
			return nil
		}),

		// Save, SaveAs and SaveAll are one behaviour here: the headless model
		// has no browser, so a document with no path keeps its dirty flag and
		// says so. See markSaved.
		action.Save:    documentHandler(func(a *App, _ *editor.Document) error { return a.markSaved() }),
		action.SaveAs:  documentHandler(func(a *App, _ *editor.Document) error { return a.markSaved() }),
		action.SaveAll: documentHandler(func(a *App, _ *editor.Document) error { return a.markSaved() }),
	}
}
