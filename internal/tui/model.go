// Package tui is the presentation layer, on Bubble Tea v2.
//
// It turns the state of internal/app into a frame and input into messages, and
// does nothing else. No editing rule, no text-layout rule and no protocol lives
// here: a decision that has to hold when there is no terminal belongs in the
// model.
//
// model.go is the composition root, and the only file in this package allowed to
// import internal/app — rule R3 in internal/architecture fails the build
// otherwise. Surfaces receive what they need as data, which is what makes one
// deletable by removing a folder and the line that calls it.
package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/ori-team/oride/internal/app"
	"github.com/ori-team/oride/internal/keymap"
	"github.com/ori-team/oride/internal/tui/focus"
	"github.com/ori-team/oride/internal/tui/layout"
)

// windowTitle is what the terminal title bar shows.
const windowTitle = "oride"

// Model is the Bubble Tea model.
//
// It owns what is presentation: the measured size, which surface holds focus, and
// the graph that decides where focus goes next. Everything else lives in app,
// which is the single implementation of the behaviour.
type Model struct {
	application *app.App
	keys        *keymap.Map
	graph       focus.Graph

	size    layout.Size
	surface focus.Surface
}

// New builds the model over an application.
func New(application *app.App, keys *keymap.Map) Model {
	return Model{
		application: application,
		keys:        keys,
		graph:       focus.New(),
		surface:     focus.Editor,
	}
}

// Init performs no work.
//
// Init has no error path, so anything that can fail belongs in a Cmd that reports
// the failure as a message.
func (m Model) Init() tea.Cmd { return nil }

// Update handles one message.
//
// It never blocks: anything slow returns a Cmd. The switch is over message types
// rather than a chain of conditions, so an unhandled message is visibly unhandled.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch message := msg.(type) {
	case tea.WindowSizeMsg:
		m.size = layout.Size{Width: message.Width, Height: message.Height}
		return m, nil

	case tea.KeyPressMsg:
		m.handleKey(message)
		return m, nil
	}
	return m, nil
}

// handleKey routes one keystroke.
//
// Focus traversal is handled before the keymap, because Tab is navigation in
// every surface and must not be remappable into an editing action. Everything
// else goes through the keymap, and a keystroke the keymap does not resolve is
// text — which is what makes typing work at all.
func (m *Model) handleKey(key tea.KeyPressMsg) {
	chord := key.Keystroke()

	switch chord {
	case "tab":
		m.moveFocus(false)
		return
	case "shift+tab":
		m.moveFocus(true)
		return
	}

	if chord != "" {
		if err := m.application.ApplyKey(chord); err == nil {
			return
		}
	}

	// The key carries its own text. Keystroke() is a display form and comes back
	// empty for a key that only has Text set, which is exactly how a plain letter
	// arrives — so resolving the character from Keystroke() alone would type
	// nothing at all.
	text := key.Text
	if text == "" && isTypable(chord) {
		text = chord
	}
	if text != "" {
		_ = m.application.ApplyText(text)
	}
}

// moveFocus walks the graph and records the result in the model.
func (m *Model) moveFocus(backwards bool) {
	next := m.graph.Next(m.surface, m.surfaceVisible, backwards)
	m.surface = next

	// The graph speaks the presentation vocabulary; the model speaks its own.
	// Syncing here is what keeps a single source of truth for the dump while the
	// graph stays independent of it.
	m.application.Focus = focusToModel(next)
}

// surfaceVisible reports whether a surface may hold focus.
//
// The editor and the chrome are always drawn. The panels are not, and focus must
// never rest where nothing is painted.
func (m Model) surfaceVisible(surface focus.Surface) bool {
	switch surface {
	case focus.Editor, focus.MenuBar, focus.Tabs, focus.StatusBar:
		return true
	case focus.Tree:
		return m.application.ShowTree
	case focus.SCM:
		return m.application.SCM
	case focus.Terminal:
		// The terminal panel arrives with its state in a later slice.
		return false
	}
	return false
}

// focusToModel maps a surface onto the model's focus vocabulary.
func focusToModel(surface focus.Surface) app.Focus {
	switch surface {
	case focus.Tree:
		return app.FocusTree
	case focus.Terminal:
		return app.FocusTerminal
	case focus.SCM:
		return app.FocusSCM
	default:
		return app.FocusEditor
	}
}

// isTypable reports whether a keystroke is a single printable character.
//
// A named key that the keymap did not resolve — F5, Escape, a bare modifier — must
// not be inserted as text, or an unbound key would type its own name into the
// buffer.
func isTypable(chord string) bool {
	runes := []rune(chord)
	return len(runes) == 1 && runes[0] >= 0x20 && runes[0] != 0x7f
}

