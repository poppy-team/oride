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
	"github.com/ori-team/oride/internal/buffer"
	"github.com/ori-team/oride/internal/config"
	"github.com/ori-team/oride/internal/editor"
	"github.com/ori-team/oride/internal/i18n"
	"github.com/ori-team/oride/internal/keymap"
	"github.com/ori-team/oride/internal/search"
	"github.com/ori-team/oride/internal/tui/component"
	"github.com/ori-team/oride/internal/tui/editorview"
	"github.com/ori-team/oride/internal/tui/findbar"
	"github.com/ori-team/oride/internal/tui/focus"
	"github.com/ori-team/oride/internal/tui/layout"
	"github.com/ori-team/oride/internal/tui/menubar"
	"github.com/ori-team/oride/internal/tui/overlay"
	"github.com/ori-team/oride/internal/tui/statusbar"
	"github.com/ori-team/oride/internal/tui/tabs"
	"github.com/ori-team/oride/internal/tui/theme"
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
	// theme is resolved once, from configuration and a profile, and handed to
	// every surface. Resolving it per frame would mean rebuilding the same styles
	// on every keystroke.
	theme theme.Theme

	size    layout.Size
	surface focus.Surface

	// The overlay takes every keystroke while it is open. Its filter and
	// selection live here rather than inside a surface because they are state,
	// and state belongs where the dump can see it.
	overlay overlay.Model
}

// New builds the model over an application.
func New(application *app.App, keys *keymap.Map) Model {
	return Model{
		application: application,
		keys:        keys,
		graph:       focus.New(),
		catalog:     i18n.LoadRegistry(application.Workspace).Get(i18n.ParseID(application.Config.Locale)),
		theme:       resolveTheme(application),
		surface:     focus.Editor,
	}
}

// WithProfile rebuilds the theme for a colour profile.
//
// A method rather than a constructor argument because the profile comes from the
// terminal the program is actually talking to, which is known after the model
// exists — and injecting it keeps a test off the process environment.
func (m Model) WithProfile(profile theme.Profile) Model {
	m.theme = theme.New(m.palette().UI, m.palette().Syntax, profile)
	return m
}

// palette resolves the colours the theme is built from.
//
// The named theme supplies them; Config.UI and Config.Syntax are the user's
// overrides on top. Reading Config.UI alone — which an earlier version did — gives
// an empty palette, because the shipped configuration names a theme instead of
// spelling out its colours, and the frame comes out monochrome with every test
// still passing.
func (m Model) palette() config.ThemeDefinition {
	return paletteOf(m.application)
}

// resolveTheme builds the theme a fresh model starts with.
func resolveTheme(application *app.App) theme.Theme {
	definition := paletteOf(application)
	return theme.New(definition.UI, definition.Syntax, theme.TrueColor)
}

// paletteOf resolves the effective theme definition for an application.
func paletteOf(application *app.App) config.ThemeDefinition {
	registry := config.LoadThemeRegistry(application.Workspace)
	if definition, found := registry.Get(config.NormalizeThemeName(application.Config.Theme)); found {
		return definition
	}
	// A theme name that resolves to nothing falls back to the configuration's own
	// colours rather than to an empty palette.
	return config.ThemeDefinition{UI: application.Config.UI, Syntax: application.Config.Syntax}
}

