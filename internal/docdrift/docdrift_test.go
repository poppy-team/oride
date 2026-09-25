package docdrift

import (
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/ori-team/oride/internal/config"
	"github.com/ori-team/oride/internal/keymap"
)

var update = flag.Bool("update", false, "regenera os documentos derivados")

// keymapDoc is the canonical table, relative to the repository root.
const keymapDoc = "docs/ui-ux/keymap.md"

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("diretório de trabalho: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("não achei a raiz do repositório")
		}
		dir = parent
	}
}

// TestDocumentedKeymapMatchesTheBindings is the gate: the document and the code
// must not disagree in either direction.
//
// Comparing the whole generated table answers both questions at once — a binding
// missing from the document and a row in the document that no longer exists both
// show up as a difference. Reported as a per-line diff so the failure names the
// offending key instead of printing two walls of table.
func TestDocumentedKeymapMatchesTheBindings(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, filepath.FromSlash(keymapDoc))

	document, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("lendo %s: %v", keymapDoc, err)
	}

	want := KeymapTable(config.DefaultKeyBindings())

	if *update {
		updated, replaceErr := ReplaceGenerated(string(document), want)
		if replaceErr != nil {
			t.Fatalf("%v", replaceErr)
		}
		if writeErr := os.WriteFile(path, []byte(updated), 0o644); writeErr != nil {
			t.Fatalf("escrevendo %s: %v", keymapDoc, writeErr)
		}
		t.Logf("%s regenerado a partir de %d bindings", keymapDoc, len(config.DefaultKeyBindings()))
		return
	}

	got, err := GeneratedRegion(string(document))
	if err != nil {
		t.Fatalf("%v", err)
	}

	if normalise(got) != normalise(want) {
		t.Errorf("a tabela de %s divergiu das bindings.\n%s", keymapDoc, describeDiff(got, want))
		t.Errorf("regenerar: go test ./internal/docdrift/ -update")
	}
}

// TestDocumentedChordsAreCanonical proves the document speaks the same dialect
// the keymap does: every chord printed there must parse.
func TestDocumentedChordsAreCanonical(t *testing.T) {
	root := repoRoot(t)
	document, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(keymapDoc)))
	if err != nil {
		t.Fatalf("lendo %s: %v", keymapDoc, err)
	}

	region, err := GeneratedRegion(string(document))
	if err != nil {
		t.Fatalf("%v", err)
	}

	rows := tableRows(region)
	if len(rows) == 0 {
		t.Fatal("não achei nenhuma linha de tabela no documento")
	}
	for _, row := range rows {
		canonical := CanonicalChord(row[0])
		if _, parseErr := keymap.Parse(canonical); parseErr != nil {
			t.Errorf("tecla documentada não parseia: %q → %q: %v", row[0], canonical, parseErr)
		}
	}
}

// TestEveryBindingIsDocumented makes the count visible: a table that silently
// lost rows would otherwise still look plausible.
func TestEveryBindingIsDocumented(t *testing.T) {
	root := repoRoot(t)
	document, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(keymapDoc)))
	if err != nil {
		t.Fatalf("lendo %s: %v", keymapDoc, err)
	}
	region, err := GeneratedRegion(string(document))
	if err != nil {
		t.Fatalf("%v", err)
	}

	bindings := config.DefaultKeyBindings()
	if got := len(tableRows(region)); got != len(bindings) {
		t.Errorf("o documento tem %d teclas e as bindings têm %d", got, len(bindings))
	}
}

// TestRoundTripOfChordSpelling guards the display rewriting: a chord that does
// not survive pretty → canonical has silently become a different key.
func TestRoundTripOfChordSpelling(t *testing.T) {
	bindings := config.DefaultKeyBindings()
	for chord := range bindings {
		if got := CanonicalChord(PrettyChord(chord)); got != chord {
			t.Errorf("%q → %q → %q", chord, PrettyChord(chord), got)
		}
	}
}

