package action

import (
	"fmt"
	"strings"
)

// ParseError says which id was rejected and what was expected.
type ParseError struct {
	ID string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("unknown action id %q", e.ID)
}

// aliases map historical and convenience spellings onto canonical ids.
//
// They exist because the Rust parser accepted them, so a configuration written
// against it keeps working after the port. Each one resolves to an id that is
// already in the canonical table — an alias never invents a new action.
var aliases = map[string]Action{
	"display_language": SelectLocale,
	"checkhealth":      HealthCheck,
	"health":           HealthCheck,
	"modal_mode":       ToggleModal,
	"tasks":            RunTasks,
}

// Parse resolves a stable id, accepting documented aliases.
//
// The id is trimmed but not case-folded: ids are lowercase by construction, and
// silently accepting "Save" would let a typo look like a valid config.
func Parse(id string) (Action, error) {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return "", ParseError{ID: id}
	}

	if canonical, ok := aliases[trimmed]; ok {
		return canonical, nil
	}
	candidate := Action(trimmed)
	if candidate.Valid() {
		return candidate, nil
	}
	return "", ParseError{ID: id}
}

// Valid reports whether a is one of the canonical ids.
func (a Action) Valid() bool {
	return index(a) >= 0
}

// String returns the id, which is the identity itself.
func (a Action) String() string { return string(a) }

// All returns the canonical table in declaration order.
//
// The slice is copied so a caller cannot reorder the table for everyone else.
func All() []Action {
	out := make([]Action, len(all))
	copy(out, all)
	return out
}

// Palette returns the curated command palette order.
func Palette() []Action {
	out := make([]Action, len(paletteOrder))
	copy(out, paletteOrder)
	return out
}

// index finds the position of a in the canonical table, or -1.
//
// Linear scan on purpose: the table is a hundred entries, the call is not on a
// hot path, and a map would need a second structure that can disagree with the
// slice. One table means one truth.
func index(a Action) int {
	if a == "" {
		return -1
	}
	for i, candidate := range all {
		if candidate == a {
			return i
		}
	}
	return -1
}
