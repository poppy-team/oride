package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// rawFile mirrors the schema with everything optional.
//
// Pointers rather than zero values because "absent" and "false" are different
// answers: a file that omits `mouse` must not turn it off, and a file that sets
// it to false must not be ignored.
type rawFile struct {
	Theme           *string           `toml:"theme"`
	Locale          *string           `toml:"locale"`
	ShowLineNumbers *bool             `toml:"show_line_numbers"`
	SoftWrap        *bool             `toml:"soft_wrap"`
	Mouse           *bool             `toml:"mouse"`
	Editor          *rawEditor        `toml:"editor"`
	UI              *rawUI            `toml:"ui"`
	Syntax          *rawSyntax        `toml:"syntax"`
	Tree            *rawTree          `toml:"tree"`
	Terminal        *rawTerminal      `toml:"terminal"`
	LSP             *rawLSP           `toml:"lsp"`
	Markdown        *rawMarkdown      `toml:"markdown"`
	Keys            map[string]string `toml:"keys"`
	Languages       []rawLanguage     `toml:"languages"`
}

type rawEditor struct {
	TabSize           *int  `toml:"tab_size"`
	InsertSpaces      *bool `toml:"insert_spaces"`
	FormatOnSave      *bool `toml:"format_on_save"`
	UseEditorconfig   *bool `toml:"use_editorconfig"`
	CompletionAuto    *bool `toml:"completion_auto"`
	CompletionMinChar *int  `toml:"completion_min_chars"`
	ModalMode         *bool `toml:"modal_mode"`
}

type rawUI struct {
	Background  *string `toml:"background"`
	Foreground  *string `toml:"foreground"`
	LineNumber  *string `toml:"line_number"`
	StatusBG    *string `toml:"status_bg"`
	StatusFG    *string `toml:"status_fg"`
	StatusDirty *string `toml:"status_dirty"`
	CursorBG    *string `toml:"cursor_bg"`
	CursorFG    *string `toml:"cursor_fg"`
	GutterWidth *int    `toml:"gutter_width"`
}

type rawSyntax struct {
	Comment     *string `toml:"comment"`
	Keyword     *string `toml:"keyword"`
	String      *string `toml:"string"`
	Number      *string `toml:"number"`
	TypeName    *string `toml:"type_name"`
	Function    *string `toml:"function"`
	Operator    *string `toml:"operator"`
	Punctuation *string `toml:"punctuation"`
	Variable    *string `toml:"variable"`
	Constant    *string `toml:"constant"`
	Property    *string `toml:"property"`
	Tag         *string `toml:"tag"`
	Attribute   *string `toml:"attribute"`
	Heading     *string `toml:"heading"`
	Emphasis    *string `toml:"emphasis"`
	Strong      *string `toml:"strong"`
	Link        *string `toml:"link"`
	Code        *string `toml:"code"`
	ListMarker  *string `toml:"list_marker"`
	Quote       *string `toml:"quote"`
}

type rawTree struct {
	Width      *int  `toml:"width"`
	ShowHidden *bool `toml:"show_hidden"`
	GitStatus  *bool `toml:"git_status"`
}

type rawTerminal struct {
	Shell         *string `toml:"shell"`
	DefaultHeight *int    `toml:"default_height"`
}

type rawLSP struct {
	Enabled          *bool               `toml:"enabled"`
	OriscriptCommand []string            `toml:"oriscript_command"`
	Servers          map[string][]string `toml:"servers"`
	TimeoutMS        *int                `toml:"timeout_ms"`
}

type rawMarkdown struct {
	TerminalImages *bool `toml:"terminal_images"`
}

type rawLanguage struct {
	ID              string   `toml:"id"`
	Name            *string  `toml:"name"`
	Extensions      []string `toml:"extensions"`
	Filenames       []string `toml:"filenames"`
	LineComment     *string  `toml:"line_comment"`
	BlockCommentEnd *string  `toml:"block_comment_close"`
	LSPCommand      []string `toml:"lsp_command"`
	TabSize         *int     `toml:"tab_size"`
	InsertSpaces    *bool    `toml:"insert_spaces"`
	SoftWrap        *bool    `toml:"soft_wrap"`
	CompletionWords []string `toml:"completion_words"`
}

