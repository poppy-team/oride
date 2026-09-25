package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestBuiltInThemeNames pins the shipped set. A theme disappearing from the
// picker is a product change, and the list is the cheapest place to notice.
func TestBuiltInThemeNames(t *testing.T) {
	want := []string{
		"Default Dark", "Dracula", "Nord", "One Dark", "Tokyo Night",
		"Catppuccin Mocha", "Monokai", "Solarized Light", "Solarized Dark", "GitHub Dark",
	}

	registry := NewThemeRegistry()
	got := registry.ListNames()
	if len(got) != len(want) {
		t.Fatalf("temas = %v, esperado %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("tema %d = %q, esperado %q", i, got[i], want[i])
		}
	}
}

// TestEveryBuiltInThemeResolvesEveryColour: a theme that inherits a section must
// end up with the product defaults, not with empty strings. An unresolved colour
// renders as "no colour" rather than as an error, so nothing else would catch it.
func TestEveryBuiltInThemeResolvesEveryColour(t *testing.T) {
	for _, theme := range BuiltInThemes() {
		if theme.Name == "" {
			t.Error("tema sem nome")
		}
		if theme.UI.Background == "" || theme.UI.Foreground == "" {
			t.Errorf("tema %q sem cor de superfície ou texto", theme.Name)
		}
		if theme.UI.GutterWidth < 1 {
			t.Errorf("tema %q com gutter_width %d", theme.Name, theme.UI.GutterWidth)
		}
		if theme.Syntax.Comment == "" || theme.Syntax.Keyword == "" || theme.Syntax.String == "" {
			t.Errorf("tema %q com cores de sintaxe não resolvidas", theme.Name)
		}
	}
}

func TestNormalizeThemeName(t *testing.T) {
	cases := map[string]string{
		"One Dark":         "one-dark",
		"one_dark":         "one-dark",
		"ONE-DARK":         "one-dark",
		"  Tokyo Night  ":  "tokyo-night",
		"Catppuccin Mocha": "catppuccin-mocha",
		"solarized_light":  "solarized-light",
	}
	for input, want := range cases {
		if got := NormalizeThemeName(input); got != want {
			t.Errorf("NormalizeThemeName(%q) = %q, esperado %q", input, got, want)
		}
	}
}

func TestGetAcceptsAnySpellingOfTheName(t *testing.T) {
	registry := NewThemeRegistry()
	for _, spelling := range []string{"Dracula", "dracula", "DRACULA", " dracula "} {
		if _, ok := registry.Get(spelling); !ok {
			t.Errorf("Get(%q) não encontrou o tema", spelling)
		}
	}
	if _, ok := registry.Get("tema-que-nao-existe"); ok {
		t.Error("um tema inexistente foi encontrado")
	}
}

// TestThemeFilesLoadAndOverride pins the on-disk shape and the merge: a file
// that sets one colour keeps the rest of the palette.
func TestThemeFilesLoadAndOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "meu-tema.toml")
	content := `
name = "Meu Tema"
is_dark = false

[ui]
background = "#101010"

[syntax]
keyword = "#ff0000"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("escrevendo tema: %v", err)
	}

	registry := NewThemeRegistry()
	if err := registry.LoadFromDir(dir); err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}

	theme, ok := registry.Get("Meu Tema")
	if !ok {
		t.Fatal("o tema do arquivo não foi registrado")
	}
	if theme.IsDark {
		t.Error("is_dark = false não foi respeitado")
	}
	if theme.UI.Background != "#101010" {
		t.Errorf("background = %q", theme.UI.Background)
	}
	if theme.Syntax.Keyword != "#ff0000" {
		t.Errorf("keyword = %q", theme.Syntax.Keyword)
	}
	// Untouched fields come from the defaults.
	if theme.UI.Foreground == "" || theme.Syntax.String == "" {
		t.Error("cores não declaradas ficaram vazias em vez de herdadas")
	}
}

func TestMissingThemeDirectoryIsNotAnError(t *testing.T) {
	registry := NewThemeRegistry()
	if err := registry.LoadFromDir(filepath.Join(t.TempDir(), "nao-existe")); err != nil {
		t.Errorf("diretório ausente virou erro: %v", err)
	}
}

// TestMalformedThemeFileNamesTheFile: a theme that silently fails to load looks
// exactly like a theme that does not exist.
func TestMalformedThemeFileNamesTheFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "quebrado.toml")
	if err := os.WriteFile(path, []byte("name = \n[ui\n"), 0o644); err != nil {
		t.Fatalf("escrevendo: %v", err)
	}

	err := NewThemeRegistry().LoadFromDir(dir)
	if err == nil {
		t.Fatal("arquivo malformado foi aceito")
	}
	if !contains(err.Error(), "quebrado.toml") {
		t.Errorf("erro %q não nomeia o arquivo", err)
	}
}

func TestThemeWithoutANameIsRejected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sem-nome.toml")
	if err := os.WriteFile(path, []byte("[ui]\nbackground = \"#000000\"\n"), 0o644); err != nil {
		t.Fatalf("escrevendo: %v", err)
	}

	err := NewThemeRegistry().LoadFromDir(dir)
	if err == nil {
		t.Fatal("tema sem nome foi aceito")
	}
	if !contains(err.Error(), "name") {
		t.Errorf("erro %q não menciona o campo ausente", err)
	}
}

func TestRegisterReplacesWithoutDuplicating(t *testing.T) {
	registry := NewThemeRegistry()
	before := registry.Len()

	registry.Register(ThemeDefinition{Name: "Dracula", IsDark: false})
	if registry.Len() != before {
		t.Errorf("registrar de novo mudou a contagem: %d → %d", before, registry.Len())
	}
	theme, _ := registry.Get("Dracula")
	if theme.IsDark {
		t.Error("o registro novo não substituiu o antigo")
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
