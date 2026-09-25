package term

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestSpawnAndEcho drives a real shell on a real PTY.
//
// This is the one test that depends on the environment, so it waits for the
// expected output instead of sleeping for a guessed duration: a fixed sleep is
// what makes a test like this flaky on a loaded machine.
func TestSpawnAndEcho(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("o PTY é Unix; no Windows o suporte exige ConPTY")
	}
	if _, err := os.Stat("/bin/sh"); err != nil {
		t.Skip("/bin/sh não existe")
	}

	terminal, err := Spawn(t.TempDir(), 80, 24, "/bin/sh")
	if err != nil {
		t.Fatalf("Spawn: %v", err)
	}
	defer func() {
		if err := terminal.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	if err := terminal.WriteString("echo HELLO_ORIDE\r"); err != nil {
		t.Fatalf("WriteString: %v", err)
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		terminal.Poll()
		for _, line := range terminal.VisibleLines(40) {
			if strings.Contains(line, "HELLO_ORIDE") {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("HELLO_ORIDE não apareceu; último conteúdo: %q", terminal.VisibleLines(40))
}

// TestSpawnRejectsAnImpossibleShell: a missing shell is a value, not a crash.
func TestSpawnRejectsAnImpossibleShell(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("o PTY é Unix")
	}

	_, err := Spawn(t.TempDir(), 80, 24, "/nao/existe/este/shell")
	if err == nil {
		t.Fatal("um shell inexistente foi aceito")
	}
}

// TestSpawnUsesMinimalDimensions: a shell given a zero-sized terminal cannot
// render a prompt, so the size is clamped rather than passed through.
func TestSpawnUsesMinimalDimensions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("o PTY é Unix")
	}

	terminal, err := Spawn(t.TempDir(), 0, 0, "/bin/sh")
	if err != nil {
		t.Fatalf("Spawn: %v", err)
	}
	defer func() { _ = terminal.Close() }()

	// Nothing to assert about the PTY size through this API, but the call must
	// succeed rather than fail on the invalid request.
	terminal.Resize(0, 0)
}

func TestCloseIsIdempotent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("o PTY é Unix")
	}

	terminal, err := Spawn(t.TempDir(), 80, 24, "/bin/sh")
	if err != nil {
		t.Fatalf("Spawn: %v", err)
	}

	for range 3 {
		if err := terminal.Close(); err != nil {
			t.Fatalf("Close repetido: %v", err)
		}
	}
}

// TestPollReportsAShellThatExited: the panel must say the shell is gone rather
// than sit silently empty.
func TestPollReportsAShellThatExited(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("o PTY é Unix")
	}

	terminal, err := Spawn(t.TempDir(), 80, 24, "/bin/sh")
	if err != nil {
		t.Fatalf("Spawn: %v", err)
	}
	defer func() { _ = terminal.Close() }()

	if err := terminal.WriteString("exit\r"); err != nil {
		t.Fatalf("WriteString: %v", err)
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		terminal.Poll()
		if terminal.LastError() != "" {
			if !errors.Is(errors.New(terminal.LastError()), ErrClosed) && terminal.LastError() != ErrClosed.Error() {
				t.Errorf("erro = %q, esperado %q", terminal.LastError(), ErrClosed.Error())
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("o encerramento do shell não foi reportado")
}

func TestWriteToAClosedTerminalFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("o PTY é Unix")
	}

	terminal, err := Spawn(t.TempDir(), 80, 24, "/bin/sh")
	if err != nil {
		t.Fatalf("Spawn: %v", err)
	}
	if err := terminal.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if err := terminal.WriteString("eco\n"); err == nil {
		t.Error("escrever num terminal fechado não falhou")
	}
}

// TestSpawnRunsInTheGivenDirectory: the shell must start where the editor is,
// or every command runs somewhere else.
func TestSpawnRunsInTheGivenDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("o PTY é Unix")
	}

	dir := t.TempDir()
	canonical, err := filepath.EvalSymlinks(dir)
	if err != nil {
		canonical = dir
	}

	terminal, err := Spawn(dir, 80, 24, "/bin/sh")
	if err != nil {
		t.Fatalf("Spawn: %v", err)
	}
	defer func() { _ = terminal.Close() }()

	if err := terminal.WriteString("pwd\r"); err != nil {
		t.Fatalf("WriteString: %v", err)
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		terminal.Poll()
		for _, line := range terminal.VisibleLines(40) {
			if strings.Contains(line, canonical) || strings.Contains(line, dir) {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("o diretório de trabalho não apareceu: %q", terminal.VisibleLines(40))
}

func TestIsInteractiveShell(t *testing.T) {
	cases := map[string]bool{
		"/bin/bash":     true,
		"/usr/bin/zsh":  true,
		"/usr/bin/fish": true,
		"/bin/sh":       true,
		"/bin/dash":     false,
		"/bin/ksh":      false,
	}
	for shell, want := range cases {
		if got := isInteractiveShell(shell); got != want {
			t.Errorf("isInteractiveShell(%q) = %v, esperado %v", shell, got, want)
		}
	}
}
