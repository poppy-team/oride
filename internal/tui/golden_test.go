package tui

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ori-team/oride/internal/action"
	"github.com/ori-team/oride/internal/tui/focus"
	"github.com/ori-team/oride/internal/tui/overlay"
	"github.com/ori-team/oride/internal/tui/theme"
)

var update = flag.Bool("update", false, "regenera os golden frames")

// frameCase is one state worth pinning.
//
// The matrix is curated rather than cartesian: every combination of
// surface × width × colour would be a hundred files, and most of them would
// differ in nothing. Each case here exists because something about the frame
// changes in it.
type frameCase struct {
	name    string
	width   int
	height  int
	arrange func(*Model)
}

func frameCases() []frameCase {
	return []frameCase{
		{name: "baseline", width: 100, height: 30},

		// The three width classes, where the layout changes shape.
		{name: "compact-60", width: 60, height: 24},
		{name: "standard-80", width: 80, height: 24},
		{name: "wide-140", width: 140, height: 40},

		// The tree appears beside the editor, or floats over it.
		{name: "tree-side-by-side", width: 120, height: 30, arrange: func(m *Model) {
			m.application.ShowTree = true
		}},
		{name: "tree-overlaid", width: 60, height: 30, arrange: func(m *Model) {
			m.application.ShowTree = true
		}},

		// Focus on a surface other than the editor.
		{name: "focus-tree", width: 120, height: 30, arrange: func(m *Model) {
			m.application.ShowTree = true
			m.surface = focus.Tree
		}},

		// A status message is visible state.
		{name: "status-message", width: 100, height: 24, arrange: func(m *Model) {
			m.application.Status = "tarefa em background"
		}},

		// A frame at the smallest size the layout still treats as usable.
		{name: "minimum", width: 40, height: 6},

		// The overlays, which capture input and float above the surfaces.
		{name: "overlay-palette", width: 100, height: 30, arrange: func(m *Model) {
			m.overlay = overlay.Palette
		}},
		{name: "overlay-whichkey", width: 100, height: 30, arrange: func(m *Model) {
			m.overlay = overlay.WhichKey
		}},
		{name: "overlay-filtered", width: 100, height: 30, arrange: func(m *Model) {
			m.overlay = overlay.Palette
			m.filter = "undo"
		}},
		{name: "overlay-empty-filter", width: 100, height: 30, arrange: func(m *Model) {
			m.overlay = overlay.Palette
			m.filter = "zzzzzz"
		}},
		{name: "overlay-compact", width: 60, height: 20, arrange: func(m *Model) {
			m.overlay = overlay.Help
		}},

		// The search bar, closed and open, with and without the replacement field.
		{name: "find-bar", width: 100, height: 24, arrange: func(m *Model) {
			m.application.Store.OpenEmpty()
			_ = m.application.Apply(action.Find)
			m.application.Find.Query = "alfa"
			m.application.Find.Matches = nil
		}},
		{name: "find-bar-replace", width: 100, height: 24, arrange: func(m *Model) {
			m.application.Store.OpenEmpty()
			_ = m.application.Apply(action.Find)
			_ = m.application.Apply(action.Replace)
			m.application.Find.Query = "alfa"
			m.application.Find.Replace = "ALFA"
		}},
		{name: "find-matches", width: 100, height: 24, arrange: func(m *Model) {
			m.application.Store.OpenEmpty()
			document, err := m.application.Store.Active()
			if err == nil {
				_ = document.InsertText("alfa beta\ngama alfa\nbeta alfa\n")
			}
			_ = m.application.Apply(action.Find)
			m.application.Find.Query = "alfa"
			m.application.Find.Recompute(documentText(m))
			_ = m.application.Apply(action.FindNext)
		}},

		{name: "find-bar-invalid-regex", width: 100, height: 24, arrange: func(m *Model) {
			m.application.Store.OpenEmpty()
			_ = m.application.Apply(action.Find)
			m.application.Find.Query = "[sem fechar"
			m.application.Find.Options.UseRegex = true
			m.application.Find.Recompute("alfa")
		}},
	}
}

