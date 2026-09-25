package architecture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot walks up from the test's directory until it finds go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("diretório de trabalho: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("não achei a raiz do repositório (go.mod)")
		}
		dir = parent
	}
}

// modulePath reads the module path from go.mod, so the fixture does not drift
// from the real one.
func modulePath(t *testing.T, root string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("lendo go.mod: %v", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if rest, found := strings.CutPrefix(line, "module "); found {
			return strings.TrimSpace(rest)
		}
	}
	t.Fatal("go.mod sem linha module")
	return ""
}

func load(t *testing.T) Workspace {
	t.Helper()

	root := repoRoot(t)
	graph, err := Load(root, modulePath(t, root))
	if err != nil {
		t.Fatalf("carregando o grafo: %v", err)
	}
	return Workspace{Root: root, Graph: graph}
}

// TestGraphIsNotEmpty guards every other assertion in this file: a graph that
// failed to load would make all seven rules pass vacuously.
func TestGraphIsNotEmpty(t *testing.T) {
	workspace := load(t)

	packages := workspace.Graph.Packages()
	if len(packages) < 10 {
		t.Fatalf("o grafo tem %d pacotes; esperado o módulo inteiro: %v", len(packages), packages)
	}

	// The domain leaves and the composition root must be visible, or the rules
	// that name them are checking nothing.
	for _, expected := range []string{"internal/buffer", "internal/app", "cmd/oride", "conformance"} {
		if !workspace.Graph.DirectoryExists(expected) {
			t.Errorf("%s não está no grafo", expected)
		}
	}

	// Skill example programs must not leak in, or R4 and R5 would police code
	// that is not part of the product.
	for _, dir := range packages {
		if len(dir) > 0 && dir[0] == '.' {
			t.Errorf("diretório oculto no grafo: %s", dir)
		}
		if len(dir) >= 7 && dir[:7] == "crates/" {
			t.Errorf("crates/ no grafo: %s", dir)
		}
	}
}

// TestArchitectureContract runs every clause of the contract against the real
// repository.
func TestArchitectureContract(t *testing.T) {
	workspace := load(t)

	for _, rule := range Rules() {
		t.Run(rule.Name, func(t *testing.T) {
			violations := rule.Check(workspace)
			if len(violations) == 0 {
				return
			}

			t.Errorf("%s\n%s\n", rule.Rationale, rule.Name)
			for _, violation := range violations {
				t.Errorf("  • %s", violation)
			}
		})
	}
}

// TestRulesCanFail is the contraprova.
//
// A rule that can never fail is not measuring anything, so each one is given a
// workspace that breaks it and must report the violation — and a workspace that
// satisfies it, and must stay silent. Both halves matter: the first catches a
// rule that is vacuous, the second a rule that always complains.
func TestRulesCanFail(t *testing.T) {
	// A leaf imports no internal package at all — that is what R4 asserts, and
	// an earlier version of this fixture broke it while claiming to be clean.
	clean := &Graph{edges: map[string][]string{
		"internal/action":     {},
		"internal/buffer":     {},
		"internal/tui":        {"internal/app", "internal/tui/editor"},
		"internal/tui/editor": {"internal/buffer"},
		"cmd/oride":           {"internal/tui"},
		"conformance":         {"internal/app"},
	}}

	violating := map[string]*Graph{
		"R1-sem-ciclos": {edges: map[string][]string{
			"internal/a": {"internal/b"},
			"internal/b": {"internal/a"},
		}},
		"R2-app-nao-importa-consumidor": {edges: map[string][]string{
			"internal/app": {"internal/tui"},
		}},
		"R3-superficies-nao-importam-app": {edges: map[string][]string{
			"internal/tui/editor": {"internal/app"},
		}},
		"R4-folhas-permanecem-folhas": {edges: map[string][]string{
			"internal/buffer": {"internal/editor"},
		}},
		"R5-nomes-proibidos": {edges: map[string][]string{
			"internal/utils": {"internal/buffer"},
		}},
		"R7-harness-nao-e-dependencia": {edges: map[string][]string{
			"internal/app": {"conformance"},
		}},
	}

	root := t.TempDir()
	for _, rule := range Rules() {
		t.Run("falha/"+rule.Name, func(t *testing.T) {
			if rule.Name == "R6-readme-por-pasta" {
				// Needs a real filesystem, not a synthetic graph.
				if err := os.MkdirAll(filepath.Join(root, "internal", "sem-readme"), 0o755); err != nil {
					t.Fatalf("criando fixture: %v", err)
				}
				if got := rule.Check(Workspace{Root: root, Graph: clean}); len(got) == 0 {
					t.Fatal("a regra não acusou uma pasta sem README")
				}
				return
			}

			graph, ok := violating[rule.Name]
			if !ok {
				t.Fatalf("a regra %s não tem caso de violação — não está sendo provada", rule.Name)
			}
			if got := rule.Check(Workspace{Root: root, Graph: graph}); len(got) == 0 {
				t.Errorf("a regra não acusou uma violação deliberada:\n%s", rule.Rationale)
			}
		})

		t.Run("silencia/"+rule.Name, func(t *testing.T) {
			if rule.Name == "R6-readme-por-pasta" {
				// A tree with no README-less folder must pass. The temp root has
				// no internal/ at all, which owes nothing.
				if got := rule.Check(Workspace{Root: t.TempDir(), Graph: clean}); len(got) != 0 {
					t.Errorf("a regra acusou um workspace limpo: %v", got)
				}
				return
			}
			if got := rule.Check(Workspace{Root: root, Graph: clean}); len(got) != 0 {
				t.Errorf("a regra acusou um grafo limpo: %v", got)
			}
		})
	}
}

// TestCycleMessageNamesTheChain checks the message is actionable: a cycle report
// that does not say which packages form it forces the reader to re-derive it.
func TestCycleMessageNamesTheChain(t *testing.T) {
	graph := &Graph{edges: map[string][]string{
		"internal/a": {"internal/b"},
		"internal/b": {"internal/c"},
		"internal/c": {"internal/a"},
	}}

	cycles := graph.Cycles()
	if len(cycles) != 1 {
		t.Fatalf("ciclos = %v, esperado exatamente um", cycles)
	}
	if cycles[0] != "internal/a → internal/b → internal/c → internal/a" {
		t.Errorf("cadeia = %q", cycles[0])
	}
}

// TestForbiddenMatchesWholeSegments guards against a prefix rule that would
// forbid "internal/tui" from matching an unrelated "internal/tui-utils".
func TestForbiddenMatchesWholeSegments(t *testing.T) {
	graph := &Graph{edges: map[string][]string{
		"internal/app": {"internal/tuix", "internal/conformance-extra", "internal/tui"},
	}}

	violations := graph.Forbidden("internal/app", "internal/tui", "conformance")
	if len(violations) != 1 {
		t.Fatalf("violações = %v, esperado só internal/tui", violations)
	}
}
