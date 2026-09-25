// Package theme resolves the configured colours into styles.
//
// It is the only place that knows a colour exists. Surfaces ask for a role —
// selection, gutter, status — and get a style, so a palette change does not reach
// into the code that draws.
package theme

import (
	"image/color"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/ori-team/oride/internal/config"
)

// Profile is what the terminal can render.
//
// Ordered from least to most capable so a comparison answers "does this terminal
// support at least that", which is how every colour decision is made.
type Profile int

// The profiles, matching docs/ui-ux/degradation.md.
const (
	// NoColor renders nothing: every non-primitive token resolves to no styling
	// at all. A theme that claims to suppress colour and paints one is an error.
	NoColor Profile = iota
	ANSI16
	ANSI256
	TrueColor
)

// Theme resolves colours into styles for one profile.
type Theme struct {
	profile Profile
	ui      config.UIConfig
	syntax  config.SyntaxColors
}

// New builds a theme.
//
// The profile is a parameter rather than read from the environment: a test must be
// able to ask for NoColor without writing to the process environment, which is
// shared state.
func New(ui config.UIConfig, syntax config.SyntaxColors, profile Profile) Theme {
	return Theme{profile: profile, ui: ui, syntax: syntax}
}

// Profile reports what this theme renders for.
func (t Theme) Profile() Profile { return t.profile }

// Selection marks selected text.
func (t Theme) Selection() lipgloss.Style {
	return t.style(t.ui.Background, t.ui.CursorBG, t.ui.CursorFG, false)
}

// Gutter dims the line numbers.
func (t Theme) Gutter() lipgloss.Style {
	return t.style(t.ui.Background, "", t.ui.LineNumber, false)
}

// Status is the bottom bar.
func (t Theme) Status() lipgloss.Style {
	return t.style(t.ui.StatusBG, "", t.ui.StatusFG, false)
}

// StatusDirty is the bottom bar when the document has unsaved edits.
func (t Theme) StatusDirty() lipgloss.Style {
	return t.style(t.ui.StatusBG, "", t.ui.StatusDirty, true)
}

// Comment is the syntax role the surfaces use first.
//
// One role for now rather than twenty: the mapping from a chroma token to a role
// arrives with the highlighting engine, and declaring nineteen unused accessors
// would be abstraction that does not pay.
func (t Theme) Comment() lipgloss.Style {
	return t.style(t.ui.Background, "", t.syntax.Comment, false)
}

// style builds a style, resolving nothing when colour is suppressed.
func (t Theme) style(background, foreground, text string, bold bool) lipgloss.Style {
	style := lipgloss.NewStyle()
	if t.profile == NoColor {
		// The zero style applies no attribute, so the text renders as-is. This is
		// the whole no-colour guarantee, and it lives in one place.
		return style
	}

	if c := parse(background); c != nil {
		style = style.Background(c)
	}
	if c := parse(foreground); c != nil {
		style = style.Foreground(c)
	}
	if c := parse(text); c != nil {
		style = style.Foreground(c)
	}
	if bold {
		style = style.Bold(true)
	}
	return style
}

// ansiNames maps the colour names a configuration may use onto ANSI indices.
//
// The style library resolves hex and ANSI numbers, and silently ignores a name it
// does not know. That silence is what made an earlier version of this package ship
// a monochrome editor while every test passed: the tests used hex, the shipped
// configuration uses names, and an unresolved name is a no-op rather than an
// error.
var ansiNames = map[string]int{
	"black": 0, "red": 1, "green": 2, "yellow": 3,
	"blue": 4, "magenta": 5, "cyan": 6, "white": 7,
	"brightblack": 8, "darkgray": 8, "darkgrey": 8, "gray": 8, "grey": 8,
	"brightred": 9, "brightgreen": 10, "brightyellow": 11,
	"brightblue": 12, "brightmagenta": 13, "brightcyan": 14, "brightwhite": 15,
}

// parse reads a colour from configuration.
//
// An empty value and the literal "reset" both mean "no colour", and both are
// normal in the shipped themes. A recognised name becomes its ANSI index, so the
// value the style library receives is one it can actually render.
func parse(value string) color.Color {
	switch value {
	case "", "reset", "none":
		return nil
	}
	if index, known := ansiNames[strings.ToLower(strings.TrimSpace(value))]; known {
		return lipgloss.Color(strconv.Itoa(index))
	}
	return lipgloss.Color(value)
}
