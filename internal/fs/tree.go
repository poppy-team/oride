// Package fs exposes the project tree.
//
// The tree is lazy: a directory's children are read when it is expanded, not
// when the workspace opens. Opening a repository with a hundred thousand files
// must not cost a hundred thousand stat calls before the first frame.
package fs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Errors returned by the tree.
var (
	ErrNotDirectory  = errors.New("not a directory")
	ErrInvalidName   = errors.New("invalid name")
	ErrAlreadyExists = errors.New("path already exists")
)

// ignoredDirectories are skipped at every level.
//
// `target` and `node_modules` are build output: they are large, they are not
// edited, and listing them buries the files the user cares about.
var ignoredDirectories = map[string]bool{
	"target":       true,
	"node_modules": true,
}

// Row is one flattened tree row.
type Row struct {
	Depth    int
	Name     string
	Path     string
	IsDir    bool
	Expanded bool
}

type node struct {
	name     string
	path     string
	isDir    bool
	expanded bool
	// children is nil until the directory is expanded.
	children []*node
	loaded   bool
}

// Tree is a project tree with lazy expansion.
type Tree struct {
	root       string
	rootName   string
	showHidden bool
	children   []*node
	selected   int
}

// Open reads a directory's first level.
func Open(root string, showHidden bool) (*Tree, error) {
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("resolvendo %s: %w", root, err)
	}
	info, err := os.Stat(canonical)
	if err != nil {
		return nil, fmt.Errorf("abrindo %s: %w", canonical, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrNotDirectory, canonical)
	}

	children, err := readDirNodes(canonical, showHidden)
	if err != nil {
		return nil, err
	}
	return &Tree{
		root:       canonical,
		rootName:   filepath.Base(canonical),
		showHidden: showHidden,
		children:   children,
	}, nil
}

// Root returns the workspace directory.
func (t *Tree) Root() string { return t.root }

// RootName returns the workspace's directory name.
func (t *Tree) RootName() string { return t.rootName }

// SelectedIndex returns the index of the selected row.
func (t *Tree) SelectedIndex() int { return t.selected }

// CountVisibleRows returns how many rows are visible, including the root.
func (t *Tree) CountVisibleRows() int {
	return 1 + countVisible(t.children)
}

// SetSelected moves the selection, clamped to the visible rows.
func (t *Tree) SetSelected(index int) {
	total := t.CountVisibleRows()
	if total == 0 {
		t.selected = 0
		return
	}
	t.selected = min(index, total-1)
}

// MoveSelection moves the selection by a delta, wrapping around.
//
// Wrapping rather than clamping: at the last row, one more step goes to the
// first, which is what a list keyed by a single direction does everywhere else.
func (t *Tree) MoveSelection(delta int) {
	total := t.CountVisibleRows()
	if total == 0 {
		return
	}
	t.selected = ((t.selected+delta)%total + total) % total
}

// SelectedRow returns the row under the selection.
func (t *Tree) SelectedRow() (Row, bool) {
	if t.selected == 0 {
		return Row{
			Depth:    0,
			Name:     t.rootName,
			Path:     t.root,
			IsDir:    true,
			Expanded: true,
		}, true
	}
	remaining := t.selected - 1
	return findNodeAt(t.children, 1, &remaining)
}

// FlatRows returns every visible row, root first.
func (t *Tree) FlatRows() []Row {
	rows := []Row{{
		Depth:    0,
		Name:     t.rootName,
		Path:     t.root,
		IsDir:    true,
		Expanded: true,
	}}
	flatten(t.children, 1, &rows)
	return rows
}

// ToggleSelected expands or collapses the selected directory.
func (t *Tree) ToggleSelected() error {
	row, ok := t.SelectedRow()
	if !ok {
		return nil
	}
	if !row.IsDir {
		return nil
	}
	if row.Path == t.root {
		return nil
	}
	return t.togglePath(row.Path)
}

// ExpandSelected expands the selected directory without collapsing it.
func (t *Tree) ExpandSelected() error {
	row, ok := t.SelectedRow()
	if !ok || !row.IsDir || row.Path == t.root {
		return nil
	}
	return t.setExpanded(row.Path, true)
}

