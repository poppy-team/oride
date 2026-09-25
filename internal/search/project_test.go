package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeProject(t *testing.T, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	for name, content := range files {
		target := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatalf("criando %s: %v", filepath.Dir(target), err)
		}
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			t.Fatalf("escrevendo %s: %v", target, err)
		}
	}
	return root
}

func paths(hits []Hit) []string {
	out := make([]string, 0, len(hits))
	for _, hit := range hits {
		out = append(out, hit.Path)
	}
	return out
}

// TestFindProjectLocatesMatches across the two backends. The assertions are
// intentionally independent of which one answered, because both must agree on
// what a match is.
func TestFindProjectLocatesMatches(t *testing.T) {
	root := makeProject(t, map[string]string{
		"a.txt":     "alfa\nbeta\n",
		"b.txt":     "gama\n",
		"sub/c.txt": "beta de novo\n",
	})

	hits, _, err := FindProject(root, Query{Text: "beta"})
	if err != nil {
		t.Fatalf("FindProject: %v", err)
	}
	if len(hits) != 2 {
		t.Fatalf("hits = %+v", hits)
	}
	for _, hit := range hits {
		if hit.Text != "beta" && hit.Text != "beta de novo" {
			t.Errorf("linha = %q", hit.Text)
		}
		if hit.Line < 1 {
			t.Errorf("linha = %d, esperado 1-based", hit.Line)
		}
		if hit.Column < 1 {
			t.Errorf("coluna = %d, esperado 1-based", hit.Column)
		}
	}
}

func TestFindProjectIsCaseInsensitiveByDefault(t *testing.T) {
	root := makeProject(t, map[string]string{"a.txt": "Beta\nbeta\nBETA\n"})

	insensitive, _, err := FindProject(root, Query{Text: "beta"})
	if err != nil {
		t.Fatalf("FindProject: %v", err)
	}
	if len(insensitive) != 3 {
		t.Errorf("sem case sensitive: %d hits, esperado 3", len(insensitive))
	}

	sensitive, _, err := FindProject(root, Query{Text: "beta", CaseSensitive: true})
	if err != nil {
		t.Fatalf("FindProject: %v", err)
	}
	if len(sensitive) != 1 {
		t.Errorf("com case sensitive: %d hits, esperado 1", len(sensitive))
	}
}

// TestFindProjectTreatsTheQueryLiterallyByDefault: a query with a regex
// metacharacter must find that character, not act as a pattern.
func TestFindProjectTreatsTheQueryLiterallyByDefault(t *testing.T) {
	root := makeProject(t, map[string]string{"a.txt": "total = 42\na1 = b2\n"})

	literal, _, err := FindProject(root, Query{Text: "= 42"})
	if err != nil {
		t.Fatalf("FindProject: %v", err)
	}
	if len(literal) != 1 {
		t.Fatalf("literal: %+v", literal)
	}

	// As a pattern, \d+ matches digits — but a hit is a line, not an
	// occurrence, so `a1 = b2` counts once despite holding two numbers.
	pattern, _, err := FindProject(root, Query{Text: `\d+`, UseRegex: true})
	if err != nil {
		t.Fatalf("FindProject: %v", err)
	}
	if len(pattern) != 2 {
		t.Errorf("regex: %d hits, esperado 2 (uma por linha)", len(pattern))
	}
}

// TestFindProjectSkipsBuildOutput: nobody wants a search over compiled output.
// TestFindProjectSkipsIgnoredBuildOutput: inside a real repository — the normal
// case — both backends agree, because `rg` honours `.gitignore` and the fallback's
// list happens to cover the same entries.
//
// The `.git` directory is not decoration: `rg` deliberately ignores a
// `.gitignore` that is not inside a repository, so a bare temp directory would
// measure nothing about how it behaves in a project.
func TestFindProjectSkipsIgnoredBuildOutput(t *testing.T) {
	root := makeProject(t, map[string]string{
		"src/a.txt":               "alvo\n",
		"target/debug/gerado.txt": "alvo\n",
		"node_modules/pkg/i.txt":  "alvo\n",
		".gitignore":              "target/\nnode_modules/\n",
		".git/HEAD":               "ref: refs/heads/main\n",
	})

	hits, _, err := FindProject(root, Query{Text: "alvo"})
	if err != nil {
		t.Fatalf("FindProject: %v", err)
	}
	if len(hits) != 1 || hits[0].Path != "src/a.txt" {
		t.Errorf("hits = %v, esperado só src/a.txt", paths(hits))
	}
}

