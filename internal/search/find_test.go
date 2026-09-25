package search

import (
	"strings"
	"testing"
)

func spans(haystack string, matches []Match) []string {
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		out = append(out, haystack[match.Start:match.End])
	}
	return out
}

func TestFindAllIsCaseInsensitiveByDefault(t *testing.T) {
	haystack := "ab x AB y ab"
	matches := FindAll(haystack, "Ab", false, false)
	if len(matches) != 3 {
		t.Fatalf("matches = %v, esperado 3", spans(haystack, matches))
	}
	for _, got := range spans(haystack, matches) {
		if !strings.EqualFold(got, "ab") {
			t.Errorf("casou %q, esperado uma variação de \"ab\"", got)
		}
	}
}

func TestFindAllCaseSensitiveSkipsOtherCases(t *testing.T) {
	haystack := "Ab ab AB"
	matches := FindAll(haystack, "ab", true, false)
	if len(matches) != 1 {
		t.Fatalf("matches = %v, esperado 1", spans(haystack, matches))
	}
	if matches[0].Start != 3 {
		t.Errorf("offset = %d, esperado 3", matches[0].Start)
	}
}

// TestAccentFoldingFindsBothSpellings is the behaviour users notice: searching
// for the unaccented form finds the accented one.
func TestAccentFoldingFindsBothSpellings(t *testing.T) {
	haystack := "ação e acao"
	matches := FindAll(haystack, "acao", false, true)
	if len(matches) != 2 {
		t.Fatalf("matches = %v, esperado 2", spans(haystack, matches))
	}
	if got := spans(haystack, matches); got[0] != "ação" || got[1] != "acao" {
		t.Errorf("casou %v", got)
	}
}

// TestOffsetsIndexTheOriginalText: folding must not shift the reported range, or
// every caller that slices the text would cut in the wrong place.
func TestOffsetsIndexTheOriginalText(t *testing.T) {
	haystack := "café ção"
	matches := FindAll(haystack, "cao", false, true)
	if len(matches) != 1 {
		t.Fatalf("matches = %v, esperado 1", spans(haystack, matches))
	}
	if got := haystack[matches[0].Start:matches[0].End]; got != "ção" {
		t.Errorf("o intervalo [%d,%d) corta %q, esperado \"ção\"",
			matches[0].Start, matches[0].End, got)
	}
}

func TestCaseSensitiveWithAccentFoldingStillFoldsCase(t *testing.T) {
	// The reference folds accents after case normalisation, so "Arvore" case
	// sensitively matches "Árvore" but not "árvore".
	matches := FindAll("Árvore arvore", "Arvore", true, true)
	if len(matches) != 1 {
		t.Fatalf("matches = %v, esperado 1", matches)
	}
	if matches[0].Start != 0 {
		t.Errorf("offset = %d, esperado 0", matches[0].Start)
	}
}

func TestWholeWordSkipsMatchesInsideIdentifiers(t *testing.T) {
	text := "GUI UI UIKit use UI."
	options := Options{CaseSensitive: true, WholeWord: true}
	state := &State{Query: "UI", Options: options}
	state.Recompute(text)

	// " GUI ", " UI " and " UI." — not GUI, not UIKit. The reference counts two
	// because "GUI UI" and the trailing "UI." are the word-bounded ones.
	if len(state.Matches) != 2 {
		t.Fatalf("matches = %v", spans(text, state.Matches))
	}
	for _, match := range state.Matches {
		got := text[match.Start:match.End]
		if got != "UI" {
			t.Errorf("casou %q", got)
		}
		if !IsWholeWord(text, match.Start, match.End) {
			t.Errorf("match %q não é palavra isolada", got)
		}
	}
}

func TestWholeWordOffMatchesSubstrings(t *testing.T) {
	matches := FindAll("GUI UI", "UI", true, false)
	if len(matches) != 2 {
		t.Fatalf("matches = %v, esperado 2 (dentro de GUI e solto)", spans("GUI UI", matches))
	}
}

func TestIsWholeWordBoundaries(t *testing.T) {
	text := "ab cd ab"
	if !IsWholeWord(text, 0, 2) {
		t.Error("\"ab\" no início deveria ser palavra isolada")
	}
	if !IsWholeWord(text, 6, 8) {
		t.Error("\"ab\" no fim deveria ser palavra isolada")
	}
	if !IsWholeWord(text, 3, 5) {
		t.Error("\"cd\" deveria ser palavra isolada")
	}
}

