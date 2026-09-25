// Package i18n holds the interface translations.
//
// Catalogs are TOML and they are data, not code: a user can add a language by
// dropping a file in a directory, without rebuilding anything. The two shipped
// catalogs live next to this package and are embedded, so the binary always has
// a working fallback even when no directory exists.
package i18n

import (
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

//go:embed catalogs/*.toml
var embedded embed.FS

// Locale identifies a translation.
type Locale string

// Shipped locales.
const (
	PtBR Locale = "pt-BR"
	EnUS Locale = "en-US"
)

// Default is the locale the product starts in.
const Default = PtBR

// Menu is the `[menu]` table: the menu bar's top-level labels.
type Menu struct {
	File string `toml:"file"`
	Edit string `toml:"edit"`
	View string `toml:"view"`
	Go   string `toml:"go"`
	Git  string `toml:"git"`
	Help string `toml:"help"`
}

// FileMenu is the `[file_menu]` table.
type FileMenu struct {
	NewTab     string `toml:"new_tab"`
	OpenFile   string `toml:"open_file"`
	OpenFolder string `toml:"open_folder"`
	Save       string `toml:"save"`
	SaveAs     string `toml:"save_as"`
	SaveAll    string `toml:"save_all"`
	ReloadFile string `toml:"reload_file"`
	Quit       string `toml:"quit"`
}

// EditMenu is the `[edit_menu]` table.
type EditMenu struct {
	Undo          string `toml:"undo"`
	Redo          string `toml:"redo"`
	Cut           string `toml:"cut"`
	Copy          string `toml:"copy"`
	Paste         string `toml:"paste"`
	SelectAll     string `toml:"select_all"`
	Find          string `toml:"find"`
	FindInProject string `toml:"find_in_project"`
	Replace       string `toml:"replace"`
	ToggleComment string `toml:"toggle_comment"`
}

// ViewMenu is the `[view_menu]` table.
type ViewMenu struct {
	CommandPalette  string `toml:"command_palette"`
	ToggleTree      string `toml:"toggle_tree"`
	ToggleTerminal  string `toml:"toggle_terminal"`
	ToggleSCM       string `toml:"toggle_scm"`
	ColorTheme      string `toml:"color_theme"`
	DisplayLanguage string `toml:"display_language"`
	ToggleMouse     string `toml:"toggle_mouse"`
}

// Palette is the `[palette]` table.
type Palette struct {
	ThemePickerTitle  string `toml:"theme_picker_title"`
	ThemePickerHint   string `toml:"theme_picker_hint"`
	LocalePickerTitle string `toml:"locale_picker_title"`
	LocalePickerHint  string `toml:"locale_picker_hint"`
	PaletteHint       string `toml:"palette_hint"`
	ThemeApplied      string `toml:"theme_applied"`
	ThemeCancelled    string `toml:"theme_cancelled"`
	LocaleApplied     string `toml:"locale_applied"`
}

// Status is the `[status]` table.
type Status struct {
	MouseOn  string `toml:"mouse_on"`
	MouseOff string `toml:"mouse_off"`
}

// Catalog is one locale file.
type Catalog struct {
	ID      string   `toml:"id"`
	Name    string   `toml:"name"`
	Menu    Menu     `toml:"menu"`
	File    FileMenu `toml:"file_menu"`
	Edit    EditMenu `toml:"edit_menu"`
	View    ViewMenu `toml:"view_menu"`
	Palette Palette  `toml:"palette"`
	Status  Status   `toml:"status"`
}

// Registry holds the catalogs a session can switch between.
type Registry struct {
	catalogs map[Locale]Catalog
	order    []Locale
}

// NewRegistry returns the shipped catalogs.
func NewRegistry() *Registry {
	registry := &Registry{catalogs: map[Locale]Catalog{}}

	entries, err := embedded.ReadDir("catalogs")
	if err != nil {
		// The embed directive guarantees these files exist; a failure here would
		// be a build problem, not a runtime one.
		return registry
	}
	// The default catalog loads first so the others have something to inherit
	// from, whatever order the embedded files come back in.
	defaultFile := string(Default) + ".toml"
	for _, entry := range entries {
		if entry.Name() != defaultFile {
			continue
		}
		if catalog, err := readEmbedded(entry.Name()); err == nil {
			registry.Register(catalog)
		}
	}
	for _, entry := range entries {
		if entry.Name() == defaultFile {
			continue
		}
		if catalog, err := readEmbedded(entry.Name()); err == nil {
			registry.Register(catalog)
		}
	}
	return registry
}

// readEmbedded parses one embedded catalog.
//
// path.Join, not filepath: embedded paths always use forward slashes, and
// filepath would emit a backslash on Windows.
func readEmbedded(name string) (Catalog, error) {
	data, err := embedded.ReadFile(path.Join("catalogs", name))
	if err != nil {
		return Catalog{}, err
	}
	return parse(data)
}

// Register adds or replaces a catalog.
func (r *Registry) Register(catalog Catalog) {
	id := NormalizeID(catalog.ID)
	catalog = withFallback(catalog, r.Get(Default))
	if _, exists := r.catalogs[id]; !exists {
		r.order = append(r.order, id)
	}
	r.catalogs[id] = catalog
}

// Get returns a catalog, falling back to the default locale.
//
// A missing translation must not leave the interface empty: showing the default
// language is worse than showing the user's, and far better than showing
// nothing.
func (r *Registry) Get(locale Locale) Catalog {
	if catalog, ok := r.catalogs[NormalizeID(string(locale))]; ok {
		return catalog
	}
	if catalog, ok := r.catalogs[Default]; ok {
		return catalog
	}
	return Catalog{}
}

// Available returns the registered locale ids, sorted.
func (r *Registry) Available() []Locale {
	out := make([]Locale, 0, len(r.catalogs))
	for id := range r.catalogs {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// LoadFromDir registers every `.toml` catalog in a directory.
//
// A missing directory is fine; a malformed file is reported with its path, for
// the same reason a malformed theme is: a language that quietly failed to load
// is indistinguishable from one that does not exist.
func (r *Registry) LoadFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("lendo locales em %s: %w", dir, err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".toml") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("lendo locale %s: %w", path, err)
		}
		catalog, err := parse(data)
		if err != nil {
			return fmt.Errorf("locale %s: %w", path, err)
		}
		if strings.TrimSpace(catalog.ID) == "" {
			return fmt.Errorf("locale %s: campo `id` ausente", path)
		}
		r.Register(catalog)
	}
	return nil
}

