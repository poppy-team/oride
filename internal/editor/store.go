package editor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Store holds the open documents and which one is active.
type Store struct {
	nextID uint64
	docs   []*Document
	active *ID
}

// NewStore returns an empty store.
func NewStore() *Store { return &Store{} }

func (s *Store) allocID() ID {
	id := ID(s.nextID)
	s.nextID++
	return id
}

// OpenEmpty adds an empty document with no path and activates it.
func (s *Store) OpenEmpty() ID {
	id := s.allocID()
	s.docs = append(s.docs, NewDocument(id))
	s.active = &id
	return id
}

// OpenPath reads a file and activates it.
//
// A file already open is activated rather than opened twice: two views of one
// file that can diverge is a data-loss bug wearing a feature's clothes.
func (s *Store) OpenPath(path string) (ID, error) {
	canonical := path
	if resolved, err := filepath.Abs(path); err == nil {
		canonical = resolved
	}
	if existing, ok := s.DocumentIDForPath(canonical); ok {
		if err := s.SetActive(existing); err != nil {
			return 0, err
		}
		return existing, nil
	}

	content, err := os.ReadFile(canonical)
	if err != nil {
		return 0, fmt.Errorf("abrindo %s: %w", path, err)
	}

	id := s.allocID()
	document := NewDocumentFromText(id, canonical, string(content))
	s.docs = append(s.docs, document)
	s.active = &id
	return id, nil
}

// Get returns a document by id.
func (s *Store) Get(id ID) (*Document, bool) {
	for _, document := range s.docs {
		if document.ID() == id {
			return document, true
		}
	}
	return nil, false
}

// Active returns the active document.
func (s *Store) Active() (*Document, error) {
	if s.active == nil {
		return nil, ErrDocumentNotFound
	}
	document, ok := s.Get(*s.active)
	if !ok {
		return nil, ErrDocumentNotFound
	}
	return document, nil
}

// ActiveID returns the active document id, and whether there is one.
func (s *Store) ActiveID() (ID, bool) {
	if s.active == nil {
		return 0, false
	}
	return *s.active, true
}

// SetActive makes a document active.
func (s *Store) SetActive(id ID) error {
	if _, ok := s.Get(id); !ok {
		return fmt.Errorf("%w: %d", ErrDocumentNotFound, id)
	}
	s.active = &id
	return nil
}

// TabIDs returns the open document ids in tab order.
func (s *Store) TabIDs() []ID {
	out := make([]ID, 0, len(s.docs))
	for _, document := range s.docs {
		out = append(out, document.ID())
	}
	return out
}

// Len returns the number of open documents.
func (s *Store) Len() int { return len(s.docs) }

// DirtyCount returns how many documents have unsaved changes.
func (s *Store) DirtyCount() int {
	count := 0
	for _, document := range s.docs {
		if document.IsDirty() {
			count++
		}
	}
	return count
}

// DocumentIDForPath finds an open document by path.
func (s *Store) DocumentIDForPath(path string) (ID, bool) {
	canonical := path
	if resolved, err := filepath.Abs(path); err == nil {
		canonical = resolved
	}
	for _, document := range s.docs {
		if stored, ok := document.Path(); ok && stored == canonical {
			return document.ID(), true
		}
	}
	return 0, false
}

// OpenPaths returns the paths of documents that have one, in tab order.
func (s *Store) OpenPaths() []string {
	out := make([]string, 0, len(s.docs))
	for _, document := range s.docs {
		if path, ok := document.Path(); ok {
			out = append(out, path)
		}
	}
	return out
}

// ActivateNextTab cycles forward and returns the new active id.
func (s *Store) ActivateNextTab() (ID, bool) { return s.cycle(1) }

// ActivatePrevTab cycles backward and returns the new active id.
func (s *Store) ActivatePrevTab() (ID, bool) { return s.cycle(-1) }

func (s *Store) cycle(delta int) (ID, bool) {
	if len(s.docs) == 0 {
		return 0, false
	}
	index := 0
	if s.active != nil {
		for i, document := range s.docs {
			if document.ID() == *s.active {
				index = i
				break
			}
		}
	}
	index = ((index+delta)%len(s.docs) + len(s.docs)) % len(s.docs)
	id := s.docs[index].ID()
	s.active = &id
	return id, true
}

// Close removes a document and reports whether any documents remain.
func (s *Store) Close(id ID) (bool, error) {
	index := -1
	for i, document := range s.docs {
		if document.ID() == id {
			index = i
			break
		}
	}
	if index < 0 {
		return false, fmt.Errorf("%w: %d", ErrDocumentNotFound, id)
	}

	s.docs = append(s.docs[:index], s.docs[index+1:]...)

	if s.active != nil && *s.active == id {
		s.active = nil
		if len(s.docs) > 0 {
			next := s.docs[min(index, len(s.docs)-1)].ID()
			s.active = &next
		}
	}
	return len(s.docs) > 0, nil
}

// TabSummary is the label a tab bar needs.
type TabSummary struct {
	ID    ID
	Title string
	Dirty bool
	Path  string
}

// TabSummaries returns a summary per open document, in tab order.
func (s *Store) TabSummaries() []TabSummary {
	out := make([]TabSummary, 0, len(s.docs))
	for _, document := range s.docs {
		path, _ := document.Path()
		out = append(out, TabSummary{
			ID:    document.ID(),
			Title: document.TabTitle(),
			Dirty: document.IsDirty(),
			Path:  path,
		})
	}
	return out
}

// SortedPaths returns the open paths in a stable order, for state dumps.
func (s *Store) SortedPaths() []string {
	paths := s.OpenPaths()
	sort.Strings(paths)
	return paths
}

// ErrNoDocuments is returned when an operation needs an open document.
var ErrNoDocuments = errors.New("no documents open")