// TestBackendsSkipTheGitDirectory: `.git` is the one entry both skip
// unconditionally, and it is what makes a search usable in any repository.
func TestBackendsSkipTheGitDirectory(t *testing.T) {
	root := makeProject(t, map[string]string{
		"src/a.txt":        "alvo\n",
		".git/objects/abc": "alvo\n",
	})

	hits, _, err := FindProject(root, Query{Text: "alvo"})
	if err != nil {
		t.Fatalf("FindProject: %v", err)
	}
	for _, hit := range hits {
		if strings.HasPrefix(hit.Path, ".git/") {
			t.Errorf("hit dentro de .git: %q", hit.Path)
		}
	}
}

// TestTheFallbackIgnoresMoreThanRipgrep pins a measured difference between the
// two backends: `rg` skips only what git ignores, so in a directory with no
// `.gitignore` it does search `target/`. The fallback skips build output
// unconditionally. Recorded as B25 in the parity ledger.
func TestTheFallbackIgnoresMoreThanRipgrep(t *testing.T) {
	root := makeProject(t, map[string]string{
		"src/a.txt":               "alvo\n",
		"target/debug/gerado.txt": "alvo\n",
	})

	walked, err := searchWithWalk(root, Query{Text: "alvo", MaxHits: 100})
	if err != nil {
		t.Fatalf("searchWithWalk: %v", err)
	}
	if len(walked) != 1 || walked[0].Path != "src/a.txt" {
		t.Errorf("o fallback = %v, esperado só src/a.txt", paths(walked))
	}
}

func TestFindProjectRespectsGlobs(t *testing.T) {
	root := makeProject(t, map[string]string{
		"a.txt": "alvo\n",
		"b.rs":  "alvo\n",
		"c.md":  "alvo\n",
	})

	hits, _, err := FindProject(root, Query{Text: "alvo", Globs: []string{"*.rs"}})
	if err != nil {
		t.Fatalf("FindProject: %v", err)
	}
	if len(hits) != 1 || hits[0].Path != "b.rs" {
		t.Errorf("com include: %+v", paths(hits))
	}

	// An exclusion alone means "everything except this".
	excluded, _, err := FindProject(root, Query{Text: "alvo", Globs: []string{"!*.md"}})
	if err != nil {
		t.Fatalf("FindProject: %v", err)
	}
	for _, hit := range excluded {
		if hit.Path == "c.md" {
			t.Errorf("um arquivo excluído foi pesquisado: %+v", paths(excluded))
		}
	}
}

// TestFindProjectTreatsABinaryFileAsHavingNoLines: hits inside a binary would be
// control characters, which help nobody.
func TestFindProjectTreatsABinaryFileAsHavingNoLines(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "bin.dat"), []byte("alvo\x00mais alvo"), 0o644); err != nil {
		t.Fatalf("escrevendo: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "texto.txt"), []byte("alvo\n"), 0o644); err != nil {
		t.Fatalf("escrevendo: %v", err)
	}

	hits, _, err := FindProject(root, Query{Text: "alvo"})
	if err != nil {
		t.Fatalf("FindProject: %v", err)
	}
	for _, hit := range hits {
		if hit.Path == "bin.dat" {
			t.Errorf("um arquivo binário foi pesquisado: %+v", hits)
		}
	}
}

func TestFindProjectRejectsAnEmptyQuery(t *testing.T) {
	if _, _, err := FindProject(t.TempDir(), Query{Text: "   "}); err == nil {
		t.Error("consulta vazia foi aceita")
	}
}

func TestFindProjectBoundsTheResults(t *testing.T) {
	files := map[string]string{}
	for i := range 20 {
		files[filepath.Join("d", string(rune('a'+i))+".txt")] = "alvo\nalvo\n"
	}
	root := makeProject(t, files)

	hits, _, err := FindProject(root, Query{Text: "alvo", MaxHits: 5})
	if err != nil {
		t.Fatalf("FindProject: %v", err)
	}
	if len(hits) != 5 {
		t.Errorf("hits = %d, esperado 5", len(hits))
	}
}

func TestFindProjectOnAnEmptyDirectory(t *testing.T) {
	hits, backend, err := FindProject(t.TempDir(), Query{Text: "nada"})
	if err != nil {
		t.Fatalf("FindProject: %v", err)
	}
	if len(hits) != 0 {
		t.Errorf("hits = %+v", hits)
	}
	if backend == "" {
		t.Error("nenhum backend reportado")
	}
}