// withFallback fills the fields a catalog left empty from another catalog.
//
// A partial translation is the normal case: someone adds a language, translates
// the menus, and leaves the rest for later. Without this, everything they have
// not translated yet renders as blanks — an interface that looks broken rather
// than unfinished. The reference leaves them empty; this is a deliberate
// improvement, recorded in the parity ledger.
func withFallback(catalog, base Catalog) Catalog {
	fill := func(field *string, fallback string) {
		if *field == "" {
			*field = fallback
		}
	}

	if catalog.ID == "" {
		catalog.ID = base.ID
	}
	if catalog.Name == "" {
		catalog.Name = base.Name
	}

	fill(&catalog.Menu.File, base.Menu.File)
	fill(&catalog.Menu.Edit, base.Menu.Edit)
	fill(&catalog.Menu.View, base.Menu.View)
	fill(&catalog.Menu.Go, base.Menu.Go)
	fill(&catalog.Menu.Git, base.Menu.Git)
	fill(&catalog.Menu.Help, base.Menu.Help)

	fill(&catalog.File.NewTab, base.File.NewTab)
	fill(&catalog.File.OpenFile, base.File.OpenFile)
	fill(&catalog.File.OpenFolder, base.File.OpenFolder)
	fill(&catalog.File.Save, base.File.Save)
	fill(&catalog.File.SaveAs, base.File.SaveAs)
	fill(&catalog.File.SaveAll, base.File.SaveAll)
	fill(&catalog.File.ReloadFile, base.File.ReloadFile)
	fill(&catalog.File.Quit, base.File.Quit)

	fill(&catalog.Edit.Undo, base.Edit.Undo)
	fill(&catalog.Edit.Redo, base.Edit.Redo)
	fill(&catalog.Edit.Cut, base.Edit.Cut)
	fill(&catalog.Edit.Copy, base.Edit.Copy)
	fill(&catalog.Edit.Paste, base.Edit.Paste)
	fill(&catalog.Edit.SelectAll, base.Edit.SelectAll)
	fill(&catalog.Edit.Find, base.Edit.Find)
	fill(&catalog.Edit.FindInProject, base.Edit.FindInProject)
	fill(&catalog.Edit.Replace, base.Edit.Replace)
	fill(&catalog.Edit.ToggleComment, base.Edit.ToggleComment)

	fill(&catalog.View.CommandPalette, base.View.CommandPalette)
	fill(&catalog.View.ToggleTree, base.View.ToggleTree)
	fill(&catalog.View.ToggleTerminal, base.View.ToggleTerminal)
	fill(&catalog.View.ToggleSCM, base.View.ToggleSCM)
	fill(&catalog.View.ColorTheme, base.View.ColorTheme)
	fill(&catalog.View.DisplayLanguage, base.View.DisplayLanguage)
	fill(&catalog.View.ToggleMouse, base.View.ToggleMouse)

	fill(&catalog.Palette.ThemePickerTitle, base.Palette.ThemePickerTitle)
	fill(&catalog.Palette.ThemePickerHint, base.Palette.ThemePickerHint)
	fill(&catalog.Palette.LocalePickerTitle, base.Palette.LocalePickerTitle)
	fill(&catalog.Palette.LocalePickerHint, base.Palette.LocalePickerHint)
	fill(&catalog.Palette.PaletteHint, base.Palette.PaletteHint)
	fill(&catalog.Palette.ThemeApplied, base.Palette.ThemeApplied)
	fill(&catalog.Palette.ThemeCancelled, base.Palette.ThemeCancelled)
	fill(&catalog.Palette.LocaleApplied, base.Palette.LocaleApplied)

	fill(&catalog.Status.MouseOn, base.Status.MouseOn)
	fill(&catalog.Status.MouseOff, base.Status.MouseOff)

	return catalog
}

