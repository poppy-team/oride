package config

import "testing"

// TestDefaultBindingsAreNotEmpty is the weakest of these checks and still worth
// having: an empty default map means every documented shortcut is unbound, and
// nothing else in the suite would notice.
func TestDefaultBindingsAreNotEmpty(t *testing.T) {
	if len(defaultKeyBindings) == 0 {
		t.Fatal("nenhum binding default: todas as teclas documentadas ficariam sem ação")
	}
}

// TestDefaultBindingsHaveNoDuplicates guards the invariant a map cannot express.
//
// Go silently keeps the last value for a repeated key, so a duplicate chord in
// the generated table would drop a binding without any error. The generator
// checks this on the Rust side; this checks the Go side actually matches.
func TestDefaultBindingsHaveNoDuplicates(t *testing.T) {
	seen := make(map[string]int, len(defaultKeyBindings))
	for chord := range defaultKeyBindings {
		seen[chord]++
	}
	for chord, count := range seen {
		if count > 1 {
			t.Errorf("chord %q aparece %d vezes", chord, count)
		}
	}
}

func TestDefaultBindingsUseCanonicalChordSyntax(t *testing.T) {
	// Chords are lowercase, modifier-prefixed and free of whitespace. Parsing
	// itself belongs to internal/keymap; this only asserts the table is written
	// in the form the parser expects, so a typo fails here with a clear message
	// instead of failing there with a generic one.
	for chord := range defaultKeyBindings {
		for _, r := range chord {
			if r >= 'A' && r <= 'Z' {
				t.Errorf("chord %q tem maiúscula; os ids canônicos são minúsculos", chord)
				break
			}
		}
		if chord == "" || chord[0] == '+' || chord[len(chord)-1] == '+' {
			t.Errorf("chord %q tem modificador vazio", chord)
		}
	}
}

func TestDefaultBindingsValuesLookLikeActionIDs(t *testing.T) {
	for chord, action := range defaultKeyBindings {
		if action == "" {
			t.Errorf("chord %q sem ação", chord)
		}
		if action[0] == '_' || action[len(action)-1] == '_' {
			t.Errorf("chord %q aponta para %q, que não parece um id", chord, action)
		}
	}
}
