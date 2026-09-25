package app

import (
	"github.com/ori-team/oride/internal/action"
	"github.com/ori-team/oride/internal/buffer"
	"github.com/ori-team/oride/internal/editor"
)

// searchCommands drive the in-buffer search.
//
// The matcher and the state already existed and were already compared against the
// oracle; what was missing was the commands that move through the matches. Those
// are here, and a conformance case exercises them, so the semantics are checked
// against the Rust rather than asserted by me.
func searchCommands() map[action.Action]Handler {
	return map[action.Action]Handler{
		action.Find: documentHandler(func(a *App, _ *editor.Document) error {
			// Find opens the bar without the replacement field; revealing the
			// field is a separate action, and a replace field nobody asked for
			// takes a row from the results.
			a.Find.ShowReplace = false
			// The overlay is model state because the oracle dumps it: a
			// conformance case showed the Rust reporting "find" here, which is
			// how an earlier version of this file learned it was not a TUI
			// concern after all.
			a.Overlay = "find"
			return nil
		}),

		action.FindNext: documentHandler(func(a *App, document *editor.Document) error {
			if _, ok := a.Find.Next(); !ok {
				return nil
			}
			return selectCurrentMatch(a, document)
		}),

		action.FindPrev: documentHandler(func(a *App, document *editor.Document) error {
			if _, ok := a.Find.Prev(); !ok {
				return nil
			}
			return selectCurrentMatch(a, document)
		}),

		action.Replace: documentHandler(func(a *App, _ *editor.Document) error {
			// Replace reveals the replacement field; it does not substitute.
			//
			// The conformance case corrected this: an earlier version performed
			// the edit here, and the oracle showed the Rust leaving the document
			// untouched and the selection on the match. Substituting is a
			// separate flow, and modelling it as one action made a keystroke do
			// something the reference does not.
			a.Find.ShowReplace = true
			return nil
		}),
	}
}

// selectCurrentMatch puts the selection on the current match.
//
// Selecting rather than only moving the caret: the match has to be visible, and a
// replacement acts on a selection. Selecting is also what makes the search
// observable in the state dump, which is how the conformance case checks it.
func selectCurrentMatch(a *App, document *editor.Document) error {
	match, ok := a.Find.CurrentMatch()
	if !ok {
		return nil
	}
	document.SelectByteRange(buffer.Offset(match.Start), buffer.Offset(match.End))
	return nil
}
