// Package app wires the editor into something a scripted case can drive.
//
// It is deliberately not the TUI. The point is a headless surface that produces
// the same observable state the Rust oracle does, so the differential harness
// has something to compare. The TUI is a presentation of this state, not a
// second implementation of it.
package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ori-team/oride/internal/action"
	"github.com/ori-team/oride/internal/buffer"
	"github.com/ori-team/oride/internal/config"
	"github.com/ori-team/oride/internal/editor"
	"github.com/ori-team/oride/internal/fs"
	"github.com/ori-team/oride/internal/keymap"
	"github.com/ori-team/oride/internal/search"
)

// ErrNotImplemented marks an action the Go side does not handle yet.
//
// Explicit rather than silent: an action that quietly does nothing looks exactly
// like an action that worked, and a case exercising it would report parity it
// never measured.
var ErrNotImplemented = errors.New("not implemented in the Go headless runner")

// Focus is the panel holding the keyboard.
type Focus string

// Focus values.
const (
	FocusEditor   Focus = "editor"
	FocusTree     Focus = "tree"
	FocusTerminal Focus = "terminal"
	FocusSCM      Focus = "scm"
)

// App is the headless editor state.
type App struct {
	Store  *editor.Store
	Config config.Config
	Keymap *keymap.Map
	Focus  Focus
	Status string
	Quit   bool
	SCM    bool
	// ShowTree is whether the tree panel is visible. Distinct from Tree, which
	// is the tree itself: a panel can be hidden without the workspace being
	// closed.
	ShowTree bool
	// Workspace is the directory that reported paths are relative to.
	//
	// State dumps replace this prefix with a placeholder, so a fixture recorded
	// in one directory is valid in another. Without it every dump would carry the
	// machine's temporary path and no two runs would ever agree.
	Workspace string
	// Tree is the project tree, when one was opened.
	Tree *fs.Tree
	// Find is the in-buffer search state.
	Find search.State
	// Overlay names what floats above the surfaces, in the oracle's vocabulary.
	//
	// It is model state, not presentation: the Rust dumps it, so which overlay is
	// open is observable behaviour — and a conformance case corrected an earlier
	// version of this file that treated opening find as a TUI concern.
	Overlay string
}

// WorkspacePlaceholder stands in for the workspace path in a dump.
const WorkspacePlaceholder = "<workspace>"

// New builds an app from a configuration and keymap.
func New(store *editor.Store, cfg config.Config, keys *keymap.Map) *App {
	return &App{
		Store:    store,
		Config:   cfg,
		Keymap:   keys,
		Focus:    FocusEditor,
		ShowTree: true,
		Find:     search.NewState(),
		Overlay:  "none",
	}
}

// ApplyKey resolves a canonical chord and applies what it is bound to.
func (a *App) ApplyKey(chord string) error {
	parsed, err := keymap.Parse(chord)
	if err != nil {
		return err
	}
	bound, ok := a.Keymap.Resolve(parsed)
	if !ok {
		return fmt.Errorf("chord %q não tem binding", chord)
	}
	return a.Apply(bound)
}

