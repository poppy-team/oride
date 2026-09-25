package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ori-team/oride/internal/app"
	"github.com/ori-team/oride/internal/config"
	"github.com/ori-team/oride/internal/editor"
	"github.com/ori-team/oride/internal/keymap"
)

// workspace builds a temporary workspace with a local session directory.
//
// The local directory is what keeps the test off the user's cache: the session
// look-up prefers <workspace>/.oride/session.toml when that directory exists, so
// nothing outside the temporary tree is written.
func workspace(t *testing.T, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".oride"), 0o755); err != nil {
		t.Fatalf("criando .oride: %v", err)
	}
	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("escrevendo %s: %v", name, err)
		}
	}
	return root
}

func buildApp(t *testing.T) *app.App {
	t.Helper()

	keys, err := keymap.FromBindings(config.Default().Keys)
	if err != nil {
		t.Fatalf("keymap: %v", err)
	}
	return app.New(editor.NewStore(), config.Default(), keys)
}

// TestSessionRoundTrip is the whole feature: what was open comes back.
func TestSessionRoundTrip(t *testing.T) {
	root := workspace(t, map[string]string{"a.txt": "um\n", "b.txt": "dois\n"})

	first := buildApp(t)
	first.Workspace = root
	for _, name := range []string{"a.txt", "b.txt"} {
		if _, err := first.Store.OpenPath(filepath.Join(root, name)); err != nil {
			t.Fatalf("abrindo %s: %v", name, err)
		}
	}
	first.ShowTree = false
	if err := persist(first, root); err != nil {
		t.Fatalf("gravando a sessão: %v", err)
	}

	second := buildApp(t)
	second.Workspace = root
	if err := restore(second, nil, root); err != nil {
		t.Fatalf("restaurando: %v", err)
	}

	if got := second.Store.Len(); got != 2 {
		t.Errorf("documentos restaurados = %d, esperado 2", got)
	}
	if second.ShowTree {
		t.Error("ShowTree não foi restaurado")
	}
}

// TestAFileArgumentWinsOverTheSession: someone who typed a path is asking for that
// file, not for their previous session plus that file.
func TestAFileArgumentWinsOverTheSession(t *testing.T) {
	root := workspace(t, map[string]string{"a.txt": "um\n", "b.txt": "dois\n"})

	first := buildApp(t)
	first.Workspace = root
	if _, err := first.Store.OpenPath(filepath.Join(root, "a.txt")); err != nil {
		t.Fatalf("abrindo: %v", err)
	}
	if _, err := first.Store.OpenPath(filepath.Join(root, "b.txt")); err != nil {
		t.Fatalf("abrindo: %v", err)
	}
	if err := persist(first, root); err != nil {
		t.Fatalf("gravando: %v", err)
	}

	second := buildApp(t)
	second.Workspace = root
	if err := restore(second, []string{filepath.Join(root, "b.txt")}, root); err != nil {
		t.Fatalf("restaurando: %v", err)
	}

	if got := second.Store.Len(); got != 1 {
		t.Errorf("documentos = %d, esperado só o da linha de comando", got)
	}
}

// TestAMissingFileIsSkippedNotFatal: refusing to open the editor because one old
// tab vanished would make the session a liability.
func TestAMissingFileIsSkippedNotFatal(t *testing.T) {
	root := workspace(t, map[string]string{"a.txt": "um\n"})

	first := buildApp(t)
	first.Workspace = root
	if _, err := first.Store.OpenPath(filepath.Join(root, "a.txt")); err != nil {
		t.Fatalf("abrindo: %v", err)
	}
	if err := persist(first, root); err != nil {
		t.Fatalf("gravando: %v", err)
	}

	// O arquivo some entre as duas execuções.
	if err := os.Remove(filepath.Join(root, "a.txt")); err != nil {
		t.Fatalf("removendo: %v", err)
	}

	second := buildApp(t)
	second.Workspace = root
	if err := restore(second, nil, root); err != nil {
		t.Fatalf("restaurar falhou por um arquivo ausente: %v", err)
	}
	if got := second.Store.Len(); got != 1 {
		t.Errorf("documentos = %d, esperado um buffer vazio no lugar", got)
	}
}

// TestNoSessionOpensAnEmptyBuffer: a first run is a normal state, not a failure.
func TestNoSessionOpensAnEmptyBuffer(t *testing.T) {
	root := workspace(t, nil)

	fresh := buildApp(t)
	fresh.Workspace = root
	if err := restore(fresh, nil, root); err != nil {
		t.Fatalf("restaurar sem sessão: %v", err)
	}
	if got := fresh.Store.Len(); got != 1 {
		t.Errorf("documentos = %d, esperado um buffer vazio", got)
	}
}