// TestBacktickInABindingSurvivesTheDocument is the regression guard for the
// defect this generator already had once: a chord containing a backtick collided
// with the code-span delimiter, the row stopped parsing, and the binding vanished
// from the document while the table still looked complete.
func TestBacktickInABindingSurvivesTheDocument(t *testing.T) {
	bindings := config.DefaultKeyBindings()

	var withBacktick []string
	for chord := range bindings {
		if strings.Contains(chord, "`") {
			withBacktick = append(withBacktick, chord)
		}
	}
	if len(withBacktick) == 0 {
		t.Skip("nenhuma binding contém crase")
	}

	table := KeymapTable(bindings)
	documented := map[string]bool{}
	for _, row := range tableRows(table) {
		documented[CanonicalChord(row[0])] = true
	}
	for _, chord := range withBacktick {
		if !documented[chord] {
			t.Errorf("a tecla %q se perdeu na geração do documento", chord)
		}
	}
}

// TestCodeSpanRoundTrips covers the delimiter widening directly.
func TestCodeSpanRoundTrips(t *testing.T) {
	cases := []string{"a", "Ctrl+s", "Ctrl+`", "``", "`início", "fim`", "a`b`c"}
	for _, text := range cases {
		cell := CodeSpan(text)
		back, ok := StripCodeSpan(cell)
		if !ok {
			t.Errorf("CodeSpan(%q) = %q não foi reconhecido", text, cell)
			continue
		}
		if back != text {
			t.Errorf("CodeSpan(%q) = %q → %q", text, cell, back)
		}
	}
}

// TestReplaceGeneratedRefusesADocumentWithoutMarkers: appending instead of
// replacing would grow a duplicate table and still let the check pass.
func TestReplaceGeneratedRefusesADocumentWithoutMarkers(t *testing.T) {
	if _, err := ReplaceGenerated("sem marcador nenhum", "conteúdo"); err == nil {
		t.Error("um documento sem marcadores foi aceito")
	}
	if _, err := GeneratedRegion("sem marcador nenhum"); err == nil {
		t.Error("um documento sem marcadores foi lido")
	}
}

// tableRows reads the `| chord | action |` rows out of a table region.
//
// The cells are split by hand rather than matched with one regex, because a
// keybinding can contain a backtick and the delimiter length is therefore not
// fixed. A regex would either miss that row or need lookaround Go does not have.
func tableRows(region string) [][2]string {
	var out [][2]string

	for _, line := range strings.Split(region, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") || !strings.HasSuffix(line, "|") {
			continue
		}

		cells := strings.Split(line, "|")
		if len(cells) != 4 {
			continue
		}
		chord, okChord := StripCodeSpan(strings.TrimSpace(cells[1]))
		action, okAction := StripCodeSpan(strings.TrimSpace(cells[2]))
		if !okChord || !okAction {
			continue
		}
		out = append(out, [2]string{chord, action})
	}
	return out
}

// normalise trims each line so trailing whitespace never fails the comparison.
func normalise(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for index, line := range lines {
		lines[index] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

// describeDiff names the offending keys instead of printing two tables.
func describeDiff(got, want string) string {
	gotSet := map[string]bool{}
	for _, row := range tableRows(got) {
		gotSet[row[0]+" → "+row[1]] = true
	}
	wantSet := map[string]bool{}
	for _, row := range tableRows(want) {
		wantSet[row[0]+" → "+row[1]] = true
	}

	var missing, extra []string
	for row := range wantSet {
		if !gotSet[row] {
			missing = append(missing, row)
		}
	}
	for row := range gotSet {
		if !wantSet[row] {
			extra = append(extra, row)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)

	var out strings.Builder
	if len(missing) > 0 {
		out.WriteString("  nas bindings e fora do documento:\n")
		for _, row := range missing {
			out.WriteString("    " + row + "\n")
		}
	}
	if len(extra) > 0 {
		out.WriteString("  no documento e fora das bindings:\n")
		for _, row := range extra {
			out.WriteString("    " + row + "\n")
		}
	}
	return out.String()
}
