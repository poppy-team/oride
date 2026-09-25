package conformance

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/ori-team/oride/internal/buffer"
)

// TestBufferMatchesTheOracle compares the Go buffer against the oracle's dump.
//
// It runs over every case, but only asserts on frames where the document was
// never edited — `version == 0` means no edit reached the buffer, so its text is
// exactly the file the case declared. That rule lets the test grow with the case
// suite instead of needing a hand-picked list, and it refuses to compare an
// edited document against the original file, which would be a false failure.
//
// Four things are verified at once, and all four are places a port silently
// drifts: byte count, scalar count, line count, and the byte↔caret conversion.
func TestBufferMatchesTheOracle(t *testing.T) {
	oracle := oracleForTest(t)
	paths, err := LoadCases(casesDir(t))
	if err != nil {
		t.Fatalf("descobrindo casos: %v", err)
	}

	compared := 0

	for _, path := range paths {
		loaded, err := LoadCase(path)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if loaded.Open == "" {
			continue
		}
		content, ok := fileContent(loaded, loaded.Open)
		if !ok {
			t.Fatalf("%s: `open` aponta para %q, que o caso não declara em `files`",
				filepath.Base(path), loaded.Open)
		}

		report := runOracle(t, oracle, path)
		for _, frame := range report.Frames {
			document, ok := documentOf(frame)
			if !ok || documentVersion(document) != 0 {
				continue
			}

			name := fmt.Sprintf("%s/passo-%d", filepath.Base(path), frame.AfterStep)
			t.Run(name, func(t *testing.T) {
				assertBufferMatches(t, document, content)
			})
			compared++
		}
	}

	if compared == 0 {
		t.Fatal("nenhum quadro comparável: o teste passaria sem medir nada")
	}
}

func assertBufferMatches(t *testing.T, document map[string]any, content string) {
	t.Helper()

	b := buffer.FromText(content)

	if want := intAt(t, document, "bytes"); b.LenBytes() != want {
		t.Errorf("LenBytes() = %d, o oráculo diz %d", b.LenBytes(), want)
	}
	if want := intAt(t, document, "chars"); b.CharCount() != want {
		t.Errorf("CharCount() = %d, o oráculo diz %d", b.CharCount(), want)
	}
	if want := intAt(t, document, "lines"); b.LineCount() != want {
		t.Errorf("LineCount() = %d, o oráculo diz %d", b.LineCount(), want)
	}
	if got := b.String(); got != content {
		t.Error("o texto do buffer difere do arquivo declarado pelo caso")
	}

	// The caret is derived from the selection head, so this verifies the
	// byte↔caret conversion against real oracle output rather than a fixture.
	selection, ok := document["selection"].(map[string]any)
	if !ok {
		t.Fatal("documento sem objeto `selection`: o dump mudou de forma")
	}
	head := intAt(t, selection, "head")

	caret, err := b.ByteToCaret(buffer.Offset(head))
	if err != nil {
		t.Fatalf("ByteToCaret(%d): %v", head, err)
	}
	oracleCaret, ok := document["caret"].(map[string]any)
	if !ok {
		t.Fatal("documento sem objeto `caret`: o dump mudou de forma")
	}
	wantLine, wantColumn := intAt(t, oracleCaret, "line"), intAt(t, oracleCaret, "column")
	if caret.Line != wantLine || caret.Column != wantColumn {
		t.Errorf("ByteToCaret(%d) = {%d %d}, o oráculo diz {%d %d}",
			head, caret.Line, caret.Column, wantLine, wantColumn)
	}

	// And the inverse, which is what every edit path depends on.
	back, err := b.CaretToByte(caret)
	if err != nil {
		t.Fatalf("CaretToByte(%+v): %v", caret, err)
	}
	if int(back) != head {
		t.Errorf("CaretToByte(%+v) = %d, o oráculo diz %d", caret, back, head)
	}
}

// fileContent finds a file the case declared.
func fileContent(c *Case, path string) (string, bool) {
	for _, file := range c.Files {
		if file.Path == path {
			return file.Content, true
		}
	}
	return "", false
}

func documentOf(frame Frame) (map[string]any, bool) {
	state, ok := frame.State.(map[string]any)
	if !ok {
		return nil, false
	}
	document, ok := state["document"].(map[string]any)
	return document, ok
}

func documentVersion(document map[string]any) int64 {
	number, ok := document["version"].(json.Number)
	if !ok {
		return -1
	}
	parsed, err := number.Int64()
	if err != nil {
		return -1
	}
	return parsed
}

// intAt reads an integer field.
//
// It fails rather than defaulting: a field that silently disappeared would make
// the comparison vacuous, and a vacuous comparison reports parity nobody
// measured.
func intAt(t *testing.T, object map[string]any, key string) int {
	t.Helper()

	value, ok := object[key]
	if !ok {
		t.Fatalf("campo %q ausente: o dump mudou de forma", key)
	}
	number, ok := value.(json.Number)
	if !ok {
		t.Fatalf("campo %q não é número: %T", key, value)
	}
	parsed, err := number.Int64()
	if err != nil {
		t.Fatalf("campo %q: %v", key, err)
	}
	return int(parsed)
}