// TestFindProjectReportsTheBackend: the two do not ignore the same files, so a
// caller that wants to explain a missing result needs to know which answered.
func TestFindProjectReportsTheBackend(t *testing.T) {
	root := makeProject(t, map[string]string{"a.txt": "alvo\n"})
	_, backend, err := FindProject(root, Query{Text: "alvo"})
	if err != nil {
		t.Fatalf("FindProject: %v", err)
	}
	if backend != BackendRipgrep && backend != BackendWalk {
		t.Errorf("backend = %q", backend)
	}
}

func TestGlobAllows(t *testing.T) {
	cases := []struct {
		globs    []string
		relative string
		want     bool
	}{
		{nil, "src/a.rs", true},
		{[]string{"*.rs"}, "src/a.rs", true},
		{[]string{"*.rs"}, "src/a.txt", false},
		{[]string{"src/**"}, "src/deep/a.txt", true},
		{[]string{"**/*.rs"}, "src/deep/a.rs", true},
		{[]string{"!*.md"}, "notas.md", false},
		{[]string{"!*.md"}, "a.rs", true},
		{[]string{"*.rs", "!test_*.rs"}, "test_a.rs", false},
		{[]string{"*.rs", "!test_*.rs"}, "a.rs", true},
	}
	for _, tc := range cases {
		if got := globAllows(tc.globs, tc.relative); got != tc.want {
			t.Errorf("globAllows(%v, %q) = %v, esperado %v", tc.globs, tc.relative, got, tc.want)
		}
	}
}

// TestWalkBackendMatchesTheSameThings exercises the fallback directly, so it is
// covered even on a machine where `rg` answers first.
func TestWalkBackendMatchesTheSameThings(t *testing.T) {
	root := makeProject(t, map[string]string{
		"a.txt":     "alfa\nbeta\n",
		"sub/b.txt": "beta\n",
	})

	hits, err := searchWithWalk(root, Query{Text: "beta", MaxHits: 100})
	if err != nil {
		t.Fatalf("searchWithWalk: %v", err)
	}
	if len(hits) != 2 {
		t.Fatalf("hits = %+v", hits)
	}
	// Sorted by path, then line: the order is what the results panel shows.
	if hits[0].Path != "a.txt" || hits[1].Path != "sub/b.txt" {
		t.Errorf("ordem = %v", paths(hits))
	}
}

func TestWalkBackendReportsColumns(t *testing.T) {
	root := makeProject(t, map[string]string{"a.txt": "xx alvo yy\n"})
	hits, err := searchWithWalk(root, Query{Text: "alvo", MaxHits: 10})
	if err != nil {
		t.Fatalf("searchWithWalk: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("hits = %+v", hits)
	}
	if hits[0].Column != 4 {
		t.Errorf("coluna = %d, esperado 4 (1-based, \"alvo\" começa depois de \"xx \")", hits[0].Column)
	}
	if hits[0].Text != "xx alvo yy" {
		t.Errorf("texto = %q", hits[0].Text)
	}
}

func TestWalkBackendSkipsHiddenEntries(t *testing.T) {
	root := makeProject(t, map[string]string{
		"visivel.txt":   "alvo\n",
		".oculto.txt":   "alvo\n",
		".config/x.txt": "alvo\n",
	})

	hits, err := searchWithWalk(root, Query{Text: "alvo", MaxHits: 100})
	if err != nil {
		t.Fatalf("searchWithWalk: %v", err)
	}
	if len(hits) != 1 || hits[0].Path != "visivel.txt" {
		t.Errorf("hits = %v", paths(hits))
	}
}

// TestWalkBackendSearchesAHiddenRoot: a workspace under a dot directory must
// still be searched, or nothing is.
func TestWalkBackendSearchesAHiddenRoot(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, ".workspace")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("criando: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("alvo\n"), 0o644); err != nil {
		t.Fatalf("escrevendo: %v", err)
	}

	hits, err := searchWithWalk(root, Query{Text: "alvo", MaxHits: 10})
	if err != nil {
		t.Fatalf("searchWithWalk: %v", err)
	}
	if len(hits) != 1 {
		t.Errorf("hits = %+v, esperado um — a raiz oculta não pode ser pulada", hits)
	}
}