// Resize sets the measured size.
//
// Exported for callers that render without a runtime — the frame-printing
// command and the golden tests — so they do not have to fabricate a message to
// tell the model how big the terminal is.
func (m Model) Resize(width, height int) Model {
	m.size = layout.Size{Width: width, Height: height}
	return m
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
		if m.application.Quit {
			// The model sets the flag and the runtime ends the program. Nothing
			// read the flag before, so the editor could not be closed at all.
			return m, tea.Quit
		}
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
	if m.capturesInput() {
		m.handleOverlayKey(key)
		return
	}

	chord := key.Keystroke()

	// The focused surface gets the key first. Before this, the arrows moved the
	// document caret whatever the focus was, and the tree could be looked at but
	// not walked.
	if m.handleFocusedSurfaceKey(chord) {
		return
	}

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
	if !regions.FindBar.Empty() {
		rows = append(rows, findbar.Render(regions.FindBar.Width, m.findView())...)
	}
	rows = append(rows, statusbar.Render(m.size.Width, m.statusView()))

	for len(rows) < m.size.Height {
		rows = append(rows, strings.Repeat(" ", m.size.Width))
	}
	if len(rows) > m.size.Height {
		rows = rows[:m.size.Height]
	}

	if m.overlay.Active() {
		rows = m.paintOverlay(rows, regions)
	}
	return strings.Join(rows, "\n")
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
	return tabs.View{Tabs: list, Focused: m.surface == focus.Tabs, Theme: m.theme}
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
		Theme:   m.theme,
	}

	document, err := m.application.Store.Active()
	if err != nil {
		return view
	}
	if caret, caretErr := document.Caret(); caretErr == nil {
		view.Caret = editorview.Position{Line: caret.Line, Column: caret.Column}
		view.HasCaret = true
	}
	view.Selection = selectionOf(document)
	view.Matches, view.CurrentMatch = matchesOf(document, m.application.Find)
	view.Lines, view.Offset = visibleLines(document, regions.Editor.Height, view.Caret.Line)
	return view
}

// selectionOf converts the document's byte-range selection into the line and
// column coordinates the viewport paints in.
//
// The document keeps offsets in bytes, because that is the canonical index; the
// viewport paints cells, so the conversion happens here, once, at the boundary
// between the two — and a selection that cannot be resolved paints nothing rather
// than failing the frame.
func selectionOf(document *editor.Document) editorview.Selection {
	selection := document.Selection()
	if selection.IsEmpty() {
		return editorview.Selection{Empty: true}
	}

	start, startErr := document.Buffer().ByteToCaret(selection.Start())
	end, endErr := document.Buffer().ByteToCaret(selection.End())
	if startErr != nil || endErr != nil {
		return editorview.Selection{Empty: true}
	}

	return editorview.Selection{
		Start: editorview.Position{Line: start.Line, Column: start.Column},
		End:   editorview.Position{Line: end.Line, Column: end.Column},
	}
}

