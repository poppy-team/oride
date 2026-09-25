package app

import (
	"github.com/ori-team/oride/internal/action"
	"github.com/ori-team/oride/internal/editor"
)

// movement is one cursor direction, in both of its variants.
//
// Plain and extend are the same movement with the selection flag set, which is
// the domain truth rather than a coincidence: declaring them as a pair keeps the
// two from drifting apart, and stops eight directions becoming sixteen entries.
type movement struct {
	plain  action.Action
	extend action.Action
	move   func(*editor.Document, bool) error
}

// movements pairs every direction with the document method it calls.
var movements = []movement{
	{action.MoveLeftPlain, action.MoveLeftExtend, (*editor.Document).MoveLeft},
	{action.MoveRightPlain, action.MoveRightExtend, (*editor.Document).MoveRight},
	{action.MoveUpPlain, action.MoveUpExtend, (*editor.Document).MoveUp},
	{action.MoveDownPlain, action.MoveDownExtend, (*editor.Document).MoveDown},
	{action.MoveLineStartPlain, action.MoveLineStartExtend, (*editor.Document).MoveLineStart},
	{action.MoveLineEndPlain, action.MoveLineEndExtend, (*editor.Document).MoveLineEnd},
	{action.MoveDocStartPlain, action.MoveDocStartExtend, (*editor.Document).MoveBufferStart},
	{action.MoveDocEndPlain, action.MoveDocEndExtend, (*editor.Document).MoveBufferEnd},
}

// movementCommands derives both variants of every direction.
func movementCommands() map[action.Action]Handler {
	table := make(map[action.Action]Handler, len(movements)*2)

	for _, direction := range movements {
		table[direction.plain] = documentHandler(func(_ *App, document *editor.Document) error {
			return direction.move(document, false)
		})
		table[direction.extend] = documentHandler(func(_ *App, document *editor.Document) error {
			return direction.move(document, true)
		})
	}
	return table
}
