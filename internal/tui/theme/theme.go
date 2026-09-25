// Package theme resolves the configured colours into styles.
//
// It is the only place that knows a colour exists. Surfaces ask for a role —
// selection, gutter, status — and get a style, so a palette change does not reach
// into the code that draws.
package theme

import (
	"image/color"

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

// parse reads a colour from configuration.
//
// An empty value and the literal "reset" both mean "no colour", and both are
// normal in the shipped themes — passing either to the style library would be a
// value it cannot resolve, so they are filtered here rather than at every call.
func parse(value string) color.Color {
	switch value {
	case "", "reset", "none":
		return nil
	}
	return lipgloss.Color(value)
}
