// Package search finds text in a buffer and in a project.
//
// The in-buffer matcher is here because its semantics are subtle in ways that
// are invisible until they are wrong: case folding, accent folding, whole-word
// boundaries, and byte offsets into the original text rather than into a folded
// copy.
package search

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Match is a hit, as a half-open byte range in the original text.
//
// Byte offsets, not rune or folded-string offsets: a caller uses these to move a
// caret and to slice the text, and folding changes neither.
type Match struct {
	Start int
	End   int
}

// Options configure a search.
type Options struct {
	CaseSensitive bool
	// IgnoreAccents folds accented letters onto their base form, so "funcao"
	// finds "função". Literal searches only — a regex is the user's own
	// expression and folding it would change what it means.
	IgnoreAccents bool
	// WholeWord requires a non-word character (or an edge) on both sides.
	WholeWord bool
	// UseRegex interprets the query as a regular expression.
	UseRegex bool
}

// State is the in-buffer search state.
type State struct {
	Query   string
	Replace string
	Options Options
	Matches []Match
	Current int
	// RegexError holds a compilation failure, which is reported rather than
	// thrown: a half-typed pattern is normal while someone is typing.
	RegexError  string
	ShowReplace bool
}

// NewState returns the search state the product ships with.
//
// Accent folding starts on: a Brazilian user searching "funcao" means "função",
// and having to enable that is friction for the common case. Case sensitivity
// starts off for the same reason in reverse.
func NewState() State {
	return State{
		Options: Options{IgnoreAccents: true},
	}
}

// Recompute finds every match for the current query.
func (s *State) Recompute(haystack string) {
	s.RegexError = ""

	var matches []Match
	if s.Options.UseRegex {
		found, err := FindAllRegex(haystack, s.Query, s.Options.CaseSensitive)
		if err != nil {
			s.RegexError = err.Error()
			matches = nil
		} else {
			matches = found
		}
	} else {
		matches = FindAll(haystack, s.Query, s.Options.CaseSensitive, s.Options.IgnoreAccents)
	}

	if s.Options.WholeWord {
		kept := matches[:0]
		for _, match := range matches {
			if IsWholeWord(haystack, match.Start, match.End) {
				kept = append(kept, match)
			}
		}
		matches = kept
	}

	s.Matches = matches
	if len(s.Matches) == 0 {
		s.Current = 0
		return
	}
	if s.Current >= len(s.Matches) {
		s.Current = len(s.Matches) - 1
	}
}

// CurrentMatch returns the match under the cursor.
func (s *State) CurrentMatch() (Match, bool) {
	if s.Current < 0 || s.Current >= len(s.Matches) {
		return Match{}, false
	}
	return s.Matches[s.Current], true
}

// Next advances to the next match, wrapping.
func (s *State) Next() (Match, bool) {
	if len(s.Matches) == 0 {
		return Match{}, false
	}
	s.Current = (s.Current + 1) % len(s.Matches)
	return s.CurrentMatch()
}

// Prev moves to the previous match, wrapping.
func (s *State) Prev() (Match, bool) {
	if len(s.Matches) == 0 {
		return Match{}, false
	}
	if s.Current == 0 {
		s.Current = len(s.Matches) - 1
	} else {
		s.Current--
	}
	return s.CurrentMatch()
}

// Label is the short count shown in the find bar.
func (s *State) Label() string {
	if s.Query == "" {
		return "digite para buscar"
	}
	if len(s.Matches) == 0 {
		return "0 matches"
	}
	return fmt.Sprintf("%d / %d matches", s.Current+1, len(s.Matches))
}

// Status is the one-line message the status bar shows.
func (s *State) Status() string {
	if s.RegexError != "" {
		return fmt.Sprintf("find: regex error — %s", s.RegexError)
	}
	if s.Query == "" {
		return "find · Alt+C/A/W/R opções · Esc fecha"
	}
	return s.Label()
}

