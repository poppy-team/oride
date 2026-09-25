package i18n

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestShippedCatalogsLoad: the embedded catalogs are the fallback the binary
// always has, so a build that loses them would leave the interface empty.
func TestShippedCatalogsLoad(t *testing.T) {
	registry := NewRegistry()

	for _, id := range []Locale{PtBR, EnUS} {
		catalog := registry.Get(id)
		if catalog.ID == "" {
			t.Errorf("catálogo %q não carregou", id)
			continue
		}
		if catalog.Name == "" {
			t.Errorf("catálogo %q sem nome de exibição", id)
		}
	}
}

// TestEveryCatalogFieldIsFilled finds a translation that was added to one
// catalog and forgotten in the other — the failure mode of any catalog pair.
func TestEveryCatalogFieldIsFilled(t *testing.T) {
	registry := NewRegistry()

	type field struct {
		locale string
		value  string
	}
	values := map[string][]field{}

	collect := func(locale Locale, catalog Catalog) {
		add := func(label, value string) {
			values[label] = append(values[label], field{string(locale), value})
		}
		add("menu.file", catalog.Menu.File)
		add("menu.edit", catalog.Menu.Edit)
		add("menu.view", catalog.Menu.View)
		add("menu.go", catalog.Menu.Go)
		add("menu.git", catalog.Menu.Git)
		add("menu.help", catalog.Menu.Help)

		add("file_menu.new_tab", catalog.File.NewTab)
		add("file_menu.open_file", catalog.File.OpenFile)
		add("file_menu.open_folder", catalog.File.OpenFolder)
		add("file_menu.save", catalog.File.Save)
		add("file_menu.save_as", catalog.File.SaveAs)
		add("file_menu.save_all", catalog.File.SaveAll)
		add("file_menu.reload_file", catalog.File.ReloadFile)
		add("file_menu.quit", catalog.File.Quit)

		add("edit_menu.undo", catalog.Edit.Undo)
		add("edit_menu.redo", catalog.Edit.Redo)
		add("edit_menu.cut", catalog.Edit.Cut)
		add("edit_menu.copy", catalog.Edit.Copy)
		add("edit_menu.paste", catalog.Edit.Paste)
		add("edit_menu.select_all", catalog.Edit.SelectAll)
		add("edit_menu.find", catalog.Edit.Find)
		add("edit_menu.find_in_project", catalog.Edit.FindInProject)
		add("edit_menu.replace", catalog.Edit.Replace)
		add("edit_menu.toggle_comment", catalog.Edit.ToggleComment)

		add("view_menu.command_palette", catalog.View.CommandPalette)
		add("view_menu.toggle_tree", catalog.View.ToggleTree)
		add("view_menu.toggle_terminal", catalog.View.ToggleTerminal)
		add("view_menu.toggle_scm", catalog.View.ToggleSCM)
		add("view_menu.color_theme", catalog.View.ColorTheme)
		add("view_menu.display_language", catalog.View.DisplayLanguage)
		add("view_menu.toggle_mouse", catalog.View.ToggleMouse)

		add("palette.theme_picker_title", catalog.Palette.ThemePickerTitle)
		add("palette.theme_picker_hint", catalog.Palette.ThemePickerHint)
		add("palette.locale_picker_title", catalog.Palette.LocalePickerTitle)
		add("palette.locale_picker_hint", catalog.Palette.LocalePickerHint)
		add("palette.palette_hint", catalog.Palette.PaletteHint)
		add("palette.theme_applied", catalog.Palette.ThemeApplied)
		add("palette.theme_cancelled", catalog.Palette.ThemeCancelled)
		add("palette.locale_applied", catalog.Palette.LocaleApplied)

		add("status.mouse_on", catalog.Status.MouseOn)
		add("status.mouse_off", catalog.Status.MouseOff)
	}

	collect(PtBR, registry.Get(PtBR))
	collect(EnUS, registry.Get(EnUS))

	for label, entries := range values {
		for _, entry := range entries {
			if entry.value == "" {
				t.Errorf("%s vazio no locale %s", label, entry.locale)
			}
		}
		if entries[0].value == entries[1].value && label != "view_menu.display_language" {
			// A few labels are deliberately identical across languages — a
			// proper noun, for instance. Everything else should have been
			// translated.
			t.Logf("%s é idêntico nos dois locales: %q", label, entries[0].value)
		}
	}
}

func TestNormalizeID(t *testing.T) {
	cases := map[string]Locale{
		"pt-BR":   "pt-BR",
		"pt_br":   "pt-BR",
		"PT-BR":   "pt-BR",
		"  en-us": "en-US",
		"en_US":   "en-US",
		"pt":      "pt",
		"":        Default,
	}
	for input, want := range cases {
		if got := NormalizeID(input); got != want {
			t.Errorf("NormalizeID(%q) = %q, esperado %q", input, got, want)
		}
	}
}

// TestUnknownLocaleFallsBack: an interface with no strings at all is worse than
// one in the wrong language.
func TestUnknownLocaleFallsBack(t *testing.T) {
	registry := NewRegistry()
	catalog := registry.Get("xx-YY")
	if catalog.ID == "" {
		t.Fatal("locale desconhecido não caiu no padrão")
	}
	if catalog.ID != string(Default) {
		t.Errorf("caiu em %q, esperado %q", catalog.ID, Default)
	}
}

func TestAvailableListsBothCatalogs(t *testing.T) {
	available := NewRegistry().Available()
	if len(available) < 2 {
		t.Fatalf("locales disponíveis = %v", available)
	}
	found := map[Locale]bool{}
	for _, id := range available {
		found[id] = true
	}
	for _, want := range []Locale{PtBR, EnUS} {
		if !found[want] {
			t.Errorf("%q ausente de %v", want, available)
		}
	}
}

func TestCustomCatalogLoadsAndOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "es-ES.toml")
	content := "id = \"es-ES\"\nname = \"Español\"\n\n[menu]\nfile = \"Archivo\"\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("escrevendo: %v", err)
	}

	registry := NewRegistry()
	if err := registry.LoadFromDir(dir); err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}

	catalog := registry.Get("es-ES")
	if catalog.Name != "Español" {
		t.Errorf("nome = %q", catalog.Name)
	}
	if catalog.Menu.File != "Archivo" {
		t.Errorf("menu.file = %q", catalog.Menu.File)
	}
	// Fields the file omits fall back to the default catalog rather than to
	// empty strings, so a partial translation is still usable.
	if catalog.Menu.Edit == "" {
		t.Error("campo não traduzido ficou vazio em vez de herdar")
	}
}

func TestMissingLocaleDirectoryIsNotAnError(t *testing.T) {
	if err := NewRegistry().LoadFromDir(filepath.Join(t.TempDir(), "nao-existe")); err != nil {
		t.Errorf("diretório ausente virou erro: %v", err)
	}
}

func TestMalformedCatalogNamesTheFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "quebrado.toml")
	if err := os.WriteFile(path, []byte("id = \n[menu\n"), 0o644); err != nil {
		t.Fatalf("escrevendo: %v", err)
	}

	err := NewRegistry().LoadFromDir(dir)
	if err == nil {
		t.Fatal("catálogo malformado foi aceito")
	}
	if !strings.Contains(err.Error(), "quebrado.toml") {
		t.Errorf("erro %q não nomeia o arquivo", err)
	}
}