// TestGoldenFrames pins the rendering.
//
// The rule from the Bubble Tea discipline: a frame and the code that draws it
// change in the same commit. Regenerating is deliberate and shows up in the diff,
// which is the point — a golden file updated without being read is just a
// snapshot of whatever the code happens to do.
func TestGoldenFrames(t *testing.T) {
	for _, testCase := range frameCases() {
		t.Run(testCase.name, func(t *testing.T) {
			// Golden frames are generated with colour suppressed. A frame full of
			// escape sequences is unreadable in a diff, and the point of a golden
			// file is that a human reviews the change; colour has its own test
			// below, which needs no golden.
			model := sized(t, newModel(t, nil), testCase.width, testCase.height).WithProfile(theme.NoColor)
			if testCase.arrange != nil {
				testCase.arrange(&model)
			}

			got := model.Frame()
			path := filepath.Join("testdata", "frames", testCase.name+".golden")

			if *update {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatalf("criando testdata: %v", err)
				}
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatalf("escrevendo %s: %v", path, err)
				}
				return
			}

			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("lendo %s: %v\nregenere: go test ./internal/tui/ -update", path, err)
			}
			if got != string(want) {
				t.Errorf("frame divergiu de %s.\n%s\nregenere: go test ./internal/tui/ -update",
					path, describeFrameDiff(string(want), got))
			}
		})
	}
}

// TestGoldenCaseNamesAreUnique: two cases sharing a name would silently overwrite
// each other, and the matrix would look complete while missing one.
func TestGoldenCaseNamesAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, testCase := range frameCases() {
		if seen[testCase.name] {
			t.Errorf("nome de caso repetido: %s", testCase.name)
		}
		seen[testCase.name] = true
	}
}

// TestGoldenFramesCoverTheWidthClasses keeps the matrix honest: the gate names
// compact, standard and wide, and a case that quietly lost its width would leave
// a class uncovered.
func TestGoldenFramesCoverTheWidthClasses(t *testing.T) {
	classes := map[string]bool{}
	for _, testCase := range frameCases() {
		classes[className(testCase.width)] = true
	}
	for _, want := range []string{"compact", "standard", "wide"} {
		if !classes[want] {
			t.Errorf("nenhum golden frame na classe %s", want)
		}
	}
}

func className(width int) string {
	switch {
	case width < 80:
		return "compact"
	case width < 120:
		return "standard"
	default:
		return "wide"
	}
}

// TestEveryGoldenFileHasACase is the other direction: a file left behind by a
// removed case is a frame nothing regenerates and nobody notices is stale.
func TestEveryGoldenFileHasACase(t *testing.T) {
	known := map[string]bool{}
	for _, testCase := range frameCases() {
		known[testCase.name+".golden"] = true
	}

	entries, err := os.ReadDir(filepath.Join("testdata", "frames"))
	if err != nil {
		t.Skip("ainda não há golden frames")
	}
	for _, entry := range entries {
		if !known[entry.Name()] {
			t.Errorf("%s não corresponde a nenhum caso: apague-o ou declare o caso", entry.Name())
		}
	}
}

// describeFrameDiff names the first differing line instead of printing two frames,
// because a 30-line dump buries the one row that moved.
func describeFrameDiff(want, got string) string {
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")

	for index := range max(len(wantLines), len(gotLines)) {
		wantLine := lineAt(wantLines, index)
		gotLine := lineAt(gotLines, index)
		if wantLine != gotLine {
			return fmt.Sprintf("  primeira linha diferente, %d:\n    esperado: %q\n    obtido:   %q",
				index+1, wantLine, gotLine)
		}
	}
	return "  os frames diferem apenas em comprimento"
}

// lineAt reads a line, treating a missing one as empty so a frame that lost a row
// reports the row rather than an out-of-range panic.
func lineAt(lines []string, index int) string {
	if index < len(lines) {
		return lines[index]
	}
	return ""
}

// documentText is the active buffer's text, for the cases that need to search it.
func documentText(m *Model) string {
	document, err := m.application.Store.Active()
	if err != nil {
		return ""
	}
	return document.Buffer().String()
}