// CollapseOrParent collapses the selection, or moves to the parent when it is
// already collapsed.
func (t *Tree) CollapseOrParent() error {
	row, ok := t.SelectedRow()
	if !ok {
		return nil
	}
	if row.IsDir && row.Expanded && row.Path != t.root {
		return t.setExpanded(row.Path, false)
	}
	if row.Path == t.root {
		return nil
	}
	parent := filepath.Dir(row.Path)
	for index, candidate := range t.FlatRows() {
		if candidate.Path == parent {
			t.SetSelected(index)
			return nil
		}
	}
	return nil
}

// Activate opens the selected row: a file returns its path, a directory expands
// or collapses.
func (t *Tree) Activate() (string, bool, error) {
	row, ok := t.SelectedRow()
	if !ok {
		return "", false, nil
	}
	if !row.IsDir {
		return row.Path, true, nil
	}
	if row.Path == t.root {
		return "", false, nil
	}
	// A collapsed directory expands; an expanded one stays open, so Enter on a
	// directory never hides what the user just revealed.
	if !row.Expanded {
		return "", false, t.setExpanded(row.Path, true)
	}
	return "", false, nil
}

// Refresh re-reads the directories that are currently expanded.
func (t *Tree) Refresh() error {
	children, err := readDirNodes(t.root, t.showHidden)
	if err != nil {
		return err
	}
	t.children = children
	return nil
}

func (t *Tree) togglePath(path string) error {
	return walkNodes(t.children, func(n *node) (bool, error) {
		if n.path != path {
			return false, nil
		}
		if !n.isDir {
			return true, nil
		}
		return true, t.setNodeExpanded(n, !n.expanded)
	})
}

func (t *Tree) setExpanded(path string, expanded bool) error {
	return walkNodes(t.children, func(n *node) (bool, error) {
		if n.path != path {
			return false, nil
		}
		return true, t.setNodeExpanded(n, expanded)
	})
}

func (t *Tree) setNodeExpanded(n *node, expanded bool) error {
	n.expanded = expanded
	if !expanded || n.loaded {
		return nil
	}
	children, err := readDirNodes(n.path, t.showHidden)
	if err != nil {
		return err
	}
	n.children = children
	n.loaded = true
	return nil
}

// walkNodes calls visit on each node until it reports done.
func walkNodes(nodes []*node, visit func(*node) (bool, error)) error {
	for _, n := range nodes {
		done, err := visit(n)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		if len(n.children) > 0 {
			if err := walkNodes(n.children, visit); err != nil {
				return err
			}
		}
	}
	return nil
}

// readDirNodes lists a directory, directories first and then by name.
//
// The order is observable: it is what the tree shows and what the conformance
// dump records, so it is sorted here rather than left to the filesystem.
func readDirNodes(dir string, showHidden bool) ([]*node, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("lendo %s: %w", dir, err)
	}

	names := make([]string, 0, len(entries))
	isDir := make(map[string]bool, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}
		if ignoredDirectories[name] {
			continue
		}
		names = append(names, name)
		isDir[name] = entry.IsDir()
	}

	sort.Slice(names, func(i, j int) bool {
		if isDir[names[i]] != isDir[names[j]] {
			return isDir[names[i]]
		}
		return names[i] < names[j]
	})

	nodes := make([]*node, 0, len(names))
	for _, name := range names {
		nodes = append(nodes, &node{
			name:  name,
			path:  filepath.Join(dir, name),
			isDir: isDir[name],
		})
	}
	return nodes, nil
}

func countVisible(nodes []*node) int {
	count := 0
	for _, n := range nodes {
		count++
		if n.isDir && n.expanded {
			count += countVisible(n.children)
		}
	}
	return count
}

func flatten(nodes []*node, depth int, rows *[]Row) {
	for _, n := range nodes {
		*rows = append(*rows, Row{
			Depth:    depth,
			Name:     n.name,
			Path:     n.path,
			IsDir:    n.isDir,
			Expanded: n.expanded,
		})
		if n.isDir && n.expanded {
			flatten(n.children, depth+1, rows)
		}
	}
}

func findNodeAt(nodes []*node, depth int, remaining *int) (Row, bool) {
	for _, n := range nodes {
		if *remaining == 0 {
			return Row{
				Depth:    depth,
				Name:     n.name,
				Path:     n.path,
				IsDir:    n.isDir,
				Expanded: n.expanded,
			}, true
		}
		*remaining--
		if n.isDir && n.expanded {
			if row, ok := findNodeAt(n.children, depth+1, remaining); ok {
				return row, true
			}
		}
	}
	return Row{}, false
}