// matchesOf converts the search results into the coordinates the viewport paints
// in.
//
// The search keeps byte offsets because that is the canonical index; the viewport
// paints cells. The conversion happens here, once, at the boundary — and a match
// that cannot be resolved is dropped rather than failing the frame, because a
// stale offset is a normal consequence of editing while a search is open.
func matchesOf(document *editor.Document, found search.State) ([]editorview.Match, int) {
	if len(found.Matches) == 0 {
		return nil, -1
	}

	matches := make([]editorview.Match, 0, len(found.Matches))
	for _, match := range found.Matches {
		start, startErr := document.Buffer().ByteToCaret(buffer.Offset(match.Start))
		end, endErr := document.Buffer().ByteToCaret(buffer.Offset(match.End))
		if startErr != nil || endErr != nil {
			continue
		}
		matches = append(matches, editorview.Match{
			Start: editorview.Position{Line: start.Line, Column: start.Column},
			End:   editorview.Position{Line: end.Line, Column: end.Column},
		})
	}

	// The index has to be remapped: dropping a match shifts everything after it,
	// and marking the wrong result as current is worse than marking none.
	current := -1
	if found.Current >= 0 && found.Current < len(found.Matches) {
		current = found.Current
		if len(matches) != len(found.Matches) {
			current = -1
		}
	}
	return matches, current
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
		Theme:   m.theme,
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

// capturesInput reports whether an overlay owns the keyboard.
//
// Two overlays exist and they are not the same thing: the TUI's own (the palette,
// which-key and help) and the model's, which the oracle dumps and which the search
// uses. Checking only the first was a real defect — opening the search bar set the
// model's overlay, the TUI saw none, and typing went into the document behind the
// bar.
func (m Model) capturesInput() bool {
	return m.overlay.Active() || m.application.Overlay != overlayNone
}

// handleFocusedSurfaceKey gives the focused surface its own keys.
//
// It reports whether the key was consumed. The tree is the first surface with a
// local keyboard, and the shape is the one the focus graph documents: local keys
// resolve before the global keymap, and an unconsumed key falls through to it.
func (m *Model) handleFocusedSurfaceKey(chord string) bool {
	if m.application.Focus != app.FocusTree || !m.application.ShowTree {
		return false
	}

	tree := m.application.Tree
	if tree == nil {
		return false
	}

	switch chord {
	case "up":
		tree.MoveSelection(-1)
		return true
	case "down":
		tree.MoveSelection(1)
		return true
	case "right":
		_ = tree.ExpandSelected()
		return true
	case "left":
		_ = tree.CollapseOrParent()
		return true
	case "enter":
		// Activating opens a file or toggles a directory, and the store takes
		// the path — the panel does not know about documents.
		if path, opened, err := tree.Activate(); err == nil && opened {
			_, _ = m.application.Store.OpenPath(path)
		}
		return true
	}
	return false
}

// handleFindKey routes a keystroke inside the search bar.
//
// Typing goes into the query and the matches are recomputed on every keystroke, so
// the highlight follows what is being typed. Nothing here reaches the document:
// the capture rule holds in both directions.
func (m *Model) handleFindKey(key tea.KeyPressMsg) {
	chord := key.Keystroke()

	switch chord {
	case "escape", "esc", "ctrl+c":
		m.application.Overlay = overlayNone
		return
	case "enter":
		_ = m.application.Apply(action.FindNext)
		return
	case "shift+enter":
		_ = m.application.Apply(action.FindPrev)
		return
	case "tab":
		// The replacement field is revealed, not switched to: the query stays
		// visible while the replacement is typed, because the two are read
		// together.
		m.application.Find.ShowReplace = !m.application.Find.ShowReplace
		return
	case "backspace":
		m.backspaceFind()
		return
	}

	text := key.Text
	if text == "" && isTypable(chord) {
		text = chord
	}
	if text != "" && isTypable(text) {
		m.appendFind(text)
	}
}

// appendFind adds typed text to whichever field is being edited, then searches.
func (m *Model) appendFind(text string) {
	if m.application.Find.ShowReplace {
		m.application.Find.Replace += text
		return
	}
	m.application.Find.Query += text
	m.recomputeFind()
}

// backspaceFind removes the last rune from whichever field is being edited.
//
// A rune, not a byte: an accented letter would otherwise take two presses to
// delete and leave an invalid fragment behind.
func (m *Model) backspaceFind() {
	target := &m.application.Find.Query
	if m.application.Find.ShowReplace {
		target = &m.application.Find.Replace
	}

	runes := []rune(*target)
	if len(runes) == 0 {
		return
	}
	*target = string(runes[:len(runes)-1])

	if !m.application.Find.ShowReplace {
		m.recomputeFind()
	}
}

// recomputeFind re-runs the search against the document.
//
// A missing document leaves the previous matches alone rather than clearing them:
// an empty buffer would otherwise look like a query that stopped matching.
func (m *Model) recomputeFind() {
	document, err := m.application.Store.Active()
	if err != nil {
		return
	}
	m.application.Find.Recompute(document.Buffer().String())
}

// overlayNone is the model's value for no overlay open.
const overlayNone = "none"

// wants is what the model asks the layout to show.
func (m Model) wants() layout.Wants {
	return layout.Wants{
		ShowTree:   m.application.ShowTree,
		TreeWidth:  treeWidth,
		FindHeight: m.findHeight(),
	}
}

// findHeight is how many rows the search bar wants.
//
// The overlay name is the model's, and the extra row for the replacement field is
// the search state's: the bar grows when the field is revealed, and the layout is
// told before it divides the screen.
func (m Model) findHeight() int {
	if m.application.Overlay != overlayFind {
		return 0
	}
	if m.application.Find.ShowReplace {
		return 2
	}
	return 1
}

// overlayFind is the value the model and the oracle use for the search overlay.
//
// A constant rather than a literal because the string is a contract: it is
// compared against the Rust in the conformance harness, and a typo here would
// become a parity failure rather than a compile error.
const overlayFind = "find"

// findView builds the search bar's view from the search state.
func (m Model) findView() findbar.View {
	found := m.application.Find

	return findbar.View{
		Query:         found.Query,
		Replace:       found.Replace,
		ShowReplace:   found.ShowReplace,
		CaseSensitive: found.Options.CaseSensitive,
		IgnoreAccents: found.Options.IgnoreAccents,
		WholeWord:     found.Options.WholeWord,
		UseRegex:      found.Options.UseRegex,
		Current:       found.Current + 1,
		Total:         len(found.Matches),
		RegexError:    found.RegexError,
		Theme:         m.theme,
	}
}

// treeWidth is the default tree column until the width becomes configurable.
const treeWidth = 30

// overlayCells writes text over a row starting at a cell offset, replacing as many
// cells as the text occupies.
//
// It builds on layout.Slice rather than carrying its own walk: the copy here was
// written before that helper had three callers, and two implementations of cell
// arithmetic is one more place for a wide character to come out a column short.
func overlayCells(row, text string, at int) string {
	total := layout.Width(row)
	if at >= total || total <= 0 {
		return row
	}

	width := min(layout.Width(text), total-at)
	before := layout.Slice(row, 0, at)
	after := layout.Slice(row, at+width, total-at-width)

	return layout.Pad(before, at) + layout.Pad(text, width) + layout.Pad(after, total-at-width)
}

// paintOverlay draws the open overlay over the body.
//
// It goes over the body rather than over the whole frame, so the menu bar, the tabs
// and the status bar stay visible: the reader keeps knowing what is open and where
// the cursor is while a list is in front of them.
func (m Model) paintOverlay(rows []string, regions layout.Regions) []string {
	rendered := m.overlay.View()
	if rendered == "" {
		return rows
	}

	width, height := m.overlaySize()
	region := centerRegion(regions.Editor, width, height)
	if region.Empty() {
		return rows
	}

	for index, line := range strings.Split(rendered, "\n") {
		row := region.Y + index
		if row < 0 || row >= len(rows) || index >= region.Height {
			break
		}
		rows[row] = overlayCells(rows[row], layout.Pad(line, region.Width), region.X)
	}
	return rows
}

// centerRegion places a box in the middle of an area.
func centerRegion(area layout.Region, width, height int) layout.Region {
	width = min(width, area.Width)
	height = min(height, area.Height)
	if width <= 0 || height <= 0 {
		return layout.Region{}
	}
	return layout.Region{
		X:      area.X + (area.Width-width)/2,
		Y:      area.Y + (area.Height-height)/2,
		Width:  width,
		Height: height,
	}
}

// overlaySize is how much of the screen an overlay takes.
//
// Capped rather than proportional: a palette filling a 200-column terminal would
// be a wall of text with a filter at the bottom, and the surface behind it stops
// being visible — which is what makes it an overlay rather than another screen.
func (m Model) overlaySize() (int, int) {
	width := min(max(m.size.Width-8, 24), 76)
	height := min(max(m.size.Height-8, 6), 20)
	return width, height
}

// overlayContent maps an action onto the overlay it opens.
//
// Driven by the keymap rather than by hard-coded chords, so remapping the palette
// key still opens the palette.
func overlayContent(bound action.Action, keys *keymap.Map) (overlay.Kind, []overlay.Item, string, bool) {
	switch bound {
	case action.CommandPalette:
		return overlay.Palette, paletteItems(), "Comandos", true
	case action.WhichKey:
		return overlay.WhichKey, bindingItems(keys), "Atalhos", true
	case action.Help, action.Welcome:
		return overlay.Help, bindingItems(keys), "Ajuda", true
	}
	return overlay.None, nil, "", false
}

// paletteItems is the command palette's rows.
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

// handleOverlayKey routes a keystroke inside an open overlay.
//
// The list component owns the filter, the scrolling and the selection, so the key
// goes to it rather than being interpreted here — which is the whole reason for
// using it.
func (m *Model) handleOverlayKey(key tea.KeyPressMsg) {
	if m.application.Overlay == overlayFind {
		m.handleFindKey(key)
		return
	}

	// Escape has two owners and they must not both act: while the filter is open it
	// belongs to the filter, and closing on the first Escape would discard a
	// half-typed filter the reader meant to correct.
	chord := key.Keystroke()
	if chord == "escape" || chord == "esc" || chord == "ctrl+c" {
		if !m.overlay.Filtering() {
			m.overlay.Close()
			return
		}
	}

	updated, _ := m.overlay.Update(key)
	m.overlay = updated
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

	kind, items, title, opens := overlayContent(bound, m.keys)
	if !opens {
		return false
	}

	width, height := m.overlaySize()
	m.overlay.Open(kind, title, items, width, height)
	return true
}
