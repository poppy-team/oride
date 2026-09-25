// Command oride is the Oride terminal IDE.
//
// This is the Go implementation. It is a composition root and nothing else:
// argument parsing, then delegation. Product decisions belong to the packages
// under internal/, which are usable without a terminal.
//
// The Rust implementation under crates/ is still the behavioural reference — it
// is the differential oracle for `go test ./conformance/...`.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/ori-team/oride/internal/app"
	"github.com/ori-team/oride/internal/buildinfo"
	"github.com/ori-team/oride/internal/config"
	"github.com/ori-team/oride/internal/editor"
	"github.com/ori-team/oride/internal/fs"
	"github.com/ori-team/oride/internal/keymap"
	"github.com/ori-team/oride/internal/tui"
)

const usage = `oride — terminal IDE

USAGE:
  oride [FILE...]      open the editor
  oride --version      print the version
  oride --help         print this message

Configuration is read from ~/.config/oride/config.toml and .oride/config.toml.
Remapping a key is configuration, not code: see docs/ui-ux/keymap.md for the
canonical table. Parity with the Rust implementation is tracked in
docs/migration/parity-ledger.md and enforced by conformance/.
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	for _, arg := range args {
		switch arg {
		case "--version", "-V", "version":
			fmt.Printf("oride %s\n", buildinfo.String())
			return nil
		case "--help", "-h", "help":
			fmt.Print(usage)
			return nil
		}
	}
	return openEditor(args)
}

// openEditor builds the model and hands control to the runtime.
//
// Failures before the terminal is claimed are returned as errors. Once the
// program is running, the runtime owns the screen and reports through the model —
// which is why nothing here writes to stdout after this point.
func openEditor(paths []string) error {
	workspace := workspaceOf(paths)

	configuration, err := config.LoadMerged(workspace)
	if err != nil {
		return fmt.Errorf("configuração: %w", err)
	}
	keys, err := keymap.FromBindings(configuration.Keys)
	if err != nil {
		return fmt.Errorf("keymap: %w", err)
	}

	application := app.New(editor.NewStore(), configuration, keys)
	application.Workspace = workspace
	application.Tree = openTree(workspace, configuration.Tree.ShowHidden)

	if err := openPaths(application, paths); err != nil {
		return err
	}

	program := tea.NewProgram(tui.New(application, keys))
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("TUI: %w", err)
	}
	return nil
}

// openTree opens the project tree, tolerating a directory that cannot be read.
//
// A missing tree is a missing panel, not a reason to refuse to open the editor:
// editing a file from a directory the process cannot list is still useful.
func openTree(workspace string, showHidden bool) *fs.Tree {
	tree, err := fs.Open(workspace, showHidden)
	if err != nil {
		return nil
	}
	return tree
}

// openPaths loads the files named on the command line, or opens an empty buffer.
func openPaths(application *app.App, paths []string) error {
	opened := 0
	for _, path := range paths {
		if _, err := application.Store.OpenPath(path); err != nil {
			return fmt.Errorf("abrindo %s: %w", path, err)
		}
		opened++
	}
	if opened == 0 {
		application.Store.OpenEmpty()
	}
	return nil
}

// workspaceOf is the directory reported paths are relative to.
//
// A file argument makes its directory the workspace, so the tree shows the
// neighbours of what was opened rather than wherever the process happened to
// start.
func workspaceOf(paths []string) string {
	if len(paths) == 0 {
		working, err := os.Getwd()
		if err != nil {
			return "."
		}
		return working
	}

	directory := filepath.Dir(paths[0])
	absolute, err := filepath.Abs(directory)
	if err != nil {
		return directory
	}
	return absolute
}
