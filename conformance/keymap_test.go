package conformance

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/ori-team/oride/internal/config"
	"github.com/ori-team/oride/internal/keymap"
)

// TestDefaultKeymapMatchesTheOracle is the first assertion in this repository
// where the Go implementation is compared against the Rust oracle rather than
// the oracle against itself.
//
// It verifies three things at once, because the oracle's dump carries the
// resolved keymap:
//
//   - the Go chord parser canonicalises like the Rust one, so the same binding
//     has the same spelling on both sides;
//   - the generated default key binding table matches the shipped one;
//   - every action id in that table exists in the Go action table.
//
// The fixture is the oracle's own output, so this fails the moment either side
// drifts — no hand-maintained expected list to go stale.
func TestDefaultKeymapMatchesTheOracle(t *testing.T) {
	oracle := oracleForTest(t)
	paths, err := LoadCases(casesDir(t))
	if err != nil || len(paths) == 0 {
		t.Fatalf("descobrindo casos: %v (%d casos)", err, len(paths))
	}

	// Any case works: the effective config, and therefore the keymap, is in
	// every frame regardless of what the case exercises.
	report := runOracle(t, oracle, paths[0])

	expected, err := oracleKeymap(report)
	if err != nil {
		t.Fatalf("lendo o keymap do relatório: %v", err)
	}
	if len(expected) == 0 {
		t.Fatal("o oráculo não reportou nenhum binding: o dump mudou de forma")
	}

	built, err := keymap.FromBindings(config.DefaultKeyBindings())
	if err != nil {
		t.Fatalf("construindo o keymap Go a partir dos defaults: %v", err)
	}

	actual := make(map[string]string, built.Len())
	for _, binding := range built.Bindings() {
		actual[binding.Chord.String()] = binding.Action.String()
	}

	if len(actual) != len(expected) {
		t.Errorf("keymap com %d bindings, o oráculo tem %d", len(actual), len(expected))
	}

	for chord, wantAction := range expected {
		gotAction, ok := actual[chord]
		if !ok {
			t.Errorf("chord %q existe no oráculo e não no Go", chord)
			continue
		}
		if gotAction != wantAction {
			t.Errorf("chord %q mapeia para %q, o oráculo diz %q", chord, gotAction, wantAction)
		}
	}
	for chord := range actual {
		if _, ok := expected[chord]; !ok {
			t.Errorf("chord %q existe no Go e não no oráculo", chord)
		}
	}
}

// oracleKeymap extracts the resolved keymap from a report's last frame.
func oracleKeymap(report *Report) (map[string]string, error) {
	if len(report.Frames) == 0 {
		return nil, fmt.Errorf("relatório sem quadros")
	}
	state, ok := report.Frames[len(report.Frames)-1].State.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("estado não é objeto: %T", report.Frames[len(report.Frames)-1].State)
	}
	configValue, ok := state["config"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("estado sem objeto 'config'")
	}
	keys, ok := configValue["keys"].([]any)
	if !ok {
		return nil, fmt.Errorf("config sem lista 'keys'")
	}

	out := make(map[string]string, len(keys))
	for i, entry := range keys {
		binding, ok := entry.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("binding %d não é objeto: %T", i, entry)
		}
		chord, _ := binding["chord"].(string)
		actionID, _ := binding["action"].(string)
		if chord == "" || actionID == "" {
			return nil, fmt.Errorf("binding %d sem chord ou action: %v", i, binding)
		}
		out[chord] = actionID
	}
	return out, nil
}

// TestOracleKeymapExtractionIsNotEmpty guards the helper above: if the dump
// changes shape, this test fails here with a clear message instead of making
// the parity test pass over an empty map.
func TestOracleKeymapExtractionIsNotEmpty(t *testing.T) {
	oracle := oracleForTest(t)
	paths, err := LoadCases(casesDir(t))
	if err != nil || len(paths) == 0 {
		t.Fatalf("descobrindo casos: %v", err)
	}

	report := runOracle(t, oracle, paths[0])
	extracted, err := oracleKeymap(report)
	if err != nil {
		t.Fatalf("extraindo keymap: %v", err)
	}
	if len(extracted) < 50 {
		t.Fatalf("só %d bindings extraídos de %s: o dump provavelmente mudou de forma",
			len(extracted), filepath.Base(paths[0]))
	}
}

// TestGoKeymapIsUsableWithoutTheOracle keeps the Go side verifiable on a machine
// that cannot build Rust: the same construction is asserted here, so a broken
// default table fails everywhere and not only where the oracle happens to exist.
func TestGoKeymapIsUsableWithoutTheOracle(t *testing.T) {
	built, err := keymap.FromBindings(config.DefaultKeyBindings())
	if err != nil {
		t.Fatalf("construindo o keymap Go: %v", err)
	}
	if built.Len() != len(config.DefaultKeyBindings()) {
		t.Fatalf("keymap com %d bindings para %d pares: algum chord colidiu",
			built.Len(), len(config.DefaultKeyBindings()))
	}

	// A spot check that the table is wired end to end.
	chord, err := keymap.Parse("ctrl+s")
	if err != nil {
		t.Fatalf("Parse(\"ctrl+s\"): %v", err)
	}
	got, ok := built.Resolve(chord)
	if !ok {
		t.Fatal("ctrl+s não resolve para nenhuma ação")
	}
	if got.String() != "save" {
		t.Errorf("ctrl+s resolve para %q, esperado \"save\"", got)
	}
}

// TestResolveReportsUnboundChords guards against a keymap that answers for
// everything, which would make every shortcut look configured.
func TestResolveReportsUnboundChords(t *testing.T) {
	built, err := keymap.FromBindings(map[string]string{"ctrl+s": "save"})
	if err != nil {
		t.Fatalf("construindo keymap: %v", err)
	}
	unbound, err := keymap.Parse("ctrl+alt+shift+f9")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if _, ok := built.Resolve(unbound); ok {
		t.Error("chord não ligado resolveu para alguma ação")
	}
}

// TestNilKeymapResolvesNothing makes the zero value safe to ask: an unset keymap
// must not appear to have bindings.
func TestNilKeymapResolvesNothing(t *testing.T) {
	var empty *keymap.Map
	if _, ok := empty.Resolve(keymap.MustParse("ctrl+s")); ok {
		t.Error("keymap nil resolveu um chord")
	}
	if empty.Len() != 0 {
		t.Error("keymap nil tem bindings")
	}
}