// Parse reads a configuration file's bytes over a base configuration.
func Parse(base Config, data []byte) (Config, error) {
	var raw rawFile
	if err := toml.Unmarshal(data, &raw); err != nil {
		return Config{}, fmt.Errorf("configuração inválida: %w", err)
	}
	return Merge(base, raw), nil
}

// Merge applies a parsed file over a base configuration.
func Merge(base Config, raw rawFile) Config {
	out := base

	if raw.Theme != nil {
		out.Theme = *raw.Theme
	}
	if raw.Locale != nil {
		out.Locale = *raw.Locale
	}
	if raw.ShowLineNumbers != nil {
		out.ShowLineNumbers = *raw.ShowLineNumbers
	}
	if raw.SoftWrap != nil {
		out.SoftWrap = *raw.SoftWrap
	}
	if raw.Mouse != nil {
		out.Mouse = *raw.Mouse
	}

	if editor := raw.Editor; editor != nil {
		if editor.TabSize != nil {
			out.Editor.TabSize = clampTabSize(*editor.TabSize)
		}
		if editor.InsertSpaces != nil {
			out.Editor.InsertSpaces = *editor.InsertSpaces
		}
		if editor.FormatOnSave != nil {
			out.Editor.FormatOnSave = *editor.FormatOnSave
		}
		if editor.UseEditorconfig != nil {
			out.Editor.UseEditorconfig = *editor.UseEditorconfig
		}
		if editor.CompletionAuto != nil {
			out.Editor.CompletionAuto = *editor.CompletionAuto
		}
		if editor.CompletionMinChar != nil {
			out.Editor.CompletionMinChar = clampCompletionMinChars(*editor.CompletionMinChar)
		}
		if editor.ModalMode != nil {
			out.Editor.ModalMode = *editor.ModalMode
		}
	}

	if ui := raw.UI; ui != nil {
		mergeUI(&out.UI, ui)
	}
	if syntax := raw.Syntax; syntax != nil {
		mergeSyntax(&out.Syntax, syntax)
	}

	if tree := raw.Tree; tree != nil {
		if tree.Width != nil {
			out.Tree.Width = clampTreeWidth(*tree.Width)
		}
		if tree.ShowHidden != nil {
			out.Tree.ShowHidden = *tree.ShowHidden
		}
		if tree.GitStatus != nil {
			out.Tree.GitStatus = *tree.GitStatus
		}
	}

	if terminal := raw.Terminal; terminal != nil {
		if terminal.Shell != nil {
			out.Terminal.Shell = *terminal.Shell
		}
		if terminal.DefaultHeight != nil {
			out.Terminal.DefaultHeight = clampTerminalHeight(*terminal.DefaultHeight)
		}
	}

	if lsp := raw.LSP; lsp != nil {
		if lsp.Enabled != nil {
			out.LSP.Enabled = *lsp.Enabled
		}
		if len(lsp.OriscriptCommand) > 0 {
			out.LSP.OriscriptCommand = lsp.OriscriptCommand
		}
		if out.LSP.Servers == nil {
			out.LSP.Servers = map[string][]string{}
		}
		for language, command := range lsp.Servers {
			// An empty command would spawn nothing and look like a broken server.
			if len(command) > 0 {
				out.LSP.Servers[language] = command
			}
		}
		if lsp.TimeoutMS != nil {
			out.LSP.TimeoutMS = clampLSPTimeout(*lsp.TimeoutMS)
		}
	}

	if markdown := raw.Markdown; markdown != nil && markdown.TerminalImages != nil {
		out.Markdown.TerminalImages = *markdown.TerminalImages
	}

	// Keys merge individually: a file that rebinds one shortcut keeps the rest.
	if out.Keys == nil {
		out.Keys = map[string]string{}
	}
	for chord, actionID := range raw.Keys {
		out.Keys[chord] = actionID
	}

	for _, language := range raw.Languages {
		out.Languages = upsertLanguage(out.Languages, language)
	}

	return out
}

