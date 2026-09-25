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

	"github.com/ori-team/oride/internal/action"
	"github.com/ori-team/oride/internal/app"
	"github.com/ori-team/oride/internal/editor"
	"github.com/ori-team/oride/internal/i18n"
	"github.com/ori-team/oride/internal/keymap"
	"github.com/ori-team/oride/internal/tui/component"
	"github.com/ori-team/oride/internal/tui/editorview"
	"github.com/ori-team/oride/internal/tui/focus"
	"github.com/ori-team/oride/internal/tui/layout"
	"github.com/ori-team/oride/internal/tui/menubar"
	"github.com/ori-team/oride/internal/tui/overlay"
	"github.com/ori-team/oride/internal/tui/statusbar"
	"github.com/ori-team/oride/internal/tui/tabs"
	"github.com/ori-team/oride/internal/tui/tree"
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

	// catalog supplies the menu labels, so no surface holds hard-coded text.
	catalog i18n.Catalog

	size    layout.Size
	surface focus.Surface

	// The overlay takes every keystroke while it is open. Its filter and
	// selection live here rather than inside a surface because they are state,
	// and state belongs where the dump can see it.
	overlay  overlay.Kind
	filter   string
	selected int
}

// New builds the model over an application.
func New(application *app.App, keys *keymap.Map) Model {
	return Model{
		application: application,
		keys:        keys,
		graph:       focus.New(),
		catalog:     i18n.LoadRegistry(application.Workspace).Get(i18n.ParseID(application.Config.Locale)),
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
	// The first rule of the focus graph: an active overlay captures every
	// keystroke. Nothing below runs while one is open — not the focus traversal,
	// not the keymap, not typing — because a key typed into a filter that also
	// edited the buffer behind it would be a data-loss bug, not a nuisance.
	if m.overlay.Captures() {
		m.handleOverlayKey(key)
		return
	}

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
		if m.openOverlayFor(chord) {
			return
		}
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
//
// The frame comes back exactly as tall and as wide as the terminal. A frame one
// row short leaves the previous one's remnants on screen, and a row one cell wide
// shifts everything to its right.
func (m Model) Frame() string {
	regions := layout.Compute(m.size, m.wants())
	if regions.Editor.Empty() {
		return ""
	}

	rows := make([]string, 0, m.size.Height)
	rows = append(rows, menubar.Render(m.size.Width, m.menuView()))
	rows = append(rows, tabs.Render(m.size.Width, m.tabsView()))
	rows = append(rows, m.bodyRows(regions)...)
	rows = append(rows, statusbar.Render(m.size.Width, m.statusView()))

	for len(rows) < m.size.Height {
		rows = append(rows, strings.Repeat(" ", m.size.Width))
	}
	if len(rows) > m.size.Height {
		rows = rows[:m.size.Height]
	}

	if m.overlay.Captures() {
		rows = m.paintOverlay(rows, regions)
	}
	return strings.Join(rows, "\n")
}

// paintOverlay draws the open overlay over the body.
//
// It goes over the body rather than over the whole frame, so the menu bar, the
// tabs and the status bar stay visible: the reader keeps knowing what is open and
// where the cursor is while a list is in front of them.
func (m Model) paintOverlay(rows []string, regions layout.Regions) []string {
	view := m.overlayView()
	height := min(regions.Editor.Height, overlay.Frame(len(view.Items))+2)
	region := overlay.Region(regions.Editor, regions.Editor.Width, height)
	if region.Empty() {
		return rows
	}

	painted := overlay.List(region.Width, region.Height, view)
	for index, line := range painted {
		row := region.Y + index
		if row < 0 || row >= len(rows) {
			break
		}
		rows[row] = overlayCells(rows[row], line, region.X)
	}
	return rows
}

// bodyRows composes the tree and the editor, side by side or overlaid.
func (m Model) bodyRows(regions layout.Regions) []string {
	editorRows := editorview.Render(regions.Editor.Width, regions.Editor.Height, m.editorView(regions))

	if regions.Tree.Empty() {
		return editorRows
	}
	treeRows := tree.Render(regions.Tree.Width, regions.Tree.Height, m.treeView())

	if regions.TreeOverlaid {
		for index, treeRow := range treeRows {
			if index >= len(editorRows) {
				break
			}
			editorRows[index] = overlayCells(editorRows[index], treeRow, regions.Tree.X)
		}
		return editorRows
	}

	for index := range editorRows {
		treeRow := ""
		if index < len(treeRows) {
			treeRow = treeRows[index]
		}
		editorRows[index] = layout.Pad(treeRow, regions.Tree.Width) + editorRows[index]
	}
	return editorRows
}

// menuView builds the menu bar's view from the locale catalog.
func (m Model) menuView() menubar.View {
	return menubar.View{
		Labels: menubar.LabelsFor(
			m.catalog.Menu.File, m.catalog.Menu.Edit, m.catalog.Menu.View,
			m.catalog.Menu.Go, m.catalog.Menu.Git, m.catalog.Menu.Help,
		),
		Open: -1,
	}
}

// tabsView builds the tab bar's view from the open documents.
func (m Model) tabsView() tabs.View {
	summaries := m.application.Store.TabSummaries()
	active, hasActive := m.application.Store.ActiveID()

	list := make([]tabs.Tab, 0, len(summaries))
	for _, summary := range summaries {
		list = append(list, tabs.Tab{
			Title:  summary.Title,
			Dirty:  summary.Dirty,
			Active: hasActive && summary.ID == active,
		})
	}
	return tabs.View{Tabs: list, Focused: m.surface == focus.Tabs}
}

// treeView builds the panel's view from the project tree.
func (m Model) treeView() tree.View {
	if m.application.Tree == nil {
		return tree.View{Focused: m.surface == focus.Tree}
	}

	rows := m.application.Tree.FlatRows()
	list := make([]tree.Row, 0, len(rows))
	for _, row := range rows {
		list = append(list, tree.Row{
			Depth: row.Depth, Name: row.Name, IsDir: row.IsDir, Expanded: row.Expanded,
		})
	}
	return tree.View{
		Rows:     list,
		Selected: m.application.Tree.SelectedIndex(),
		Focused:  m.surface == focus.Tree,
	}
}

// editorView builds the viewport's view, extracting only the visible lines.
func (m Model) editorView(regions layout.Regions) editorview.View {
	view := editorview.View{
		Gutter:  m.application.Config.ShowLineNumbers,
		Focused: m.surface == focus.Editor,
	}

	document, err := m.application.Store.Active()
	if err != nil {
		return view
	}
	if caret, caretErr := document.Caret(); caretErr == nil {
		view.Caret = editorview.Position{Line: caret.Line, Column: caret.Column}
		view.HasCaret = true
	}
	view.Lines, view.Offset = visibleLines(document, regions.Editor.Height, view.Caret.Line)
	return view
}

// visibleLines extracts the rows the viewport can show.
//
// The model owns the window because it is state, and because extracting every line
// of the document on each frame would be O(file) per keystroke for a screen that
// shows forty rows.
func visibleLines(document *editor.Document, height, caretLine int) ([]string, int) {
	buffer := document.Buffer()
	start, end := component.Window(buffer.LineCount(), height, caretLine, 0)

	lines := make([]string, 0, end-start)
	for line := start; line < end; line++ {
		text, err := buffer.LineText(line)
		if err != nil {
			break
		}
		lines = append(lines, strings.TrimRight(text, "\n"))
	}
	return lines, start
}

// statusView builds the status bar's view.
func (m Model) statusView() statusbar.View {
	view := statusbar.View{
		Focus:   string(m.surface),
		Message: m.application.Status,
	}

	document, err := m.application.Store.Active()
	if err != nil {
		return view
	}
	view.File = document.TabTitle()
	view.Dirty = document.IsDirty()
	if caret, caretErr := document.Caret(); caretErr == nil {
		// One-based for display: a reader counts lines from one, and a status bar
		// that says Ln 0 is a status bar that looks broken.
		view.CaretLine = caret.Line + 1
		view.CaretCol = caret.Column + 1
		view.HasCaret = true
	}
	return view
}

// openFor maps an action onto the overlay it opens.
//
// Driven by the keymap rather than by hard-coded chords, so remapping the palette
// key still opens the palette.
func overlayFor(bound action.Action) (overlay.Kind, bool) {
	switch bound {
	case action.CommandPalette:
		return overlay.Palette, true
	case action.WhichKey:
		return overlay.WhichKey, true
	case action.Help, action.Welcome:
		return overlay.Help, true
	}
	return overlay.None, false
}

// openOverlayFor opens the overlay a chord resolves to, if any.
func (m *Model) openOverlayFor(chord string) bool {
	parsed, err := keymap.Parse(chord)
	if err != nil {
		return false
	}
	bound, ok := m.keys.Resolve(parsed)
	if !ok {
		return false
	}

	kind, opens := overlayFor(bound)
	if !opens {
		return false
	}
	m.overlay = kind
	m.filter = ""
	m.selected = 0
	return true
}

// handleOverlayKey routes a keystroke inside an open overlay.
//
// Nothing here reaches the model. Closing is the only way back, which is what
// makes the capture rule hold in both directions.
func (m *Model) handleOverlayKey(key tea.KeyPressMsg) {
	chord := key.Keystroke()

	switch chord {
	case "escape", "esc", "ctrl+c":
		m.closeOverlay()
		return
	case "enter":
		m.closeOverlay()
		return
	case "up", "ctrl+p":
		m.moveSelection(-1)
		return
	case "down", "ctrl+n":
		m.moveSelection(1)
		return
	case "backspace":
		m.backspaceFilter()
		return
	}

	// The key carries its own text; Keystroke() is a display form and comes back
	// empty for a key that only has Text set.
	text := key.Text
	if text == "" && isTypable(chord) {
		text = chord
	}
	if text != "" && isTypable(text) {
		m.filter += text
		m.selected = 0
	}
}

// closeOverlay returns input to the surfaces.
func (m *Model) closeOverlay() {
	m.overlay = overlay.None
	m.filter = ""
	m.selected = 0
}

// moveSelection moves the highlight, clamped to the filtered list.
func (m *Model) moveSelection(delta int) {
	total := len(filterItems(m.overlayItems(), m.filter))
	if total == 0 {
		m.selected = 0
		return
	}
	m.selected = min(max(0, m.selected+delta), total-1)
}

// backspaceFilter removes the last filter character. It removes a rune, not a
// byte, or an accented letter would take two presses.
func (m *Model) backspaceFilter() {
	runes := []rune(m.filter)
	if len(runes) == 0 {
		return
	}
	m.filter = string(runes[:len(runes)-1])
	m.selected = 0
}

// overlayItems is the unfiltered content of the open overlay.
func (m Model) overlayItems() []overlay.Item {
	switch m.overlay {
	case overlay.Palette:
		return paletteItems()
	case overlay.WhichKey, overlay.Help:
		return bindingItems(m.keys)
	}
	return nil
}

// paletteItems lists the commands the palette offers.
func paletteItems() []overlay.Item {
	commands := action.Palette()
	items := make([]overlay.Item, 0, len(commands))
	for _, command := range commands {
		items = append(items, overlay.Item{Label: command.String()})
	}
	return items
}

// bindingItems lists every binding, sorted by chord by the keymap itself.
func bindingItems(keys *keymap.Map) []overlay.Item {
	// A nil keymap resolves nothing, and listing an empty table would look like a
	// keymap that lost its bindings.
	if keys == nil {
		return nil
	}

	bindings := keys.Bindings()
	items := make([]overlay.Item, 0, len(bindings))
	for _, binding := range bindings {
		items = append(items, overlay.Item{
			Label:  binding.Chord.String(),
			Detail: binding.Action.String(),
		})
	}
	return items
}

// filterItems keeps the rows whose label matches the filter.
//
// A subsequence match, the same shape a fuzzy finder uses, and it runs on the
// label characters rather than on a score: the score arrives with the pickers in
// M4, and inventing one here would be a second thing to keep consistent.
func filterItems(items []overlay.Item, filter string) []overlay.Item {
	if filter == "" {
		return items
	}

	needle := []rune(strings.ToLower(filter))
	out := make([]overlay.Item, 0, len(items))
	for _, item := range items {
		if subsequence(needle, []rune(strings.ToLower(item.Label))) {
			out = append(out, item)
		}
	}
	return out
}

// subsequence reports whether every needle rune appears in order in haystack.
func subsequence(needle, haystack []rune) bool {
	index := 0
	for _, r := range haystack {
		if index < len(needle) && needle[index] == r {
			index++
		}
	}
	return index == len(needle)
}

// overlayView builds the list the overlay draws.
func (m Model) overlayView() overlay.ListView {
	items := filterItems(m.overlayItems(), m.filter)

	for index := range items {
		items[index].Selected = index == m.selected
	}

	return overlay.ListView{
		Title:  m.overlayTitle(),
		Hint:   m.overlayHint(),
		Items:  items,
		Scroll: scrollFor(m.selected),
		Empty:  "(nenhum resultado)",
	}
}

// overlayTitle names the open overlay.
func (m Model) overlayTitle() string {
	switch m.overlay {
	case overlay.Palette:
		return "Comandos"
	case overlay.WhichKey:
		return "Atalhos"
	case overlay.Help:
		return "Ajuda"
	}
	return ""
}

// overlayHint states how to leave and what the keystrokes do.
func (m Model) overlayHint() string {
	if m.filter != "" {
		return "Enter fecha · Esc cancela · filtro: " + m.filter
	}
	return "digite para filtrar · ↑↓ move · Enter fecha · Esc cancela"
}

// scrollFor keeps the selection inside the window the overlay will draw.
func scrollFor(selected int) int {
	const visibleRows = 16
	if selected < visibleRows {
		return 0
	}
	return selected - visibleRows + 1
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
