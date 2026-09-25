// Package architecture turns the project's layering contract into executable
// rules.
//
// docs/architecture/clean-code-contract.md §3 states that no package may import
// its consumer, that dependency cycles are strictly forbidden and checked in the
// pipeline, and that every folder carries its own README. A clause that only
// lives in prose is a clause that decays; this package makes each one fail the
// build instead.
//
// The rules are data, not seven bespoke test functions: a Rule knows its own
// name, its rationale and how to find violations, so the test that runs them is
// one loop and adding a rule does not mean writing another test.
package architecture

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Workspace is the repository under inspection.
type Workspace struct {
	// Root is the repository root, absolute.
	Root string
	// Graph is the import graph between first-party packages.
	Graph *Graph
}

// Graph is the import graph, keyed by module-relative package directory.
//
// A key is a directory rather than an import path so the rules can talk about
// "internal/tui/editor" without repeating the module prefix, and so a package
// that is deleted leaves no dangling reference behind.
type Graph struct {
	edges map[string][]string
}

// Load parses every first-party Go package under root and returns their imports.
//
// It parses imports only, without type-checking: the rules are about which
// packages name each other, and asking the compiler to resolve symbols would
// make the check as slow as a build and as fragile as the code it inspects.
func Load(root, module string) (*Graph, error) {
	collected := map[string]map[string]bool{}
	fset := token.NewFileSet()

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// An unreadable directory is skipped rather than fatal: a permission
			// problem in an unrelated corner should not hide the whole graph.
			if entry != nil && entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return skipDirectory(root, path, entry.Name())
		}
		if !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}

		file, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return fmt.Errorf("analisando %s: %w", path, parseErr)
		}

		dir, relErr := filepath.Rel(root, filepath.Dir(path))
		if relErr != nil {
			return nil
		}
		dir = filepath.ToSlash(dir)
		if collected[dir] == nil {
			collected[dir] = map[string]bool{}
		}
		for _, spec := range file.Imports {
			importPath, unquoteErr := unquote(spec.Path.Value)
			if unquoteErr != nil {
				continue
			}
			if firstParty, ok := firstPartyDir(importPath, module); ok {
				collected[dir][firstParty] = true
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	graph := &Graph{edges: map[string][]string{}}
	for dir, imports := range collected {
		list := make([]string, 0, len(imports))
		for imported := range imports {
			list = append(list, imported)
		}
		sort.Strings(list)
		graph.edges[dir] = list
	}
	return graph, nil
}

// skipDirectory decides whether a directory is outside the product's Go graph.
func skipDirectory(root, path, name string) error {
	if path == root {
		return nil
	}
	// Hidden directories hold tooling state (.git, .agents, .ai, .prumo), and
	// the skill example programs inside them are not part of the module.
	if strings.HasPrefix(name, ".") {
		return fs.SkipDir
	}
	switch name {
	case "crates", "target", "node_modules", "testdata", "vendor", "packaging":
		return fs.SkipDir
	}
	return nil
}

// firstPartyDir converts an import path to a module-relative directory.
func firstPartyDir(importPath, module string) (string, bool) {
	if importPath == module {
		return ".", true
	}
	prefix := module + "/"
	if !strings.HasPrefix(importPath, prefix) {
		return "", false
	}
	return strings.TrimPrefix(importPath, prefix), true
}

// unquote reads a Go string literal from the import syntax tree.
func unquote(literal string) (string, error) {
	value, err := strconv.Unquote(literal)
	if err != nil {
		return "", fmt.Errorf("literal de import inválido %s: %w", literal, err)
	}
	return value, nil
}

// Packages lists every package directory, sorted.
func (g *Graph) Packages() []string {
	out := make([]string, 0, len(g.edges))
	for dir := range g.edges {
		out = append(out, dir)
	}
	sort.Strings(out)
	return out
}

// Imports lists what a package directory imports, sorted.
func (g *Graph) Imports(dir string) []string {
	list := append([]string(nil), g.edges[dir]...)
	sort.Strings(list)
	return list
}

// DirectoryExists reports whether the graph has a package at this directory.
func (g *Graph) DirectoryExists(dir string) bool {
	_, ok := g.edges[dir]
	return ok
}

// Rule is one clause of the architecture contract, made checkable.
type Rule struct {
	// Name identifies the rule in a failure message.
	Name string
	// Rationale is the contract clause it enforces, quoted so a reader who
	// breaks the rule learns why it exists without opening another document.
	Rationale string
	// Check returns one message per violation, and nothing when satisfied.
	Check func(Workspace) []string
}

// Rules returns the architecture contract as executable clauses.
func Rules() []Rule {
	return []Rule{
		{
			Name:      "R1-sem-ciclos",
			Rationale: "clean-code-contract §3: ciclos de dependência são estritamente proibidos",
			Check: func(w Workspace) []string {
				return w.Graph.Cycles()
			},
		},
		{
			Name:      "R2-app-nao-importa-consumidor",
			Rationale: "clean-code-contract §3: nenhum pacote pode importar seu consumidor",
			Check: func(w Workspace) []string {
				return w.Graph.Forbidden("internal/app", "internal/tui", "conformance")
			},
		},
		{
			Name:      "R3-superficies-nao-importam-app",
			Rationale: "o composition root de internal/tui monta as views; uma superfície que importa o modelo deixa de ser descartável",
			Check: func(w Workspace) []string {
				return w.Graph.SurfaceImportsApp()
			},
		},
		{
			Name:      "R4-folhas-permanecem-folhas",
			Rationale: "buffer e action são domínio puro; importar outro pacote interno inverteria a direção",
			Check: func(w Workspace) []string {
				var out []string
				out = append(out, w.Graph.ImportsOfLeaf("internal/buffer")...)
				out = append(out, w.Graph.ImportsOfLeaf("internal/action")...)
				return out
			},
		},
		{
			Name:      "R5-nomes-proibidos",
			Rationale: "clean-code-contract §1.3 proíbe utils, common, helpers e data",
			Check: func(w Workspace) []string {
				return ForbiddenDirectoryNames(w.Graph)
			},
		},
		{
			Name:      "R6-readme-por-pasta",
			Rationale: "clean-code-contract §3: toda pasta deve conter seu respectivo README.md",
			Check: func(w Workspace) []string {
				return MissingReadme(w)
			},
		},
		{
			Name:      "R7-harness-nao-e-dependencia",
			Rationale: "o runner de conformidade compara o produto; se o produto o importasse, a comparação deixaria de ser externa",
			Check: func(w Workspace) []string {
				var out []string
				for _, dir := range w.Graph.Packages() {
					if dir == "conformance" || strings.HasPrefix(dir, "conformance/") {
						continue
					}
					if dir == "internal/architecture" {
						continue
					}
					for _, imported := range w.Graph.Imports(dir) {
						if imported == "conformance" {
							out = append(out, fmt.Sprintf("%s importa conformance", dir))
						}
					}
				}
				sort.Strings(out)
				return out
			},
		},
	}
}

// Forbidden reports packages under from that import a forbidden directory or any
// directory below it.
func (g *Graph) Forbidden(from string, forbidden ...string) []string {
	var out []string
	for _, imported := range g.Imports(from) {
		for _, banned := range forbidden {
			if imported == banned || strings.HasPrefix(imported, banned+"/") {
				out = append(out, fmt.Sprintf("%s importa %s", from, imported))
			}
		}
	}
	sort.Strings(out)
	return out
}

// SurfaceImportsApp reports surface packages that reach for the model.
//
// The composition root may import internal/app — that is its whole job. Every
// other package under internal/tui must receive what it needs as data, or it can
// no longer be deleted by removing one line from the root.
func (g *Graph) SurfaceImportsApp() []string {
	const root = "internal/tui"
	const model = "internal/app"

	var out []string
	for _, dir := range g.Packages() {
		if dir == root || !strings.HasPrefix(dir, root+"/") {
			continue
		}
		for _, imported := range g.Imports(dir) {
			if imported == model || strings.HasPrefix(imported, model+"/") {
				out = append(out, fmt.Sprintf("%s importa %s", dir, imported))
			}
		}
	}
	sort.Strings(out)
	return out
}

// ImportsOfLeaf reports internal packages imported by a package that must stay a
// leaf.
func (g *Graph) ImportsOfLeaf(dir string) []string {
	var out []string
	for _, imported := range g.Imports(dir) {
		if strings.HasPrefix(imported, "internal/") {
			out = append(out, fmt.Sprintf("%s importa %s", dir, imported))
		}
	}
	sort.Strings(out)
	return out
}

// Cycles reports dependency cycles as human-readable chains.
func (g *Graph) Cycles() []string {
	const (
		white = 0
		grey  = 1
		black = 2
	)
	colour := map[string]int{}
	var stack []string
	var found []string
	seen := map[string]bool{}

	var visit func(string)
	visit = func(dir string) {
		colour[dir] = grey
		stack = append(stack, dir)

		for _, next := range g.Imports(dir) {
			switch colour[next] {
			case grey:
				chain := cycleFrom(stack, next)
				key := strings.Join(chain, " → ")
				if !seen[key] {
					seen[key] = true
					found = append(found, key)
				}
			case white:
				if g.DirectoryExists(next) {
					visit(next)
				}
			}
		}

		stack = stack[:len(stack)-1]
		colour[dir] = black
	}

	for _, dir := range g.Packages() {
		if colour[dir] == white {
			visit(dir)
		}
	}
	sort.Strings(found)
	return found
}

// cycleFrom extracts the cycle that closes on target, for a readable message.
func cycleFrom(stack []string, target string) []string {
	for index, dir := range stack {
		if dir == target {
			return append(append([]string(nil), stack[index:]...), target)
		}
	}
	return append(append([]string(nil), stack...), target)
}

// ForbiddenDirectoryNames reports package directories whose name is banned.
func ForbiddenDirectoryNames(g *Graph) []string {
	banned := map[string]bool{"utils": true, "common": true, "helpers": true, "data": true}

	var out []string
	for _, dir := range g.Packages() {
		for _, part := range strings.Split(dir, "/") {
			if banned[part] {
				out = append(out, fmt.Sprintf("%s: nome de pasta proibido %q", dir, part))
			}
		}
	}
	sort.Strings(out)
	return out
}

// readmeRequiredRoots are the trees whose every folder must explain itself.
var readmeRequiredRoots = []string{"internal", "cmd", "docs/ui-ux"}

// MissingReadme reports directories that owe a README and do not have one.
//
// It reads the filesystem rather than the import graph, because the contract is
// about the repository's shape, not about Go: a documentation folder owes a
// README just as a package does.
func MissingReadme(w Workspace) []string {
	var out []string

	for _, base := range readmeRequiredRoots {
		root := filepath.Join(w.Root, filepath.FromSlash(base))
		if info, err := os.Stat(root); err != nil || !info.IsDir() {
			// A tree that does not exist yet owes nothing.
			continue
		}

		walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || !entry.IsDir() {
				return nil
			}
			if strings.HasPrefix(entry.Name(), ".") {
				return fs.SkipDir
			}
			// testdata is a compiler convention, ignored by the toolchain and
			// holding fixtures rather than documentation. A README there would
			// be a document nobody reads about a file nobody reads.
			if entry.Name() == "testdata" {
				return fs.SkipDir
			}
			readme := filepath.Join(path, "README.md")
			if info, statErr := os.Stat(readme); statErr != nil || info.IsDir() {
				rel, _ := filepath.Rel(w.Root, path)
				out = append(out, filepath.ToSlash(rel)+" não tem README.md")
			}
			return nil
		})
		if walkErr != nil {
			out = append(out, fmt.Sprintf("não consegui varrer %s: %v", base, walkErr))
		}
	}

	sort.Strings(out)
	return out
}
