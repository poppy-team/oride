// Package keymap resolves keystrokes to actions.
//
// A chord is parsed from a canonical string ("ctrl+shift+s") and normalised back
// to one, so a configuration file, a help screen and a lookup key all spell the
// same binding the same way. The Rust implementation needed three encodings of
// Ctrl+Shift+S because the terminal could report it three ways; that
// normalisation is the reason this parser is strict about its own form.
package keymap

import (
	"fmt"
	"strings"
)

// Chord is a parsed keystroke: modifiers plus a canonical code token.
//
// Code holds the token itself — "s", "enter", "f1", "left", "space" — rather
// than a numeric key code. That makes String and Parse inverse by construction:
// there is no second table to keep aligned with this one.
type Chord struct {
	Ctrl  bool
	Alt   bool
	Shift bool
	Code  string
}

// ParseError says what could not be parsed and why.
type ParseError struct {
	Input  string
	Reason string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("chord %q: %s", e.Input, e.Reason)
}

// codeAliases map every accepted spelling to its canonical token.
//
// Several spellings exist because the Rust parser accepted them, so a
// configuration written against it keeps working. `cmd` and `super` are not
// modifiers here: the Rust mapping collapsed them onto ctrl, and parity means
// collapsing them the same way rather than inventing a fourth modifier.
var codeAliases = map[string]string{
	"esc": "esc", "escape": "esc",
	"enter": "enter", "return": "enter",
	"backspace": "backspace", "bs": "backspace",
	"delete": "delete", "del": "delete",
	"tab":      "tab",
	"left":     "left",
	"right":    "right",
	"up":       "up",
	"down":     "down",
	"home":     "home",
	"end":      "end",
	"pageup":   "pageup",
	"pgup":     "pageup",
	"pagedown": "pagedown",
	"pgdn":     "pagedown",
	"pgdown":   "pagedown",
	"space":    "space",
	"f1":       "f1", "f2": "f2", "f3": "f3", "f4": "f4",
	"f5": "f5", "f6": "f6", "f7": "f7", "f8": "f8",
	"f9": "f9", "f10": "f10", "f11": "f11", "f12": "f12",
	"`": "`", "grave": "`", "backtick": "`",
	"\\": "\\", "backslash": "\\",
	"/": "/", "slash": "/",
	"\"": "\"", "quote": "\"", "doublequote": "\"", "dquote": "\"",
	"'": "'", "apostrophe": "'", "squote": "'",
	"?": "?", "question": "?", "questionmark": "?",
	"=": "=", "equals": "=", "equal": "=",
	"-": "-", "minus": "-", "dash": "-", "hyphen": "-",
}

// modifierAliases are the tokens that set a modifier instead of a code.
var modifierAliases = map[string]func(*Chord){
	"ctrl":    func(c *Chord) { c.Ctrl = true },
	"control": func(c *Chord) { c.Ctrl = true },
	"cmd":     func(c *Chord) { c.Ctrl = true },
	"super":   func(c *Chord) { c.Ctrl = true },
	"alt":     func(c *Chord) { c.Alt = true },
	"option":  func(c *Chord) { c.Alt = true },
	"opt":     func(c *Chord) { c.Alt = true },
	"shift":   func(c *Chord) { c.Shift = true },
}

// Parse reads a canonical chord string.
//
// ASCII is lower-cased before splitting, matching the Rust parser: a config may
// say "Ctrl+S" and mean the same binding as "ctrl+s". Non-ASCII characters are
// left alone, so a keyboard producing "é" keeps it.
func Parse(s string) (Chord, error) {
	normalized := strings.ToLower(strings.TrimSpace(s))
	if normalized == "" {
		return Chord{}, ParseError{Input: s, Reason: "vazio"}
	}

	var chord Chord
	codeSet := false

	for _, part := range strings.Split(normalized, "+") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if set, ok := modifierAliases[part]; ok {
			set(&chord)
			continue
		}
		if codeSet {
			// Two codes in one chord ("ctrl+s+t") is a typo, not a chord. The
			// Rust parser rejected it and so does this one, rather than
			// silently keeping the last.
			return Chord{}, ParseError{
				Input:  s,
				Reason: fmt.Sprintf("mais de um código: %q já definido, veio %q", chord.Code, part),
			}
		}
		code, err := parseCode(part)
		if err != nil {
			return Chord{}, ParseError{Input: s, Reason: err.Error()}
		}
		chord.Code = code
		codeSet = true
	}

	if !codeSet {
		return Chord{}, ParseError{Input: s, Reason: "sem código"}
	}
	return chord, nil
}

// MustParse is Parse for call sites where an invalid chord is a programming
// error — table literals and tests. It panics only on a mistake in this package,
// never on user input.
func MustParse(s string) Chord {
	chord, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return chord
}

func parseCode(token string) (string, error) {
	if canonical, ok := codeAliases[token]; ok {
		return canonical, nil
	}
	// A single character is its own token, which is what makes "a" and "1" work
	// without listing every letter.
	if runes := []rune(token); len(runes) == 1 {
		return token, nil
	}
	return "", fmt.Errorf("token desconhecido %q", token)
}

// String returns the canonical form: modifiers in a fixed order, then the code.
//
// The order is ctrl, alt, shift — not the order the user typed them, so
// "shift+ctrl+s" and "ctrl+shift+s" are the same binding and the help screen
// lists one entry.
func (c Chord) String() string {
	parts := make([]string, 0, 4)
	if c.Ctrl {
		parts = append(parts, "ctrl")
	}
	if c.Alt {
		parts = append(parts, "alt")
	}
	if c.Shift {
		parts = append(parts, "shift")
	}
	parts = append(parts, c.Code)
	return strings.Join(parts, "+")
}

// IsZero reports whether the chord carries no code.
func (c Chord) IsZero() bool { return c.Code == "" }

// HasModifier reports whether any modifier is set. A chord without one is
// printable text, which is how typing is distinguished from a command.
func (c Chord) HasModifier() bool { return c.Ctrl || c.Alt || c.Shift }
