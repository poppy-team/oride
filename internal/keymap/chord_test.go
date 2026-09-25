package keymap

import (
	"strings"
	"testing"
)

func TestParseCanonicalForms(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"ctrl+s", "ctrl+s"},
		{"esc", "esc"},
		{"escape", "esc"},
		{"f12", "f12"},
		{"ctrl+shift+s", "ctrl+shift+s"},
		{"shift+left", "shift+left"},
		{"ctrl+pageup", "ctrl+pageup"},
		{"enter", "enter"},
		{"space", "space"},
		{"backspace", "backspace"},
		{"delete", "delete"},
	}

	for _, tc := range cases {
		parsed, err := Parse(tc.input)
		if err != nil {
			t.Errorf("Parse(%q): %v", tc.input, err)
			continue
		}
		if got := parsed.String(); got != tc.want {
			t.Errorf("Parse(%q).String() = %q, esperado %q", tc.input, got, tc.want)
		}
	}
}

// TestModifierOrderIsCanonical is the property that makes a chord a lookup key:
// the same binding typed in any order must produce the same string, or a help
// screen lists it twice and a config lookup misses.
func TestModifierOrderIsCanonical(t *testing.T) {
	variants := []string{"ctrl+shift+s", "shift+ctrl+s", "CTRL+SHIFT+S", " Ctrl + Shift + S "}
	first, err := Parse(variants[0])
	if err != nil {
		t.Fatalf("Parse(%q): %v", variants[0], err)
	}
	for _, variant := range variants[1:] {
		parsed, err := Parse(variant)
		if err != nil {
			t.Errorf("Parse(%q): %v", variant, err)
			continue
		}
		if parsed != first {
			t.Errorf("Parse(%q) = %v, esperado %v", variant, parsed, first)
		}
		if parsed.String() != first.String() {
			t.Errorf("Parse(%q).String() = %q, esperado %q", variant, parsed.String(), first.String())
		}
	}
}

// TestCmdAndSuperCollapseOntoCtrl mirrors the Rust mapping. Inventing a fourth
// modifier would make the same physical shortcut two different bindings
// depending on which token the user wrote.
func TestCmdAndSuperCollapseOntoCtrl(t *testing.T) {
	for _, token := range []string{"cmd", "super", "control", "ctrl"} {
		parsed, err := Parse(token + "+p")
		if err != nil {
			t.Fatalf("Parse(%q+p): %v", token, err)
		}
		if !parsed.Ctrl {
			t.Errorf("%q+p não marcou ctrl", token)
		}
		if got := parsed.String(); got != "ctrl+p" {
			t.Errorf("%q+p normalizou para %q, esperado \"ctrl+p\"", token, got)
		}
	}
}

func TestParseRejects(t *testing.T) {
	for _, input := range []string{
		"",
		"   ",
		"ctrl",
		"ctrl+",
		"ctrl+s+t",
		"ctrl+nonexistent_token",
		// Snake case is not a spelling the parser accepts; the canonical token
		// is "pagedown", with "pgup"/"pgdn" as documented aliases.
		"page_down",
	} {
		if _, err := Parse(input); err == nil {
			t.Errorf("Parse(%q) foi aceito, mas não é um chord válido", input)
		}
	}
}

func TestParseErrorNamesInput(t *testing.T) {
	_, err := Parse("ctrl+nope")
	if err == nil {
		t.Fatal("token desconhecido foi aceito")
	}
	if !strings.Contains(err.Error(), "ctrl+nope") {
		t.Errorf("erro %q não nomeia a entrada", err.Error())
	}
}

// TestSingleCharacterCodesAreTheirOwnToken covers the rule that lets every
// letter, digit and symbol work without a table entry for each.
func TestSingleCharacterCodesAreTheirOwnToken(t *testing.T) {
	for _, code := range []string{"a", "z", "0", "9", ";", "[", "]", "é"} {
		parsed, err := Parse(code)
		if err != nil {
			t.Errorf("Parse(%q): %v", code, err)
			continue
		}
		if parsed.Code != code {
			t.Errorf("Parse(%q).Code = %q", code, parsed.Code)
		}
	}
}

func TestHasModifierDistinguishesTypingFromCommands(t *testing.T) {
	plain := MustParse("a")
	if plain.HasModifier() {
		t.Error("'a' foi tratado como comando")
	}
	for _, withModifier := range []string{"ctrl+a", "alt+a", "shift+a"} {
		if !MustParse(withModifier).HasModifier() {
			t.Errorf("%q precisa ser um comando", withModifier)
		}
	}
}

func TestZeroChordIsUnset(t *testing.T) {
	var zero Chord
	if !zero.IsZero() {
		t.Error("chord zero não foi reconhecido como não definido")
	}
	if zero.HasModifier() {
		t.Error("chord zero tem modificador")
	}
}
