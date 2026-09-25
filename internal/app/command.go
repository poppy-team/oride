package app

import (
	"fmt"

	"github.com/ori-team/oride/internal/action"
	"github.com/ori-team/oride/internal/editor"
)

// Handler applies one action to the model.
//
// The document arrives as an argument rather than being looked up inside the
// handler: one lookup decides once whether the action can run at all, so no
// handler repeats the "no active document" branch. A handler that does not touch
// a document ignores the argument.
type Handler struct {
	// NeedsDocument says the action cannot run without an active document.
	//
	// It is declared rather than inferred because the palette needs the same
	// answer: an action that needs a document is one to gray out when no
	// document is open. Keeping it in one place means the dispatch and the
	// palette cannot disagree.
	NeedsDocument bool
	Apply         func(*App, *editor.Document) error
}

// documentHandler marks a handler that cannot run without an active document.
func documentHandler(apply func(*App, *editor.Document) error) Handler {
	return Handler{NeedsDocument: true, Apply: apply}
}

// modelHandler marks a handler that runs whether or not a document is open.
func modelHandler(apply func(*App, *editor.Document) error) Handler {
	return Handler{Apply: apply}
}

// commands is the merged action table, built once at start-up.
var commands = buildCommands()

// buildCommands composes the per-domain tables into one.
//
// A table rather than a switch: the domains are separate files, so a change to
// search cannot collide with a change to movement, and adding an action is one
// entry instead of a new arm in a growing function.
func buildCommands() map[action.Action]Handler {
	merged := map[action.Action]Handler{}

	for _, group := range []struct {
		name     string
		commands map[action.Action]Handler
	}{
		{"arquivo", fileCommands()},
		{"edição", editCommands()},
		{"movimento", movementCommands()},
		{"seleção", selectionCommands()},
		{"visão", viewCommands()},
	} {
		for id, handler := range group.commands {
			if _, duplicate := merged[id]; duplicate {
				// Two registrations of one id would silently pick whichever the
				// map kept. This is a programmer invariant, not a user error.
				panic(fmt.Sprintf("ação %q registrada duas vezes (grupo %s)", id, group.name))
			}
			merged[id] = handler
		}
	}
	return merged
}

// CommandCount reports how many actions the model can execute.
//
// Exported so a test can assert the table grew rather than shrank: an action
// that loses its registration stops working without failing anything.
func CommandCount() int { return len(commands) }