func mergeUI(dst *UIConfig, src *rawUI) {
	set := func(field *string, value *string) {
		if value != nil {
			*field = *value
		}
	}
	set(&dst.Background, src.Background)
	set(&dst.Foreground, src.Foreground)
	set(&dst.LineNumber, src.LineNumber)
	set(&dst.StatusBG, src.StatusBG)
	set(&dst.StatusFG, src.StatusFG)
	set(&dst.StatusDirty, src.StatusDirty)
	set(&dst.CursorBG, src.CursorBG)
	set(&dst.CursorFG, src.CursorFG)
	if src.GutterWidth != nil {
		dst.GutterWidth = clampGutterWidth(*src.GutterWidth)
	}
}

func mergeSyntax(dst *SyntaxColors, src *rawSyntax) {
	set := func(field *string, value *string) {
		if value != nil {
			*field = *value
		}
	}
	set(&dst.Comment, src.Comment)
	set(&dst.Keyword, src.Keyword)
	set(&dst.String, src.String)
	set(&dst.Number, src.Number)
	set(&dst.TypeName, src.TypeName)
	set(&dst.Function, src.Function)
	set(&dst.Operator, src.Operator)
	set(&dst.Punctuation, src.Punctuation)
	set(&dst.Variable, src.Variable)
	set(&dst.Constant, src.Constant)
	set(&dst.Property, src.Property)
	set(&dst.Tag, src.Tag)
	set(&dst.Attribute, src.Attribute)
	set(&dst.Heading, src.Heading)
	set(&dst.Emphasis, src.Emphasis)
	set(&dst.Strong, src.Strong)
	set(&dst.Link, src.Link)
	set(&dst.Code, src.Code)
	set(&dst.ListMarker, src.ListMarker)
	set(&dst.Quote, src.Quote)
}

func upsertLanguage(languages []LanguageConfig, raw rawLanguage) []LanguageConfig {
	merged := LanguageConfig{
		ID:              raw.ID,
		Extensions:      raw.Extensions,
		Filenames:       raw.Filenames,
		LSPCommand:      raw.LSPCommand,
		CompletionWords: raw.CompletionWords,
	}
	if raw.Name != nil {
		merged.Name = *raw.Name
	}
	if raw.LineComment != nil {
		merged.LineComment = *raw.LineComment
	}
	if raw.BlockCommentEnd != nil {
		merged.BlockCommentEnd = *raw.BlockCommentEnd
	}
	if raw.TabSize != nil {
		merged.TabSize = clampTabSize(*raw.TabSize)
	}
	if raw.InsertSpaces != nil {
		merged.InsertSpaces = *raw.InsertSpaces
	}
	if raw.SoftWrap != nil {
		merged.SoftWrap = *raw.SoftWrap
	}

	for i, existing := range languages {
		if equalFold(existing.ID, raw.ID) {
			languages[i] = merged
			return languages
		}
	}
	return append(languages, merged)
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range len(a) {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// LoadMerged reads the configuration layers for a workspace.
//
// Order: defaults, then the user's file, then the nearest project file. A
// missing file is not an error — most users have neither — but an unreadable or
// malformed one is, because silently ignoring it would make the editor behave
// unlike the file the user wrote.
func LoadMerged(workspace string) (Config, error) {
	out := Default()

	if path, ok := UserConfigPath(); ok {
		merged, err := loadFile(out, path)
		if err != nil {
			return Config{}, err
		}
		out = merged
	}

	if workspace != "" {
		if path, ok := ProjectConfigPath(workspace); ok {
			merged, err := loadFile(out, path)
			if err != nil {
				return Config{}, err
			}
			out = merged
		}
	}

	return out, nil
}

func loadFile(base Config, path string) (Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return base, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("lendo %s: %w", path, err)
	}
	merged, err := Parse(base, data)
	if err != nil {
		return Config{}, fmt.Errorf("%s: %w", path, err)
	}
	return merged, nil
}

// UserConfigPath returns the user's configuration file.
func UserConfigPath() (string, bool) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", false
	}
	return filepath.Join(base, "oride", "config.toml"), true
}

// ProjectConfigPath walks up from a workspace looking for a project file.
//
// Nearest wins, and the walk is bounded: an unbounded search would reach the
// filesystem root and pick up a configuration nobody intended for this project.
func ProjectConfigPath(workspace string) (string, bool) {
	const maxDepth = 32

	dir := workspace
	for range maxDepth {
		candidate := filepath.Join(dir, ".oride", "config.toml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}
