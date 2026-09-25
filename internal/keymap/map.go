package keymap

import (
	"fmt"
	"sort"

	"github.com/ori-team/oride/internal/action"
)

// Binding is one chord resolved to one action.
type Binding struct {
	Chord  Chord
	Action action.Action
}

// Map resolves chords to actions.
//
// The zero value is not usable; use New or FromBindings. A nil *Map is not a
// valid empty map — an unbound chord and a missing keymap must not look the
// same, or a misconfiguration silently disables every shortcut.
type Map struct {
	bindings map[Chord]action.Action
}

// New returns an empty map.
func New() *Map {
	return &Map{bindings: make(map[Chord]action.Action)}
}

// FromBindings builds a map from chord strings to action ids.
//
// This is the shape a TOML keymap arrives in, and it mirrors
// `Keymap::from_string_map`. An error names the offending pair, because a
// keymap that half-loads is worse than one that refuses: the user would find out
// which shortcut broke by pressing it.
func FromBindings(pairs map[string]string) (*Map, error) {
	result := New()
	// Sorted so the reported error is the same on every run; a map iterates in
	// random order and an intermittent error message is a support burden.
	chords := make([]string, 0, len(pairs))
	for chord := range pairs {
		chords = append(chords, chord)
	}
	sort.Strings(chords)

	for _, raw := range chords {
		actionID := pairs[raw]

		chord, err := Parse(raw)
		if err != nil {
			return nil, err
		}
		resolved, err := action.Parse(actionID)
		if err != nil {
			return nil, fmt.Errorf("chord %q: %w", raw, err)
		}
		result.bindings[chord] = resolved
	}
	return result, nil
}

// Bind sets one binding, replacing any previous action for the same chord.
func (m *Map) Bind(chord Chord, a action.Action) {
	m.bindings[chord] = a
}

// Resolve returns the action bound to a chord.
func (m *Map) Resolve(chord Chord) (action.Action, bool) {
	if m == nil {
		return "", false
	}
	a, ok := m.bindings[chord]
	return a, ok
}

// Len returns the number of bindings.
func (m *Map) Len() int {
	if m == nil {
		return 0
	}
	return len(m.bindings)
}

// Bindings returns every binding, sorted by canonical chord.
//
// Sorted because this feeds help screens and state dumps, and an order that
// depends on map iteration makes two runs of the same program look different.
func (m *Map) Bindings() []Binding {
	if m == nil {
		return nil
	}
	out := make([]Binding, 0, len(m.bindings))
	for chord, a := range m.bindings {
		out = append(out, Binding{Chord: chord, Action: a})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Chord.String() < out[j].Chord.String()
	})
	return out
}