// View returns the declarative view.
//
// Terminal capability is declared, not toggled per frame: the runtime reads this
// and reconciles. Toggling imperatively is how a terminal is left in the alternate
// screen after a crash.
func (m Model) View() tea.View {
	view := tea.NewView(m.Frame())
	view.AltScreen = true
	view.WindowTitle = windowTitle
	view.MouseMode = m.mouseMode()
	return view
}

// mouseMode derives the mode from configuration.
//
// Off by default, and off means off: an editor that captures the mouse without
// being asked takes text selection away from the terminal.
func (m Model) mouseMode() tea.MouseMode {
	if !m.application.Config.Mouse {
		return tea.MouseModeNone
	}
	return tea.MouseModeCellMotion
}

// Frame renders one screen.
//
// Exported so tests and golden frames can call it without a running tea.Program:
// the frame is a pure function of the model, and a golden file should exercise the
// function rather than the runtime.
func (m Model) Frame() string {
	regions := layout.Compute(m.size, m.wants())
	if regions.Editor.Empty() {
		return ""
	}

	rows := make([]string, 0, m.size.Height)
	rows = append(rows, m.renderRow(string(focus.MenuBar), regions.MenuBar))
	rows = append(rows, m.renderRow(string(focus.Tabs), regions.Tabs))

	body := m.renderBody(regions)
	rows = append(rows, body...)

	for row := len(rows); row < regions.StatusBar.Y; row++ {
		rows = append(rows, strings.Repeat(" ", m.size.Width))
	}
	rows = append(rows, layout.Pad(m.statusLine(), m.size.Width))

	return strings.Join(rows, "\n")
}

// wants is what the model asks the layout to show.
func (m Model) wants() layout.Wants {
	return layout.Wants{
		ShowTree:  m.application.ShowTree,
		TreeWidth: treeWidth,
	}
}

// treeWidth is the default tree column until the width becomes configurable.
const treeWidth = 30

// renderBody paints the body rows, including a floating tree when the layout
// decided to overlay it.
func (m Model) renderBody(regions layout.Regions) []string {
	// The body is as wide as the editor plus whatever the tree takes to its left.
	// Using the editor's width alone leaves the tree's columns off the frame, and
	// with a side-by-side tree the row comes up exactly tree-width short.
	bodyWidth := regions.Editor.X + regions.Editor.Width

	rows := make([]string, 0, regions.Editor.Height)
	for range regions.Editor.Height {
		rows = append(rows, strings.Repeat(" ", bodyWidth))
	}

	if !regions.Tree.Empty() {
		m.paintTree(rows, regions)
	}
	return rows
}

// paintTree writes the tree into the body, either in its own column or over the
// editor's left edge.
func (m Model) paintTree(rows []string, regions layout.Regions) {
	for row := 0; row < regions.Tree.Height && row < len(rows); row++ {
		text := layout.Pad("", regions.Tree.Width)
		if row == 0 {
			text = layout.Pad(" "+string(focus.Tree)+" ", regions.Tree.Width)
		}
		rows[row] = overlayCells(rows[row], text, regions.Tree.X)
	}
}

// overlayCells writes text over a row starting at a cell offset, replacing as many
// cells as the text occupies.
//
// Cell arithmetic, not byte arithmetic: the layout measured in cells, and mixing
// the two units is how a row with a wide character in it ends up one column short.
func overlayCells(row, text string, at int) string {
	total := layout.Width(row)
	if at >= total || total <= 0 {
		return row
	}

	width := min(layout.Width(text), total-at)
	before := sliceCells(row, 0, at)
	after := sliceCells(row, at+width, total-at-width)

	return layout.Pad(before, at) + layout.Pad(text, width) + layout.Pad(after, total-at-width)
}

// sliceCells extracts width cells from a row, starting at a cell offset.
func sliceCells(row string, from, width int) string {
	if width <= 0 {
		return ""
	}

	skipped := 0
	taken := 0
	var out strings.Builder

	for _, r := range row {
		cellWidth := layout.Width(string(r))
		if skipped+cellWidth <= from {
			skipped += cellWidth
			continue
		}
		if taken+cellWidth > width {
			break
		}
		out.WriteRune(r)
		taken += cellWidth
	}
	return out.String()
}

// renderRow paints a one-line surface.
func (m Model) renderRow(name string, region layout.Region) string {
	if region.Empty() {
		return ""
	}
	return layout.Pad(" "+name+" ", region.Width)
}

// statusLine is the bottom row.
func (m Model) statusLine() string {
	line := " " + string(m.surface)
	if m.application.Status != "" {
		line += "  " + m.application.Status
	}
	if m.application.Quit {
		line += "  [encerrando]"
	}
	return line
}
