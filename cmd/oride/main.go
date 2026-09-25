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
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/ori-team/oride/internal/app"
	"github.com/ori-team/oride/internal/buildinfo"
	"github.com/ori-team/oride/internal/config"
	"github.com/ori-team/oride/internal/editor"
	"github.com/ori-team/oride/internal/fs"
	"github.com/ori-team/oride/internal/keymap"
	"github.com/ori-team/oride/internal/session"
	"github.com/ori-team/oride/internal/tui"
	"github.com/ori-team/oride/internal/tui/theme"
)

const usage = `oride — terminal IDE

USAGE:
  oride [FILE...]            open the editor
  oride --print-frame [WxH]  render one frame to stdout and exit
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
	if len(args) > 0 && args[0] == "--print-frame" {
		return printFrame(os.Stdout, args[1:])
	}
	return openEditor(args)
}

// printFrame renders one frame to a writer and exits.
//
// The frame is a pure function of the model, so it can be inspected without a
// terminal — which is what makes it reviewable in a diff, in a bug report, or in
// a pipeline that has no TTY.
func printFrame(out io.Writer, args []string) error {
	width, height := 100, 30
	if len(args) > 0 {
		if w, h, ok := parseSize(args[0]); ok {
			width, height = w, h
		}
	}

	application := app.New(editor.NewStore(), config.Default(), keymap.New())
	application.Store.OpenEmpty()
	if working, err := os.Getwd(); err == nil {
		application.Workspace = working
		if tree, treeErr := fs.Open(working, false); treeErr == nil {
			application.Tree = tree
		}
	}

	model := tui.New(application, keymap.New()).
		WithProfile(theme.TrueColor).
		Resize(width, height)

	fmt.Fprintln(out, model.Frame())
	return nil
}

// parseSize reads a WxH argument.
func parseSize(arg string) (int, int, bool) {
	widthText, heightText, found := strings.Cut(arg, "x")
	if !found {
		return 0, 0, false
	}
	width, widthErr := strconv.Atoi(widthText)
	height, heightErr := strconv.Atoi(heightText)
	if widthErr != nil || heightErr != nil || width <= 0 || height <= 0 {
		return 0, 0, false
	}
	return width, height, true
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

	if err := restore(application, paths, workspace); err != nil {
		return err
	}

	program := tea.NewProgram(tui.New(application, keys))
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("TUI: %w", err)
	}

	// Written after the runtime has released the terminal, so a failure here
	// reports to a working screen instead of over the alt screen.
	return persist(application, workspace)
}

// restore opens what the command line named, or what was open last time.
//
// A file named on the command line wins outright: someone who typed a path is
// asking for that file, not for their previous session plus that file.
func restore(application *app.App, paths []string, workspace string) error {
	if len(paths) > 0 {
		return openPaths(application, paths)
	}

	previous, found := session.LoadForWorkspace(workspace)
	if !found {
		application.Store.OpenEmpty()
		return nil
	}

	for _, file := range previous.Files {
		if _, err := application.Store.OpenPath(file); err != nil {
			// A file that has since been deleted is skipped rather than fatal:
			// refusing to open the editor because one old tab vanished would make
			// the session a liability.
			continue
		}
	}
	if application.Store.Len() == 0 {
		application.Store.OpenEmpty()
	}

	application.ShowTree = session.BoolValue(previous.ShowTree, true)
	return nil
}

// persist writes the session for the next run.
//
// The active document is stored by index into the file list, not by document id:
// an id is a runtime counter and would point at nothing after a restart.
func persist(application *app.App, workspace string) error {
	files := application.Store.OpenPaths()

	active := 0
	if document, err := application.Store.Active(); err == nil {
		if path, hasPath := document.Path(); hasPath {
			active = indexOfPath(files, path)
		}
	}

	showTree := application.ShowTree

	restored := session.FromWorkspace(workspace, files, active)
	restored.ShowTree = &showTree
	return restored.Save()
}

// indexOfPath finds a path in the list, or nothing.
func indexOfPath(files []string, path string) int {
	for index, candidate := range files {
		if candidate == path {
			return index
		}
	}
	return 0
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
