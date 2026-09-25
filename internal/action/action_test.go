package action

import (
	"errors"
	"strings"
	"testing"
)

// TestIDsRoundTrip mirrors the Rust test of the same name: every id in the
// canonical table must parse back to itself. Without this, the table and the
// parser can disagree and a user's config silently stops matching a binding.
func TestIDsRoundTrip(t *testing.T) {
	for _, a := range All() {
		id := a.String()
		parsed, err := Parse(id)
		if err != nil {
			t.Errorf("id %q da tabela não parseia: %v", id, err)
			continue
		}
		if parsed != a {
			t.Errorf("id %q voltou como %q", id, parsed)
		}
	}
}

func TestTableHasNoDuplicates(t *testing.T) {
	seen := make(map[Action]bool, len(all))
	for _, a := range All() {
		if seen[a] {
			t.Errorf("ação %q aparece duas vezes na tabela", a)
		}
		seen[a] = true
	}
	if len(seen) != len(All()) {
		t.Fatalf("%d ids únicos para %d entradas", len(seen), len(All()))
	}
}

func TestTrailingWhitespaceIsNotAnAction(t *testing.T) {
	// Trimmed input is accepted; the point is that an empty or blank id is not.
	if _, err := Parse("   "); err == nil {
		t.Error("id em branco foi aceito")
	}
}

func TestParseRejectsUnknownAndIsCaseSensitive(t *testing.T) {
	// No padded whitespace here: that is tolerated on purpose and asserted by
	// TestParseToleratesSurroundingWhitespace. Listing it as rejected would make
	// the two tests contradict each other.
	for _, id := range []string{"does_not_exist", "", "Save", "SAVE", "MoveLeft", "move-left"} {
		if _, err := Parse(id); err == nil {
			t.Errorf("id %q foi aceito, mas não existe nessa forma", id)
		}
	}
}

func TestParseToleratesSurroundingWhitespace(t *testing.T) {
	for _, padded := range []string{"  move_left  ", "\tmove_left\n"} {
		parsed, err := Parse(padded)
		if err != nil {
			t.Fatalf("%q: espaços em volta deveriam ser tolerados: %v", padded, err)
		}
		if parsed != MoveLeftPlain {
			t.Errorf("%q resolveu para %q, esperado %q", padded, parsed, MoveLeftPlain)
		}
	}
}

func TestParseErrorNamesTheID(t *testing.T) {
	_, err := Parse("teleport")
	if err == nil {
		t.Fatal("id inexistente foi aceito")
	}
	var parseErr ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("erro não é ParseError: %T", err)
	}
	if parseErr.ID != "teleport" {
		t.Errorf("erro aponta %q, esperado \"teleport\"", parseErr.ID)
	}
	if !strings.Contains(err.Error(), "teleport") {
		t.Errorf("mensagem %q não nomeia o id", err.Error())
	}
}

func TestAliasesResolveToCanonicalIDs(t *testing.T) {
	for _, alias := range []string{"display_language", "checkhealth", "health", "modal_mode", "tasks"} {
		parsed, err := Parse(alias)
		if err != nil {
			t.Errorf("alias %q não parseia: %v", alias, err)
			continue
		}
		if !parsed.Valid() {
			t.Errorf("alias %q resolveu para %q, que não está na tabela", alias, parsed)
		}
		// The alias must not be an id of its own.
		if parsed.String() == alias {
			t.Errorf("alias %q resolveu para si mesmo; não é um alias", alias)
		}
	}
}

func TestPaletteIsASubsetOfTheTable(t *testing.T) {
	known := make(map[Action]bool, len(all))
	for _, a := range All() {
		known[a] = true
	}
	seen := map[Action]bool{}
	for _, a := range Palette() {
		if !known[a] {
			t.Errorf("palette lista %q, que está fora da tabela", a)
		}
		if seen[a] {
			t.Errorf("palette lista %q duas vezes", a)
		}
		seen[a] = true
	}
}

// TestAccessorsReturnCopies guards the table against a caller that sorts or
// truncates what it was handed.
func TestAccessorsReturnCopies(t *testing.T) {
	first := All()
	if len(first) == 0 {
		t.Fatal("tabela vazia")
	}
	original := first[0]
	first[0] = "mutated"
	if All()[0] != original {
		t.Error("All() devolveu a fatia interna")
	}

	palette := Palette()
	if len(palette) == 0 {
		t.Fatal("palette vazia")
	}
	paletteOriginal := palette[0]
	palette[0] = "mutated"
	if Palette()[0] != paletteOriginal {
		t.Error("Palette() devolveu a fatia interna")
	}
}

func TestEmptyActionIsInvalid(t *testing.T) {
	var zero Action
	if zero.Valid() {
		t.Error("o valor zero foi considerado uma ação válida")
	}
	if zero.String() != "" {
		t.Errorf("valor zero devolveu %q", zero.String())
	}
}
