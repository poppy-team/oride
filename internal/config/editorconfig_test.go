package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeEditorconfig(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".editorconfig"), []byte(content), 0o644); err != nil {
		t.Fatalf("escrevendo .editorconfig: %v", err)
	}
}

// TestReadsIndentFromEditorConfig mirrors the reference's own example: a global
// section sets spaces of size 2, a per-extension section raises it to 4.
func TestReadsIndentFromEditorConfig(t *testing.T) {
	dir := t.TempDir()
	writeEditorconfig(t, dir, "root = true\n\n[*]\nindent_style = space\nindent_size = 2\n\n[*.rs]\nindent_size = 4\n")

	file := filepath.Join(dir, "main.rs")
	if err := os.WriteFile(file, []byte("fn main() {}\n"), 0o644); err != nil {
		t.Fatalf("escrevendo arquivo: %v", err)
	}

	got := ResolveIndentForFile(file, EditorIndent{TabSize: 8, InsertSpaces: false})
	if !got.InsertSpaces {
		t.Error("indent_style = space não foi aplicado")
	}
	if got.TabSize != 4 {
		t.Errorf("tab_size = %d, esperado 4 (a seção [*.rs] vence [*])", got.TabSize)
	}
}

func TestIndentStyleTab(t *testing.T) {
	dir := t.TempDir()
	writeEditorconfig(t, dir, "[*]\nindent_style = tab\n")

	got := ResolveIndentForFile(filepath.Join(dir, "a.txt"), EditorIndent{TabSize: 4, InsertSpaces: true})
	if got.InsertSpaces {
		t.Error("indent_style = tab não foi aplicado")
	}
}

// TestTabWidthOnlyAppliesToTabs: tab_width describes a literal tab, so it must
// not resize indentation that is made of spaces.
func TestTabWidthOnlyAppliesToTabs(t *testing.T) {
	dir := t.TempDir()

	writeEditorconfig(t, dir, "[*]\nindent_style = space\ntab_width = 8\n")
	withSpaces := ResolveIndentForFile(filepath.Join(dir, "a.txt"), EditorIndent{TabSize: 4, InsertSpaces: true})
	if withSpaces.TabSize != 4 {
		t.Errorf("tab_width mexeu no tamanho com espaços: %d", withSpaces.TabSize)
	}

	writeEditorconfig(t, dir, "[*]\nindent_style = tab\ntab_width = 8\n")
	withTabs := ResolveIndentForFile(filepath.Join(dir, "a.txt"), EditorIndent{TabSize: 4, InsertSpaces: false})
	if withTabs.TabSize != 8 {
		t.Errorf("tab_width não foi aplicado com tabs: %d", withTabs.TabSize)
	}
}

func TestIndentSizeTabKeepsTabWidth(t *testing.T) {
	dir := t.TempDir()
	writeEditorconfig(t, dir, "[*]\nindent_style = tab\nindent_size = tab\ntab_width = 6\n")

	got := ResolveIndentForFile(filepath.Join(dir, "a.txt"), EditorIndent{TabSize: 4, InsertSpaces: false})
	if got.TabSize != 6 {
		t.Errorf("tab_size = %d, esperado 6 (indent_size = tab delega para tab_width)", got.TabSize)
	}
}

// TestRootTrueStopsTheWalk is what the setting is for: a project declares that
// nothing above it applies.
func TestRootTrueStopsTheWalk(t *testing.T) {
	parent := t.TempDir()
	writeEditorconfig(t, parent, "[*]\nindent_size = 8\n")

	child := filepath.Join(parent, "projeto")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("criando subdiretório: %v", err)
	}
	writeEditorconfig(t, child, "root = true\n[*]\nindent_size = 2\n")

	got := ResolveIndentForFile(filepath.Join(child, "a.txt"), EditorIndent{TabSize: 4})
	if got.TabSize != 2 {
		t.Errorf("tab_size = %d, esperado 2 — o `root = true` deveria parar a subida", got.TabSize)
	}
}

// TestWithoutRootTrueTheNearestWins: without it, a closer file overrides a
// farther one, which is the other half of the same rule.
func TestWithoutRootTrueTheNearestWins(t *testing.T) {
	parent := t.TempDir()
	writeEditorconfig(t, parent, "[*]\nindent_size = 8\n")

	child := filepath.Join(parent, "projeto")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("criando subdiretório: %v", err)
	}
	writeEditorconfig(t, child, "[*]\nindent_size = 2\n")

	got := ResolveIndentForFile(filepath.Join(child, "a.txt"), EditorIndent{TabSize: 4})
	if got.TabSize != 2 {
		t.Errorf("tab_size = %d, esperado 2", got.TabSize)
	}
}

func TestSectionMatching(t *testing.T) {
	cases := []struct {
		pattern  string
		ext      string
		filename string
		want     bool
	}{
		{"*", "rs", "main.rs", true},
		{"*.rs", "rs", "main.rs", true},
		{"*.rs", "toml", "config.toml", false},
		{"**/*.rs", "rs", "src/main.rs", true},
		{"**/*.rs", "toml", "Cargo.toml", false},
		{"Cargo.toml", "toml", "cargo.toml", true},
		{"Makefile", "", "makefile", true},
		{"*.RS", "rs", "main.rs", true},
	}
	for _, tc := range cases {
		if got := sectionMatches(tc.pattern, tc.ext, tc.filename); got != tc.want {
			t.Errorf("sectionMatches(%q, ext=%q, name=%q) = %v, esperado %v",
				tc.pattern, tc.ext, tc.filename, got, tc.want)
		}
	}
}

func TestCommentsAndBlankLinesAreIgnored(t *testing.T) {
	dir := t.TempDir()
	writeEditorconfig(t, dir, "# comentário\n; outro\n\n[*]\n\nindent_size = 3\n")

	got := ResolveIndentForFile(filepath.Join(dir, "a.txt"), EditorIndent{TabSize: 4})
	if got.TabSize != 3 {
		t.Errorf("tab_size = %d, esperado 3", got.TabSize)
	}
}

func TestPropertiesBeforeAnySectionAreIgnored(t *testing.T) {
	dir := t.TempDir()
	writeEditorconfig(t, dir, "indent_size = 9\n[*]\nindent_size = 2\n")

	got := ResolveIndentForFile(filepath.Join(dir, "a.txt"), EditorIndent{TabSize: 4})
	if got.TabSize != 2 {
		t.Errorf("tab_size = %d, esperado 2 — a propriedade global não deveria valer", got.TabSize)
	}
}

func TestMissingEditorconfigKeepsTheFallback(t *testing.T) {
	got := ResolveIndentForFile(filepath.Join(t.TempDir(), "a.txt"), EditorIndent{TabSize: 7, InsertSpaces: true})
	if got.TabSize != 7 || !got.InsertSpaces {
		t.Errorf("fallback não foi preservado: %+v", got)
	}
}

func TestZeroOrNegativeSizesAreIgnored(t *testing.T) {
	dir := t.TempDir()
	writeEditorconfig(t, dir, "[*]\nindent_size = 0\n")

	got := ResolveIndentForFile(filepath.Join(dir, "a.txt"), EditorIndent{TabSize: 4})
	if got.TabSize != 4 {
		t.Errorf("tab_size = %d, esperado 4 — zero não é um tamanho válido", got.TabSize)
	}
}
