package focus

import "testing"

// allVisible is the graph with every surface shown.
func allVisible(Surface) bool { return true }

// onlyEditor matches a model with no side panel open.
func onlyEditor(s Surface) bool { return s == Editor }

// TestOrderMatchesTheDocumentedGraph pins the walk order against
// docs/ui-ux/focus-graph.md. A reordering here changes where Tab goes for every
// user, so the order is asserted rather than assumed.
func TestOrderMatchesTheDocumentedGraph(t *testing.T) {
	want := []Surface{Editor, MenuBar, Tabs, Tree, SCM, Terminal, StatusBar}

	got := New().Order()
	if len(got) != len(want) {
		t.Fatalf("ordem = %v, esperado %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("posição %d = %q, esperado %q", index, got[index], want[index])
		}
	}
}

// TestTabCyclesForwardAndWraps follows the first rule of the graph: Tab walks the
// visible surfaces in order and comes back around.
func TestTabCyclesForwardAndWraps(t *testing.T) {
	graph := New()

	walk := []Surface{MenuBar, Tabs, Tree, SCM, Terminal, StatusBar, Editor}
	from := Editor
	for _, want := range walk {
		from = graph.Next(from, allVisible, false)
		if from != want {
			t.Fatalf("de %v Tab foi para %v, esperado %v", want, from, want)
		}
	}
}

func TestShiftTabCyclesBackwards(t *testing.T) {
	graph := New()

	walk := []Surface{StatusBar, Terminal, SCM, Tree, Tabs, MenuBar, Editor}
	from := Editor
	for _, want := range walk {
		from = graph.Next(from, allVisible, true)
		if from != want {
			t.Fatalf("Shift+Tab foi para %v, esperado %v", from, want)
		}
	}
}

// TestHiddenSurfacesAreSkipped is the rule that keeps the focus off something the
// user cannot see: the next keystroke would otherwise go somewhere invisible.
func TestHiddenSurfacesAreSkipped(t *testing.T) {
	graph := New()
	visible := func(s Surface) bool { return s == Editor || s == Tree }

	if got := graph.Next(Editor, visible, false); got != Tree {
		t.Errorf("Tab foi para %v, esperado Tree", got)
	}
	if got := graph.Next(Tree, visible, false); got != Editor {
		t.Errorf("Tab de Tree foi para %v, esperado Editor", got)
	}
	if got := graph.Next(Editor, visible, true); got != Tree {
		t.Errorf("Shift+Tab foi para %v, esperado Tree", got)
	}
}

// TestFocusStaysPutWhenNothingElseIsVisible: with only the editor shown, Tab must
// not be a way to lose the cursor.
func TestFocusStaysPutWhenNothingElseIsVisible(t *testing.T) {
	graph := New()

	for _, backwards := range []bool{false, true} {
		if got := graph.Next(Editor, onlyEditor, backwards); got != Editor {
			t.Errorf("Tab/backwards=%v foi para %v, esperado Editor", backwards, got)
		}
	}
}

// TestAnUnknownSurfaceFallsBackToTheEditor: a surface that left the graph — a
// panel removed in a later release — must not leave focus dangling.
func TestAnUnknownSurfaceFallsBackToTheEditor(t *testing.T) {
	graph := New()

	if got := graph.Next(Surface("inexistente"), allVisible, false); got != Editor {
		t.Errorf("de uma superfície desconhecida foi para %v, esperado Editor", got)
	}
}

func TestHolds(t *testing.T) {
	graph := New()

	if !graph.Holds(Tree) {
		t.Error("Tree deveria pertencer ao grafo")
	}
	if graph.Holds(Surface("inexistente")) {
		t.Error("uma superfície desconhecida não deveria pertencer ao grafo")
	}
}

// TestOrderIsACopy: handing out the internal slice would let a caller reorder the
// graph from the outside, which is exactly the hidden coupling this package was
// extracted to remove.
func TestOrderIsACopy(t *testing.T) {
	graph := New()

	order := graph.Order()
	order[0] = Surface("adulterado")

	if graph.Order()[0] != Editor {
		t.Error("a ordem devolvida altera o grafo")
	}
}

// TestEverySurfaceIsReachable proves the graph has no unreachable node: a surface
// declared but never reached by Tab is a surface nobody can focus.
func TestEverySurfaceIsReachable(t *testing.T) {
	graph := New()

	seen := map[Surface]bool{}
	from := Editor
	for range graph.Order() {
		from = graph.Next(from, allVisible, false)
		seen[from] = true
	}

	for _, surface := range graph.Order() {
		if !seen[surface] {
			t.Errorf("%v nunca é alcançável por Tab", surface)
		}
	}
}