func parse(data []byte) (Catalog, error) {
	var catalog Catalog
	if err := toml.Unmarshal(data, &catalog); err != nil {
		return Catalog{}, err
	}
	return catalog, nil
}

// NormalizeID canonicalises a locale id.
//
// `pt_br`, `PT-BR` and `pt-BR` name the same language, and a user writing any of
// them means the same thing.
func NormalizeID(id string) Locale {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return Default
	}
	parts := strings.Split(strings.ReplaceAll(trimmed, "_", "-"), "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		if i == 0 {
			parts[i] = strings.ToLower(part)
			continue
		}
		parts[i] = strings.ToUpper(part)
	}
	return Locale(strings.Join(parts, "-"))
}

// ParseID reads a locale from a configuration value, falling back to the
// default rather than failing: an unknown locale is a typo, not a broken editor.
func ParseID(id string) Locale {
	normalized := NormalizeID(id)
	if normalized == "" {
		return Default
	}
	return normalized
}

// LoadRegistry returns the shipped catalogs plus any the user and the workspace
// provide.
func LoadRegistry(workspace string) *Registry {
	registry := NewRegistry()

	if base, err := os.UserConfigDir(); err == nil {
		_ = registry.LoadFromDir(filepath.Join(base, "oride", "locales"))
	}
	if workspace != "" {
		_ = registry.LoadFromDir(filepath.Join(workspace, ".oride", "locales"))
	}
	return registry
}