func TestMatchesDoNotOverlap(t *testing.T) {
	// A one-character query must not report the same position twice.
	matches := FindAll("aaa", "a", true, false)
	if len(matches) != 3 {
		t.Fatalf("matches = %v, esperado 3", spans("aaa", matches))
	}
	for i := 1; i < len(matches); i++ {
		if matches[i].Start < matches[i-1].End {
			t.Errorf("match %d começa em %d, antes do fim do anterior em %d",
				i, matches[i].Start, matches[i-1].End)
		}
	}
}

func TestEmptyQueryMatchesNothing(t *testing.T) {
	if matches := FindAll("abc", "", false, false); len(matches) != 0 {
		t.Errorf("query vazia casou %v", matches)
	}
}

func TestRegexMatches(t *testing.T) {
	haystack := "a1 b22 c"
	state := &State{Query: `\d+`, Options: Options{UseRegex: true}}
	state.Recompute(haystack)
	if len(state.Matches) != 2 {
		t.Fatalf("matches = %v", spans(haystack, state.Matches))
	}
	if state.Matches[0].Start != 1 || state.Matches[1].Start != 4 {
		t.Errorf("offsets = %d e %d, esperado 1 e 4",
			state.Matches[0].Start, state.Matches[1].Start)
	}
}

// TestRegexErrorIsReportedNotThrown: a half-typed pattern is normal while
// someone is typing, so it becomes a message rather than a crash.
func TestRegexErrorIsReportedNotThrown(t *testing.T) {
	state := &State{Query: "([unclosed", Options: Options{UseRegex: true}}
	state.Recompute("texto")

	if state.RegexError == "" {
		t.Fatal("padrão inválido não produziu erro")
	}
	if len(state.Matches) != 0 {
		t.Errorf("padrão inválido casou %v", state.Matches)
	}
	if !strings.Contains(state.Status(), "regex error") {
		t.Errorf("status = %q, esperado mencionar o erro", state.Status())
	}
}

func TestNavigationWraps(t *testing.T) {
	state := &State{Query: "x"}
	state.Recompute("x x x")
	if len(state.Matches) != 3 {
		t.Fatalf("matches = %d, esperado 3", len(state.Matches))
	}
	if state.Current != 0 {
		t.Fatalf("current = %d, esperado 0", state.Current)
	}

	state.Next()
	state.Next()
	if state.Current != 2 {
		t.Fatalf("current = %d, esperado 2", state.Current)
	}
	state.Next()
	if state.Current != 0 {
		t.Errorf("Next deveria voltar ao início, current = %d", state.Current)
	}
	state.Prev()
	if state.Current != 2 {
		t.Errorf("Prev deveria voltar ao fim, current = %d", state.Current)
	}
}

func TestLabels(t *testing.T) {
	state := &State{}
	if state.Label() != "digite para buscar" {
		t.Errorf("label sem query = %q", state.Label())
	}

	state.Query = "zzz"
	state.Recompute("abc")
	if state.Label() != "0 matches" {
		t.Errorf("label sem matches = %q", state.Label())
	}

	state.Query = "x"
	state.Recompute("x x")
	if state.Label() != "1 / 2 matches" {
		t.Errorf("label = %q, esperado \"1 / 2 matches\"", state.Label())
	}
}

// TestCurrentIsClampedWhenMatchesShrink: recomputing with a shorter haystack must
// not leave the cursor pointing past the end of the list.
func TestCurrentIsClampedWhenMatchesShrink(t *testing.T) {
	state := &State{Query: "x"}
	state.Recompute("x x x x")
	state.Current = 3

	state.Recompute("x")
	if state.Current != 0 {
		t.Errorf("current = %d, esperado 0 após a lista encolher", state.Current)
	}
	if _, ok := state.CurrentMatch(); !ok {
		t.Error("CurrentMatch falhou após o clamp")
	}
}

func TestCurrentIsResetWhenNothingMatches(t *testing.T) {
	state := &State{Query: "zzz"}
	state.Current = 5
	state.Recompute("abc")
	if state.Current != 0 {
		t.Errorf("current = %d, esperado 0", state.Current)
	}
	if _, ok := state.CurrentMatch(); ok {
		t.Error("CurrentMatch devolveu algo sem matches")
	}
}

// TestFindsAcrossMultiByteText: the scan walks runes, so a query must match
// regardless of how many bytes its characters occupy.
func TestFindsAcrossMultiByteText(t *testing.T) {
	haystack := "日本語 e 🙂fim"
	for _, query := range []string{"本", "🙂", "fim", "語 e"} {
		matches := FindAll(haystack, query, true, false)
		if len(matches) != 1 {
			t.Errorf("query %q casou %d vezes", query, len(matches))
			continue
		}
		if got := haystack[matches[0].Start:matches[0].End]; got != query {
			t.Errorf("query %q casou %q", query, got)
		}
	}
}
