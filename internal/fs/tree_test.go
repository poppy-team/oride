package fs

import (
	"os"
	"path/filepath"
	"testing"
)

// makeTree writes a small directory layout for a test.
func makeTree(t *testing.T, files []string) string {
	t.Helper()

	root := t.TempDir()
	for _, file := range files {
		target := filepath.Join(root, file)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatalf("criando %s: %v", filepath.Dir(target), err)
		}
		if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
			t.Fatalf("escrevendo %s: %v", target, err)
		}
	}
	return root
}

func names(rows []Row) []string {
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Name)
	}
	return out
}

// TestOrderIsDirectoriesThenNames pins the order, which is what the tree shows
// and what the state dump records.
func TestOrderIsDirectoriesThenNames(t *testing.T) {
	root := makeTree(t, []string{
		"zeta.txt", "alfa.txt", "pasta/interno.txt", "outra/mais.txt",
	})

	tree, err := Open(root, false)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	rows := tree.FlatRows()
	if rows[0].Depth != 0 || rows[0].Name != filepath.Base(root) {
		t.Fatalf("primeira linha = %+v, esperado a raiz na profundidade 0", rows[0])
	}

	// Directories first, then files, each group by name.
	want := []string{filepath.Base(root), "outra", "pasta", "alfa.txt", "zeta.txt"}
	got := names(rows)
	if len(got) != len(want) {
		t.Fatalf("linhas = %v, esperado %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("linha %d = %q, esperado %q (ordem: %v)", i, got[i], want[i], got)
		}
	}
}

func TestHiddenFilesAreFilteredUnlessAsked(t *testing.T) {
	root := makeTree(t, []string{"visivel.txt", ".oculto.txt", ".config/x.txt"})

	hidden, err := Open(root, false)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	for _, name := range names(hidden.FlatRows()) {
		if len(name) > 0 && name[0] == '.' {
			t.Errorf("%q apareceu com show_hidden desligado", name)
		}
	}

	shown, err := Open(root, true)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	found := false
	for _, name := range names(shown.FlatRows()) {
		if name == ".oculto.txt" {
			found = true
		}
	}
	if !found {
		t.Errorf(".oculto.txt não apareceu com show_hidden ligado: %v", names(shown.FlatRows()))
	}
}

// TestBuildOutputIsAlwaysSkipped: `target` and `node_modules` are large, are not
// edited, and bury the files the user cares about.
func TestBuildOutputIsAlwaysSkipped(t *testing.T) {
	root := makeTree(t, []string{
		"src/main.go", "target/debug/bin", "node_modules/pkg/index.js",
	})

	tree, err := Open(root, true)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	for _, name := range names(tree.FlatRows()) {
		if name == "target" || name == "node_modules" {
			t.Errorf("%q não deveria aparecer: %v", name, names(tree.FlatRows()))
		}
	}
}

// TestExpansionIsLazyAndVisible: a directory's children appear only when it is
// expanded, and the count follows.
func TestExpansionIsLazyAndVisible(t *testing.T) {
	root := makeTree(t, []string{"pasta/a.txt", "pasta/b.txt", "solto.txt"})

	tree, err := Open(root, false)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if got, want := tree.CountVisibleRows(), 3; got != want {
		t.Fatalf("linhas visíveis = %d, esperado %d (raiz, pasta, solto)", got, want)
	}

	// Select "pasta" (index 1) and expand it.
	tree.SetSelected(1)
	row, ok := tree.SelectedRow()
	if !ok || row.Name != "pasta" {
		t.Fatalf("selectedRow = %+v, esperado \"pasta\"", row)
	}
	if err := tree.ToggleSelected(); err != nil {
		t.Fatalf("ToggleSelected: %v", err)
	}
	if got, want := tree.CountVisibleRows(), 5; got != want {
		t.Fatalf("após expandir, linhas = %d, esperado %d", got, want)
	}

	children := names(tree.FlatRows())
	want := []string{filepath.Base(root), "pasta", "a.txt", "b.txt", "solto.txt"}
	for i := range want {
		if children[i] != want[i] {
			t.Fatalf("linhas = %v, esperado %v", children, want)
		}
	}

	// The expansion flag is visible in the row, which is what the dump records.
	expandedRow := tree.FlatRows()[1]
	if !expandedRow.Expanded {
		t.Error("a linha da pasta não está marcada como expandida")
	}
}

func TestCollapseHidesChildrenAgain(t *testing.T) {
	root := makeTree(t, []string{"pasta/a.txt"})

	tree, err := Open(root, false)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	tree.SetSelected(1)
	if err := tree.ToggleSelected(); err != nil {
		t.Fatalf("expandir: %v", err)
	}
	if got := tree.CountVisibleRows(); got != 3 {
		t.Fatalf("após expandir, linhas = %d, esperado 3", got)
	}
	if err := tree.ToggleSelected(); err != nil {
		t.Fatalf("colapsar: %v", err)
	}
	if got := tree.CountVisibleRows(); got != 2 {
		t.Fatalf("após colapsar, linhas = %d, esperado 2", got)
	}
}

