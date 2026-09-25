package term

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/creack/pty"
)

// Errors the caller may need to distinguish.
var (
	// ErrClosed means the shell is gone.
	ErrClosed = errors.New("shell encerrado")
)

// scrollbackLimit bounds retained output, in bytes.
//
// A long-running build can emit megabytes; keeping all of it would grow the
// editor's memory without bound, and nobody scrolls back that far.
const scrollbackLimit = 400_000

// DefaultHeightLines is the panel height a fresh terminal opens with.
const DefaultHeightLines = 12

// minRows and minCols keep the PTY large enough for a shell to be usable.
const (
	minRows = 2
	minCols = 20
)

// Terminal is a shell running on a pseudo-terminal.
type Terminal struct {
	file *os.File
	cmd  *exec.Cmd

	chunks chan []byte
	done   chan struct{}

	scrollback string
	cursorCol  int
	lastError  string
	closeOnce  sync.Once

	// Visible reports whether the panel is shown.
	Visible bool
	// HeightLines is the panel height in lines.
	HeightLines int
}

// Spawn starts an interactive shell on a PTY in the given directory.
//
// An empty configuredShell falls back to $SHELL and then to /bin/sh, matching the
// reference. The PTY is a process boundary, so every failure here is a value the
// caller turns into a status line rather than a crash.
func Spawn(cwd string, cols, rows uint16, configuredShell string) (*Terminal, error) {
	shell := strings.TrimSpace(configuredShell)
	if shell == "" {
		shell = os.Getenv("SHELL")
	}
	if shell == "" {
		shell = "/bin/sh"
	}

	command := exec.Command(shell)
	if isInteractiveShell(shell) {
		command.Args = append(command.Args, "-i")
	}
	if cwd != "" {
		command.Dir = cwd
	}
	// The shell must be told it is talking to a colour terminal, or prompts come
	// back plain and the panel looks broken.
	command.Env = append(os.Environ(), "TERM=xterm-256color", "COLORTERM=truecolor")

	file, err := pty.StartWithSize(command, &pty.Winsize{
		Rows: max(minRows, rows),
		Cols: max(minCols, cols),
	})
	if err != nil {
		return nil, fmt.Errorf("abrindo pty: %w", err)
	}

	terminal := &Terminal{
		file:        file,
		cmd:         command,
		chunks:      make(chan []byte, 128),
		done:        make(chan struct{}),
		Visible:     false,
		HeightLines: DefaultHeightLines,
	}
	go terminal.readLoop()
	return terminal, nil
}

// isInteractiveShell reports whether the shell should be started with `-i`.
//
// Only the shells that need it, and only by name: an unrecognised shell is left
// alone rather than handed a flag it may reject.
func isInteractiveShell(shell string) bool {
	base := filepath.Base(shell)
	for _, name := range []string{"bash", "zsh", "fish", "sh"} {
		if base == name {
			return true
		}
	}
	return false
}

// readLoop moves PTY output onto the chunk queue.
func (t *Terminal) readLoop() {
	defer close(t.done)

	buffer := make([]byte, 8192)
	for {
		n, err := t.file.Read(buffer)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buffer[:n])
			select {
			case t.chunks <- chunk:
			case <-t.done:
				return
			}
		}
		if err != nil {
			// Reading a PTY master after the child exits fails with EIO on
			// Linux; that is the end of the stream, not a fault.
			return
		}
	}
}

// Poll applies whatever output has arrived.
//
// Non-blocking on purpose: it is called from the editor's frame loop, which must
// never wait on a shell that has nothing to say.
func (t *Terminal) Poll() {
	for {
		select {
		case chunk := <-t.chunks:
			ApplyChunk(&t.scrollback, &t.cursorCol, string(chunk))
			trimScrollback(&t.scrollback, scrollbackLimit)
		default:
			select {
			case <-t.done:
				t.lastError = ErrClosed.Error()
			default:
			}
			return
		}
	}
}

// CursorCol is the current column, in runes.
func (t *Terminal) CursorCol() int { return t.cursorCol }

// LastError is the most recent failure, empty when there is none.
func (t *Terminal) LastError() string { return t.lastError }

// Write sends bytes to the shell.
func (t *Terminal) Write(data []byte) error {
	if _, err := t.file.Write(data); err != nil {
		t.lastError = err.Error()
		return fmt.Errorf("escrevendo no pty: %w", err)
	}
	t.lastError = ""
	return nil
}

// WriteString sends text to the shell.
func (t *Terminal) WriteString(text string) error { return t.Write([]byte(text)) }

// VisibleLines returns the last maxLines lines of scrollback.
func (t *Terminal) VisibleLines(maxLines int) []string {
	if t.scrollback == "" {
		return nil
	}
	maxLines = max(1, maxLines)

	lines := strings.Split(t.scrollback, "\n")
	// A trailing newline terminates the last line; it does not begin an empty
	// one. Splitting alone would show a blank line the shell never wrote.
	if strings.HasSuffix(t.scrollback, "\n") {
		lines = lines[:len(lines)-1]
	}

	if len(lines) <= maxLines {
		return lines
	}
	return lines[len(lines)-maxLines:]
}

// ToggleVisible shows or hides the panel.
func (t *Terminal) ToggleVisible() { t.Visible = !t.Visible }

// Resize changes the PTY size.
func (t *Terminal) Resize(cols, rows uint16) {
	_ = pty.Setsize(t.file, &pty.Winsize{
		Rows: max(minRows, rows),
		Cols: max(minCols, cols),
	})
}

// Grow makes the panel taller and shows it.
func (t *Terminal) Grow(delta int) {
	t.HeightLines = min(40, max(3, t.HeightLines+delta))
	t.Visible = true
}

// Shrink makes the panel shorter.
func (t *Terminal) Shrink(delta int) {
	t.HeightLines = max(3, t.HeightLines-delta)
}

// Close stops the shell.
func (t *Terminal) Close() error {
	var err error
	t.closeOnce.Do(func() {
		// Closing the master sends SIGHUP to the session, which is what makes an
		// interactive shell exit rather than linger.
		if closeErr := t.file.Close(); closeErr != nil {
			err = closeErr
		}
		if t.cmd.Process != nil {
			_ = t.cmd.Process.Kill()
		}
		_ = t.cmd.Wait()
	})
	return err
}

// trimScrollback keeps the last maxBytes bytes.
//
// The cut is advanced to a rune boundary, or the retained text would begin with a
// replacement character where a split character used to be.
func trimScrollback(scrollback *string, maxBytes int) {
	if len(*scrollback) <= maxBytes {
		return
	}

	cut := len(*scrollback) - maxBytes
	for cut < len(*scrollback) && !utf8.RuneStart((*scrollback)[cut]) {
		cut++
	}
	*scrollback = (*scrollback)[cut:]
}