// ApplyText types text one character at a time.
//
// One character at a time on purpose: typing and pasting are different
// operations, and a case that types must produce the same undo grouping a person
// typing would.
func (a *App) ApplyText(text string) error {
	for _, character := range text {
		if err := a.insertCharacter(character); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) insertCharacter(character rune) error {
	if a.Focus != FocusEditor {
		return nil
	}
	document, err := a.Store.Active()
	if err != nil {
		return err
	}
	if character == '\t' {
		return a.indent(document)
	}
	return document.InsertText(string(character))
}

// indent inserts the configured indentation at the caret.
func (a *App) indent(document *editor.Document) error {
	width := a.Config.TabSizeOrDefault()
	if !a.Config.Editor.InsertSpaces {
		return document.InsertText("\t")
	}
	return document.InsertText(repeat(" ", width))
}

// Apply dispatches one action.
//
// The lookup is the whole dispatch. The domain table says whether the action
// exists and whether it needs a document, so the document is resolved once here
// rather than in every handler — and a handler that does not need one keeps
// working with an empty store, which is what lets someone toggle the tree or
// quit before opening a file.
func (a *App) Apply(bound action.Action) error {
	handler, known := commands[bound]
	if !known {
		return fmt.Errorf("%w: %s", ErrNotImplemented, bound)
	}

	if !handler.NeedsDocument {
		return handler.Apply(a, nil)
	}

	document, err := a.Store.Active()
	if err != nil {
		return err
	}
	return handler.Apply(a, document)
}

// markSaved clears the dirty flag on every open document.
//
// A document with no path cannot be written, and the reference opens the
// save-as browser instead. The headless runner has no browser, so the document
// keeps its dirty flag and the divergence is visible rather than papered over.
func (a *App) markSaved() error {
	document, err := a.Store.Active()
	if err != nil {
		return err
	}
	if _, hasPath := document.Path(); !hasPath {
		a.Status = "save as"
		return nil
	}
	document.MarkSaved()
	return nil
}

// ---------------------------------------------------------------------------
// State dump
// ---------------------------------------------------------------------------

// Dump is the observable state, in the same shape the Rust oracle emits.
//
// Field names and nesting mirror the Rust `StateDump` exactly: the two are
// compared as JSON, so a renamed field is a parity failure rather than a
// silent omission.
type Dump struct {
	Schema           uint32        `json:"schema"`
	Focus            string        `json:"focus"`
	Overlay          OverlayDump   `json:"overlay"`
	Status           *string       `json:"status"`
	Tabs             []TabDump     `json:"tabs"`
	ActiveTab        *uint64       `json:"active_tab"`
	DirtyCount       int           `json:"dirty_count"`
	Document         *DocumentDump `json:"document"`
	Find             FindDump      `json:"find"`
	Split            SplitDump     `json:"split"`
	Tree             *TreeDump     `json:"tree"`
	ShowTree         bool          `json:"show_tree"`
	ShowSCM          bool          `json:"show_scm"`
	SCM              []SCMDump     `json:"scm"`
	Config           ConfigDump    `json:"config"`
	Vim              *string       `json:"vim"`
	Diagnostics      int           `json:"diagnostics"`
	LSPFailures      []string      `json:"lsp_failures"`
	TerminalAttached bool          `json:"terminal_attached"`
	ShouldQuit       bool          `json:"should_quit"`
}

// OverlayDump names the active overlay. The zero value is `none`.
type OverlayDump struct {
	Kind string `json:"overlay"`
}

// TabDump is one open tab.
type TabDump struct {
	ID     uint64  `json:"id"`
	Title  string  `json:"title"`
	Path   *string `json:"path"`
	Dirty  bool    `json:"dirty"`
	Active bool    `json:"active"`
}

// DocumentDump is the active document's state.
type DocumentDump struct {
	ID          uint64        `json:"id"`
	Path        *string       `json:"path"`
	Title       string        `json:"title"`
	Dirty       bool          `json:"dirty"`
	Version     uint64        `json:"version"`
	Lines       int           `json:"lines"`
	Bytes       int           `json:"bytes"`
	Chars       int           `json:"chars"`
	Text        string        `json:"text"`
	Caret       *CaretDump    `json:"caret"`
	Selection   SelectionDump `json:"selection"`
	ExtraCarets []int         `json:"extra_carets"`
	Selected    string        `json:"selected_text"`
	UndoLabels  []string      `json:"undo_labels"`
	RedoLabels  []string      `json:"redo_labels"`
}

// CaretDump is the caret position.
type CaretDump struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// SelectionDump is the selection as byte offsets.
type SelectionDump struct {
	Anchor int  `json:"anchor"`
	Head   int  `json:"head"`
	Empty  bool `json:"empty"`
	Start  int  `json:"start"`
	End    int  `json:"end"`
}

// FindDump is the in-buffer search state.
type FindDump struct {
	Query         string      `json:"query"`
	Replace       string      `json:"replace"`
	Matches       []MatchDump `json:"matches"`
	Current       int         `json:"current"`
	CaseSensitive bool        `json:"case_sensitive"`
	IgnoreAccents bool        `json:"ignore_accents"`
	WholeWord     bool        `json:"whole_word"`
	UseRegex      bool        `json:"use_regex"`
	RegexError    *string     `json:"regex_error"`
	ShowReplace   bool        `json:"show_replace"`
}

// MatchDump is one search hit.
type MatchDump struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// SplitDump is the pane layout.
type SplitDump struct {
	Panes        int      `json:"panes"`
	Orientation  string   `json:"orientation"`
	RatioPercent int      `json:"ratio_percent"`
	Focused      int      `json:"focused"`
	PaneDocs     []uint64 `json:"pane_docs"`
}

// TreeDump is the project tree.
type TreeDump struct {
	Rows     int           `json:"rows"`
	Selected int           `json:"selected"`
	Visible  []TreeRowDump `json:"visible"`
}

// TreeRowDump is one visible tree row.
type TreeRowDump struct {
	Depth    int    `json:"depth"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	IsDir    bool   `json:"is_dir"`
	Expanded bool   `json:"expanded"`
}

// SCMDump is one source-control entry.
type SCMDump struct {
	Badge  string `json:"badge"`
	Status string `json:"status"`
	Path   string `json:"path"`
}

// ConfigDump is the effective configuration.
type ConfigDump struct {
	Theme           string           `json:"theme"`
	Locale          string           `json:"locale"`
	ShowLineNumbers bool             `json:"show_line_numbers"`
	SoftWrap        bool             `json:"soft_wrap"`
	Mouse           bool             `json:"mouse"`
	ModalMode       bool             `json:"modal_mode"`
	TabSize         int              `json:"tab_size"`
	InsertSpaces    bool             `json:"insert_spaces"`
	TreeWidth       int              `json:"tree_width"`
	TreeShowHidden  bool             `json:"tree_show_hidden"`
	Keys            []KeyBindingDump `json:"keys"`
}

// KeyBindingDump is one resolved binding.
type KeyBindingDump struct {
	Chord  string `json:"chord"`
	Action string `json:"action"`
}

// SchemaVersion is the dump shape this package emits.
const SchemaVersion uint32 = 1

// DumpState captures the observable state as JSON.
func (a *App) DumpState() ([]byte, error) {
	dump, err := a.state()
	if err != nil {
		return nil, err
	}
	return json.Marshal(dump)
}

func (a *App) state() (Dump, error) {
	out := Dump{
		Schema:           SchemaVersion,
		Focus:            string(a.Focus),
		Overlay:          OverlayDump{Kind: a.Overlay},
		Tabs:             []TabDump{},
		DirtyCount:       a.Store.DirtyCount(),
		Find:             a.findDump(),
		Split:            SplitDump{Orientation: "vertical", RatioPercent: 50, Panes: 1},
		ShowTree:         a.ShowTree,
		ShowSCM:          a.SCM,
		Tree:             a.treeDump(),
		SCM:              []SCMDump{},
		LSPFailures:      []string{},
		TerminalAttached: false,
		ShouldQuit:       a.Quit,
		Config: ConfigDump{
			Theme:           a.Config.Theme,
			Locale:          a.Config.Locale,
			ShowLineNumbers: a.Config.ShowLineNumbers,
			SoftWrap:        a.Config.SoftWrap,
			Mouse:           a.Config.Mouse,
			ModalMode:       a.Config.Editor.ModalMode,
			TabSize:         a.Config.Editor.TabSize,
			InsertSpaces:    a.Config.Editor.InsertSpaces,
			TreeWidth:       a.Config.Tree.Width,
			TreeShowHidden:  a.Config.Tree.ShowHidden,
			Keys:            a.bindingDumps(),
		},
	}
	if a.Status != "" {
		status := a.Status
		out.Status = &status
	}

	activeID, hasActive := a.Store.ActiveID()
	for _, summary := range a.Store.TabSummaries() {
		tab := TabDump{
			ID:     uint64(summary.ID),
			Title:  summary.Title,
			Dirty:  summary.Dirty,
			Active: hasActive && summary.ID == activeID,
		}
		if summary.Path != "" {
			path := a.relativePath(summary.Path)
			tab.Path = &path
		}
		out.Tabs = append(out.Tabs, tab)
	}
	if hasActive {
		id := uint64(activeID)
		out.ActiveTab = &id
	}

	document, err := a.Store.Active()
	if err != nil {
		return out, nil
	}
	documentDump, err := dumpDocument(document, a.relativePath)
	if err != nil {
		return Dump{}, err
	}
	out.Document = documentDump
	out.Split.PaneDocs = []uint64{uint64(document.ID())}
	return out, nil
}

// findDump renders the in-buffer search state.
func (a *App) findDump() FindDump {
	out := FindDump{
		Query:         a.Find.Query,
		Replace:       a.Find.Replace,
		Matches:       make([]MatchDump, 0, len(a.Find.Matches)),
		Current:       a.Find.Current,
		CaseSensitive: a.Find.Options.CaseSensitive,
		IgnoreAccents: a.Find.Options.IgnoreAccents,
		WholeWord:     a.Find.Options.WholeWord,
		UseRegex:      a.Find.Options.UseRegex,
		ShowReplace:   a.Find.ShowReplace,
	}
	for _, match := range a.Find.Matches {
		out.Matches = append(out.Matches, MatchDump{Start: match.Start, End: match.End})
	}
	if a.Find.RegexError != "" {
		message := a.Find.RegexError
		out.RegexError = &message
	}
	return out
}

// ApplySearch sets the search state and recomputes it over the active buffer.
func (a *App) ApplySearch(query string, options search.Options, replace string) error {
	document, err := a.Store.Active()
	if err != nil {
		return err
	}
	a.Find.Query = query
	a.Find.Options = options
	if replace != "" {
		a.Find.Replace = replace
	}
	a.Find.Recompute(document.Buffer().String())
	return nil
}

// treeDump renders the visible tree rows.
//
// The root row's name is the workspace directory's basename — the one value in
// the dump that would come from the environment rather than the product. It is
// replaced with the placeholder so a fixture recorded in one directory is valid
// in another.
func (a *App) treeDump() *TreeDump {
	if a.Tree == nil {
		return nil
	}
	rows := a.Tree.FlatRows()
	out := &TreeDump{
		Rows:     a.Tree.CountVisibleRows(),
		Selected: a.Tree.SelectedIndex(),
		Visible:  make([]TreeRowDump, 0, len(rows)),
	}
	root := a.Tree.Root()
	for _, row := range rows {
		name := row.Name
		if row.Path == root {
			name = WorkspacePlaceholder
		}
		out.Visible = append(out.Visible, TreeRowDump{
			Depth:    row.Depth,
			Name:     name,
			Path:     a.relativePath(row.Path),
			IsDir:    row.IsDir,
			Expanded: row.Expanded,
		})
	}
	return out
}

// relativePath replaces the workspace prefix with the placeholder the oracle
// uses, and normalises separators so a dump recorded on one platform compares on
// another.
func (a *App) relativePath(path string) string {
	normalized := filepath.ToSlash(path)
	if a.Workspace == "" {
		return normalized
	}
	base := filepath.ToSlash(filepath.Clean(a.Workspace))
	if normalized == base {
		// Trailing slash, matching the oracle: it formats the placeholder plus
		// the relative remainder, and the remainder is empty for the root. Odd
		// enough to be worth a comment — the dump is compared byte for byte.
		return WorkspacePlaceholder + "/"
	}
	if rest, ok := strings.CutPrefix(normalized, base+"/"); ok {
		return WorkspacePlaceholder + "/" + rest
	}
	return normalized
}

func (a *App) bindingDumps() []KeyBindingDump {
	bindings := a.Keymap.Bindings()
	out := make([]KeyBindingDump, 0, len(bindings))
	for _, binding := range bindings {
		out = append(out, KeyBindingDump{
			Chord:  binding.Chord.String(),
			Action: binding.Action.String(),
		})
	}
	return out
}

func dumpDocument(document *editor.Document, relativePath func(string) string) (*DocumentDump, error) {
	caret, err := document.Caret()
	if err != nil {
		return nil, fmt.Errorf("caret: %w", err)
	}
	selection := document.Selection()
	b := document.Buffer()

	out := &DocumentDump{
		ID:      uint64(document.ID()),
		Title:   document.TabTitle(),
		Dirty:   document.IsDirty(),
		Version: document.Version(),
		Lines:   b.LineCount(),
		Bytes:   b.LenBytes(),
		Chars:   b.CharCount(),
		Text:    b.String(),
		Caret:   &CaretDump{Line: caret.Line, Column: caret.Column},
		Selection: SelectionDump{
			Anchor: int(selection.Anchor),
			Head:   int(selection.Head),
			Empty:  selection.IsEmpty(),
			Start:  int(selection.Start()),
			End:    int(selection.End()),
		},
		ExtraCarets: []int{},
		Selected:    document.SelectedText(),
		UndoLabels:  document.UndoHistoryLabels(),
		RedoLabels:  document.RedoHistoryLabels(),
	}
	if path, has := document.Path(); has {
		reported := relativePath(path)
		out.Path = &reported
	}
	for _, extra := range document.ExtraCarets() {
		out.ExtraCarets = append(out.ExtraCarets, int(extra))
	}
	if out.UndoLabels == nil {
		out.UndoLabels = []string{}
	}
	if out.RedoLabels == nil {
		out.RedoLabels = []string{}
	}
	return out, nil
}

// CaretAt is a convenience for tests that need a specific caret.
func CaretAt(document *editor.Document, line, column int) error {
	offset, err := document.Buffer().CaretToByte(buffer.Caret{Line: line, Column: column})
	if err != nil {
		return err
	}
	document.JumpToByte(offset)
	return nil
}

// SortedActions returns every action id, for diagnostics.
func SortedActions() []string {
	all := action.All()
	out := make([]string, 0, len(all))
	for _, a := range all {
		out = append(out, a.String())
	}
	sort.Strings(out)
	return out
}

func repeat(text string, count int) string {
	if count <= 0 {
		return ""
	}
	out := make([]byte, 0, len(text)*count)
	for range count {
		out = append(out, text...)
	}
	return string(out)
}