// IsWholeWord reports whether the range is bounded by non-word characters.
func IsWholeWord(haystack string, start, end int) bool {
	if start > end || end > len(haystack) {
		return false
	}
	leftOK := true
	if start > 0 {
		before := rune(0)
		for _, r := range haystack[:start] {
			before = r
		}
		leftOK = !isWordChar(before)
	}
	rightOK := true
	if end < len(haystack) {
		for _, r := range haystack[end:] {
			rightOK = !isWordChar(r)
			break
		}
	}
	return leftOK && rightOK
}

func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// FindAll locates a literal query, optionally ignoring case and accents.
//
// Folding is done on both sides and the scan walks the original text in step
// with the folded one, so a match reports offsets into the text the caller holds
// — not into a folded copy whose indices would mean nothing.
func FindAll(haystack, query string, caseSensitive, ignoreAccents bool) []Match {
	if query == "" {
		return nil
	}

	foldedQuery := foldString(query, caseSensitive, ignoreAccents)
	if foldedQuery == "" {
		return nil
	}
	queryRunes := []rune(foldedQuery)

	type foldedRune struct {
		start  int
		end    int
		folded rune
	}
	original := make([]foldedRune, 0, len(haystack))
	for index, r := range haystack {
		original = append(original, foldedRune{
			start:  index,
			end:    index + len(string(r)),
			folded: foldRune(r, caseSensitive, ignoreAccents),
		})
	}

	var out []Match
	total := len(original)
	if total < len(queryRunes) {
		return out
	}

	for i := 0; i+len(queryRunes) <= total; {
		matched := true
		for k, want := range queryRunes {
			if original[i+k].folded != want {
				matched = false
				break
			}
		}
		if matched {
			out = append(out, Match{
				Start: original[i].start,
				End:   original[i+len(queryRunes)-1].end,
			})
			// Non-overlapping: a query of one character must not match the same
			// position twice.
			i += max(len(queryRunes), 1)
			continue
		}
		i++
	}
	return out
}

// FindAllRegex locates matches of a regular expression.
func FindAllRegex(haystack, pattern string, caseSensitive bool) ([]Match, error) {
	if pattern == "" {
		return nil, nil
	}
	if !caseSensitive {
		pattern = "(?i)" + pattern
	}
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	var out []Match
	for _, index := range compiled.FindAllStringIndex(haystack, -1) {
		out = append(out, Match{Start: index[0], End: index[1]})
	}
	return out, nil
}

// foldString folds every rune of a string.
func foldString(text string, caseSensitive, ignoreAccents bool) string {
	var out strings.Builder
	out.Grow(len(text))
	for _, r := range text {
		out.WriteRune(foldRune(r, caseSensitive, ignoreAccents))
	}
	return out.String()
}

// accentFolding maps accented letters onto their base form.
//
// Both cases are listed because the table runs before case folding is applied
// when the search is case sensitive, and after otherwise — the reference behaves
// this way and a case probe pins it.
var accentFolding = map[rune]rune{
	'á': 'a', 'à': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'å': 'a',
	'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
	'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i',
	'ó': 'o', 'ò': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o',
	'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u',
	'ç': 'c',
	'ñ': 'n',
	'ý': 'y', 'ÿ': 'y',
	'Á': 'A', 'À': 'A', 'Â': 'A', 'Ã': 'A', 'Ä': 'A', 'Å': 'A',
	'É': 'E', 'È': 'E', 'Ê': 'E', 'Ë': 'E',
	'Í': 'I', 'Ì': 'I', 'Î': 'I', 'Ï': 'I',
	'Ó': 'O', 'Ò': 'O', 'Ô': 'O', 'Õ': 'O', 'Ö': 'O',
	'Ú': 'U', 'Ù': 'U', 'Û': 'U', 'Ü': 'U',
	'Ç': 'C',
	'Ñ': 'N',
	'Ý': 'Y',
}

// foldRune lower-cases and, when asked, strips an accent.
func foldRune(r rune, caseSensitive, ignoreAccents bool) rune {
	normalized := r
	if !caseSensitive {
		normalized = unicode.ToLower(r)
	}
	if !ignoreAccents {
		return normalized
	}
	if base, ok := accentFolding[normalized]; ok {
		return base
	}
	return normalized
}
