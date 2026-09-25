// Package config holds the effective product configuration.
//
// The file format is TOML and the configuration layers: built-in defaults, then
// the user's file, then the project's. Every layer is partial — a file that sets
// one key does not reset the rest — which is why the raw structs below mirror
// the schema with pointers and the merge is field by field rather than a decode
// into the final type.
package config

// Config is the effective configuration after every layer has been applied.
type Config struct {
	Theme           string
	Locale          string
	ShowLineNumbers bool
	SoftWrap        bool
	// Mouse is off by default. With capture on, the terminal stops being the
	// user's: native selection and copy stop working. The keyboard is the
	// complete path; the mouse is optional acceleration.
	Mouse     bool
	Editor    EditorConfig
	Tree      TreeConfig
	Terminal  TerminalConfig
	LSP       LSPConfig
	Markdown  MarkdownConfig
	UI        UIConfig
	Syntax    SyntaxColors
	Languages []LanguageConfig
	Keys      map[string]string
}

// EditorConfig is the `[editor]` table.
type EditorConfig struct {
	TabSize           int
	InsertSpaces      bool
	FormatOnSave      bool
	UseEditorconfig   bool
	CompletionAuto    bool
	CompletionMinChar int
	ModalMode         bool
}

// TreeConfig is the `[tree]` table.
type TreeConfig struct {
	Width      int
	ShowHidden bool
	GitStatus  bool
}

// TerminalConfig is the `[terminal]` table.
type TerminalConfig struct {
	Shell         string
	DefaultHeight int
}

// LSPConfig is the `[lsp]` table.
type LSPConfig struct {
	Enabled          bool
	OriscriptCommand []string
	Servers          map[string][]string
	TimeoutMS        int
}

// MarkdownConfig is the `[markdown]` table.
type MarkdownConfig struct {
	TerminalImages bool
}

// UIConfig is the `[ui]` table: the surface colours.
type UIConfig struct {
	Background  string
	Foreground  string
	LineNumber  string
	StatusBG    string
	StatusFG    string
	StatusDirty string
	CursorBG    string
	CursorFG    string
	GutterWidth int
}

// SyntaxColors is the `[syntax]` table.
type SyntaxColors struct {
	Comment     string
	Keyword     string
	String      string
	Number      string
	TypeName    string
	Function    string
	Operator    string
	Punctuation string
	Variable    string
	Constant    string
	Property    string
	Tag         string
	Attribute   string
	Heading     string
	Emphasis    string
	Strong      string
	Link        string
	Code        string
	ListMarker  string
	Quote       string
}

// LanguageConfig is one `[[languages]]` entry.
type LanguageConfig struct {
	ID              string
	Name            string
	Extensions      []string
	Filenames       []string
	LineComment     string
	BlockCommentEnd string
	LSPCommand      []string
	TabSize         int
	InsertSpaces    bool
	SoftWrap        bool
	CompletionWords []string
}

// Default returns the shipped configuration.
//
// Every clamp applied here is also applied when merging a file, so a value that
// arrives from disk cannot produce a state the editor cannot render — a tab size
// of zero would divide by zero on the first indent.
func Default() Config {
	return Config{
		Theme:           "default",
		Locale:          "pt-BR",
		ShowLineNumbers: true,
		SoftWrap:        false,
		Mouse:           false,
		Editor: EditorConfig{
			TabSize:           4,
			InsertSpaces:      true,
			FormatOnSave:      false,
			UseEditorconfig:   true,
			CompletionAuto:    true,
			CompletionMinChar: 2,
			ModalMode:         false,
		},
		Tree: TreeConfig{
			Width:      28,
			ShowHidden: false,
			GitStatus:  true,
		},
		Terminal: TerminalConfig{
			Shell:         "",
			DefaultHeight: 10,
		},
		LSP: LSPConfig{
			Enabled:          true,
			OriscriptCommand: []string{"oriscript", "lsp"},
			Servers:          map[string][]string{},
			TimeoutMS:        10_000,
		},
		Markdown: MarkdownConfig{TerminalImages: false},
		UI: UIConfig{
			Background:  "reset",
			Foreground:  "reset",
			LineNumber:  "darkgray",
			StatusBG:    "darkgray",
			StatusFG:    "white",
			StatusDirty: "yellow",
			CursorBG:    "white",
			CursorFG:    "black",
			GutterWidth: 5,
		},
		Syntax:    defaultSyntaxColors(),
		Keys:      DefaultKeyBindings(),
		Languages: nil,
	}
}

func defaultSyntaxColors() SyntaxColors {
	return SyntaxColors{
		Comment:     "darkgray",
		Keyword:     "magenta",
		String:      "green",
		Number:      "yellow",
		TypeName:    "cyan",
		Function:    "blue",
		Operator:    "reset",
		Punctuation: "darkgray",
		Variable:    "reset",
		Constant:    "yellow",
		Property:    "cyan",
		Tag:         "red",
		Attribute:   "yellow",
		Heading:     "magenta",
		Emphasis:    "cyan",
		Strong:      "yellow",
		Link:        "blue",
		Code:        "green",
		ListMarker:  "yellow",
		Quote:       "darkgray",
	}
}

// TabSizeOrDefault returns the configured tab size, never below one.
func (c Config) TabSizeOrDefault() int {
	if c.Editor.TabSize < 1 {
		return 1
	}
	return c.Editor.TabSize
}

// clampTabSize keeps a tab size usable.
func clampTabSize(value int) int {
	if value < 1 {
		return 1
	}
	return value
}

// clampCompletionMinChars keeps the completion threshold at one or more.
func clampCompletionMinChars(value int) int {
	if value < 1 {
		return 1
	}
	return value
}

// clampTreeWidth keeps the tree wide enough to render a name.
func clampTreeWidth(value int) int {
	if value < 8 {
		return 8
	}
	return value
}

// clampTerminalHeight keeps the terminal panel usable.
func clampTerminalHeight(value int) int {
	if value < 3 {
		return 3
	}
	return value
}

// clampGutterWidth keeps the line-number column at least one cell.
func clampGutterWidth(value int) int {
	if value < 1 {
		return 1
	}
	return value
}

// clampLSPTimeout keeps a server request from giving up instantly.
func clampLSPTimeout(value int) int {
	if value < 500 {
		return 500
	}
	return value
}
