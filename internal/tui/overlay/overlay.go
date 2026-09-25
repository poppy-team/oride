// Package overlay renders everything that floats above the surfaces.
//
// It is a thin shell over the list component rather than a list of its own. An
// earlier version hand-rolled the list, the filter, the scroll window and the
// selection marking — six behaviours that the component library already had,
// tested, in the version the migration plan named and this project never added.
package overlay

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

// Kind says which overlay is showing.
type Kind int

// The overlays. None means the surfaces receive input.
const (
	None Kind = iota
	Palette
	WhichKey
	Help
	Welcome
	QuitConfirm
	CloseConfirm
)

// Captures reports whether this overlay takes every keystroke.
//
// Every overlay does. The first rule of docs/ui-ux/focus-graph.md is that an
// active overlay captures all input: while one is open, neither the focused
// surface nor the global keymap is consulted. Anything else would let a key typed
// into a filter also edit the buffer behind it.
func (k Kind) Captures() bool { return k != None }

// Active reports whether an overlay is open.
func (k Kind) Active() bool { return k != None }

// Item is one row.
//
// It satisfies both interfaces the list needs: FilterValue for matching, and Title
// and Description for the default delegate to draw.
type Item struct {
	Label  string
	Detail string
}

// FilterValue is what the component matches a typed filter against.
func (i Item) FilterValue() string { return i.Label }

// Title is the row's main text.
func (i Item) Title() string { return i.Label }

// Description is the row's secondary text.
func (i Item) Description() string { return i.Detail }

// Model is the open overlay, backed by the list component.
type Model struct {
	kind Kind
	list list.Model
}

// New builds a closed overlay.
func New() Model { return Model{kind: None} }

// Kind reports which overlay is showing.
func (m Model) Kind() Kind { return m.kind }

// Active reports whether an overlay is open.
func (m Model) Active() bool { return m.kind.Active() }

// Open shows an overlay with the given rows.
//
// The chrome is configured here rather than by the caller: a caller that forgot to
// disable the status bar would get a row of component text in a two-row panel, and
// the component's own help line is longer than the panel it would sit in.
func (m *Model) Open(kind Kind, title string, items []Item, width, height int) {
	rows := make([]list.Item, 0, len(items))
	for _, item := range items {
		rows = append(rows, item)
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true

	m.kind = kind
	m.list = list.New(rows, delegate, width, height)
	m.list.Title = title
	m.list.SetShowHelp(false)
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(true)
	m.list.SetSize(width, height)
}

// Close hides the overlay and drops its rows.
func (m *Model) Close() {
	m.kind = None
	m.list = list.Model{}
}

// SetSize resizes the overlay, which happens whenever the terminal is resized.
func (m *Model) SetSize(width, height int) {
	if !m.Active() {
		return
	}
	m.list.SetSize(width, height)
}

// Update routes a message to the list.
//
// The component owns the filtering, the scrolling and the selection, which is the
// point of using it: those are the behaviours that were previously hand-written
// and subtly wrong.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.Active() {
		return m, nil
	}

	updated, cmd := m.list.Update(msg)
	m.list = updated
	return m, cmd
}

// View draws the overlay.
func (m Model) View() string {
	if !m.Active() {
		return ""
	}
	return m.list.View()
}

// Selected is the row the reader is on, if any.
func (m Model) Selected() (Item, bool) {
	choice, ok := m.list.SelectedItem().(Item)
	return choice, ok
}

// Filtering reports whether the reader is typing a filter.
//
// Escape belongs to the filter while it is open and to the overlay when it is not:
// closing the panel on the first Escape would also throw away a half-typed filter
// the reader meant to correct.
func (m Model) Filtering() bool { return m.list.FilterState() == list.Filtering }
