// Package focus holds the focus graph: where input can land, in what order, and
// what outranks what.
//
// The graph is data. The reference implementation resolved focus through the
// order of `if` statements in a two-thousand-line function — the order existed,
// but nothing could assert it, so an `if` inserted at the top could silently
// steal a key from a modal. Here the order is a value, and the rules in
// docs/ui-ux/focus-graph.md are testable directly.
//
// It knows nothing about internal/app. The composition root maps a Surface onto
// the model's own focus vocabulary, which is what keeps this package reusable and
// a surface disposable.
package focus

// Surface names a place input can land.
type Surface string

// The surfaces, in the order Tab walks them.
const (
	Editor    Surface = "editor"
	MenuBar   Surface = "menubar"
	Tabs      Surface = "tabs"
	Tree      Surface = "tree"
	SCM       Surface = "scm"
	Terminal  Surface = "terminal"
	StatusBar Surface = "statusbar"
)

// declaration is the canonical order, and the single source of it.
//
// Editor first: it is the root, and it is where focus goes back to whenever the
// current target stops being visible.
var declaration = []Surface{Editor, MenuBar, Tabs, Tree, SCM, Terminal, StatusBar}

// Graph is the declared focus order.
type Graph struct {
	order []Surface
}

// New returns the graph as documented.
func New() Graph {
	return Graph{order: append([]Surface(nil), declaration...)}
}

// Order lists the surfaces in walk order.
func (g Graph) Order() []Surface {
	return append([]Surface(nil), g.order...)
}

// Holds reports whether a surface belongs to the graph.
func (g Graph) Holds(surface Surface) bool {
	for _, candidate := range g.order {
		if candidate == surface {
			return true
		}
	}
	return false
}

// Visible reports whether a surface may hold focus right now.
type Visible func(Surface) bool

// Next returns the surface focus moves to from `from`.
//
// Hidden surfaces are skipped rather than landed on: focus must never rest where
// nothing is drawn, because the next keystroke would then go somewhere invisible.
// When only the editor is visible it stays put, so Tab is never a way to lose the
// cursor.
func (g Graph) Next(from Surface, visible Visible, backwards bool) Surface {
	start := g.indexOf(from)
	if start < 0 {
		return g.fallback(visible)
	}

	step := 1
	if backwards {
		step = len(g.order) - 1
	}

	// One full lap: if nothing else is visible, the editor is still the answer.
	for offset := 1; offset <= len(g.order); offset++ {
		candidate := g.order[(start+offset*step)%len(g.order)]
		if visible(candidate) {
			return candidate
		}
	}
	return g.fallback(visible)
}

// fallback is where focus goes when the current surface is unknown.
func (g Graph) fallback(visible Visible) Surface {
	if visible(Editor) {
		return Editor
	}
	// A model that has hidden the editor has hidden every surface; returning the
	// editor anyway is better than returning nothing, and the caller is told
	// which surface it asked about.
	return Editor
}

// indexOf finds a surface's position, or -1.
func (g Graph) indexOf(surface Surface) int {
	for index, candidate := range g.order {
		if candidate == surface {
			return index
		}
	}
	return -1
}