// TestMoveSelectionWraps: at the last row one more step goes to the first.
func TestMoveSelectionWraps(t *testing.T) {
	root := makeTree(t, []string{"a.txt", "b.txt"})

	tree, err := Open(root, false)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	total := tree.CountVisibleRows() // root + 2 files

	tree.SetSelected(total - 1)
	tree.MoveSelection(1)
	if got := tree.SelectedIndex(); got != 0 {
		t.Errorf("após a última linha, seleção = %d, esperado 0", got)
	}
	tree.MoveSelection(-1)
	if got := tree.SelectedIndex(); got != total-1 {
		t.Errorf("um passo para trás desde a primeira, seleção = %d, esperado %d", got, total-1)
	}
}

func TestSetSelectedClamps(t *testing.T) {
	root := makeTree(t, []string{"a.txt"})

	tree, err := Open(root, false)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	tree.SetSelected(999)
	if got := tree.SelectedIndex(); got != tree.CountVisibleRows()-1 {
		t.Errorf("seleção = %d, esperado o último índice", got)
	}
}

func TestActivateOnAFileReturnsItsPath(t *testing.T) {
	root := makeTree(t, []string{"notas.txt"})

	tree, err := Open(root, false)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	tree.SetSelected(1)
	path, isFile, err := tree.Activate()
	if err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if !isFile {
		t.Fatal("um arquivo deveria devolver um caminho para abrir")
	}
	if path != filepath.Join(root, "notas.txt") {
		t.Errorf("path = %q", path)
	}
}

// TestActivateOnADirectoryDoesNotHideIt: Enter reveals, it does not toggle, so a
// collapsed directory expands and an expanded one stays open.
func TestActivateOnADirectoryDoesNotHideIt(t *testing.T) {
	root := makeTree(t, []string{"pasta/a.txt"})

	tree, err := Open(root, false)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	tree.SetSelected(1)

	if _, _, err := tree.Activate(); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if got := tree.CountVisibleRows(); got != 3 {
		t.Fatalf("após Enter numa pasta colapsada, linhas = %d, esperado 3", got)
	}
	if _, _, err := tree.Activate(); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if got := tree.CountVisibleRows(); got != 3 {
		t.Errorf("Enter numa pasta já aberta não deveria esconder o conteúdo: linhas = %d", got)
	}
}

// TestCollapseOrParentWalksUp closes the current directory, and when it is
// already closed moves the selection to its parent rather than doing nothing.
func TestCollapseOrParentWalksUp(t *testing.T) {
	root := makeTree(t, []string{"pasta/sub/a.txt"})

	tree, err := Open(root, false)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	tree.SetSelected(1)
	if err := tree.ToggleSelected(); err != nil { // expand "pasta"
		t.Fatalf("expandir pasta: %v", err)
	}
	tree.SetSelected(2)
	if err := tree.ToggleSelected(); err != nil { // expand "sub"
		t.Fatalf("expandir sub: %v", err)
	}
	if got, want := tree.CountVisibleRows(), 4; got != want {
		t.Fatalf("linhas = %d, esperado %d", got, want)
	}

	// On "sub", expanded: collapse it.
	tree.SetSelected(2)
	if err := tree.CollapseOrParent(); err != nil {
		t.Fatalf("CollapseOrParent: %v", err)
	}
	if got, want := tree.CountVisibleRows(), 3; got != want {
		t.Fatalf("após colapsar sub, linhas = %d, esperado %d", got, want)
	}

	// On "sub", now collapsed: the selection goes to its parent.
	if err := tree.CollapseOrParent(); err != nil {
		t.Fatalf("CollapseOrParent: %v", err)
	}
	row, _ := tree.SelectedRow()
	if row.Name != "pasta" {
		t.Errorf("seleção = %q, esperado \"pasta\"", row.Name)
	}
}

func TestOpenRejectsAFile(t *testing.T) {
	root := makeTree(t, []string{"arquivo.txt"})
	if _, err := Open(filepath.Join(root, "arquivo.txt"), false); err == nil {
		t.Error("abrir um arquivo como árvore foi aceito")
	}
}

func TestRootIsCanonical(t *testing.T) {
	root := makeTree(t, []string{"a.txt"})
	tree, err := Open(root, false)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if !filepath.IsAbs(tree.Root()) {
		t.Errorf("raiz não é absoluta: %q", tree.Root())
	}
	if got := tree.RootName(); got != filepath.Base(root) {
		t.Errorf("RootName = %q, esperado %q", got, filepath.Base(root))
	}
}
