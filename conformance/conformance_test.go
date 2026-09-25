package conformance

import (
	"context"
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// oracleTimeout bounds one oracle invocation. The oracle is a local process
// replaying scripted input, so anything beyond this is a hang, not slowness.
const oracleTimeout = 60 * time.Second

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("diretório de trabalho: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod não encontrado a partir de %s", dir)
		}
		dir = parent
	}
}

func casesDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "conformance", "cases")
}

func oracleForTest(t *testing.T) Oracle {
	t.Helper()

	oracle, err := DiscoverOracle(repoRoot(t))
	if err != nil {
		t.Skipf("sem oráculo: %v", err)
	}
	return oracle
}

func runOracle(t *testing.T, oracle Oracle, casePath string) *Report {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), oracleTimeout)
	defer cancel()

	report, err := oracle.Run(ctx, casePath, t.TempDir())
	if err != nil {
		t.Fatalf("executando %s: %v", filepath.Base(casePath), err)
	}
	return report
}

// TestEveryCaseLoads needs no oracle: a malformed fixture is a harness bug, and
// it should be reported even on a machine that cannot build Rust.
func TestEveryCaseLoads(t *testing.T) {
	paths, err := LoadCases(casesDir(t))
	if err != nil {
		t.Fatalf("descobrindo casos: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("nenhum caso encontrado: o gate de paridade não estaria medindo nada")
	}

	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			if _, err := LoadCase(path); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestOracleProtocolContract checks the shape the harness depends on. A frame
// silently missing here would make every later comparison pass over a gap.
func TestOracleProtocolContract(t *testing.T) {
	oracle := oracleForTest(t)
	paths, err := LoadCases(casesDir(t))
	if err != nil {
		t.Fatalf("descobrindo casos: %v", err)
	}

	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			loadCase, err := LoadCase(path)
			if err != nil {
				t.Fatal(err)
			}

			report := runOracle(t, oracle, path)

			if report.Schema != DumpSchemaVersion {
				t.Errorf("schema %d, esperado %d", report.Schema, DumpSchemaVersion)
			}
			if report.Steps != len(loadCase.Steps) {
				t.Errorf("relatório diz %d passos, o caso tem %d", report.Steps, len(loadCase.Steps))
			}
			if report.Description != loadCase.Description {
				t.Errorf("descrição do relatório difere da do caso:\n caso: %q\n relatório: %q",
					loadCase.Description, report.Description)
			}
		})
	}
}

// TestOracleIsDeterministic is the precondition for the whole method.
//
// Differential parity only means something if the reference gives the same
// answer twice. Running the same case in two different directories also proves
// the dump carries no trace of where it ran — which is what lets a fixture live
// in git at all.
func TestOracleIsDeterministic(t *testing.T) {
	oracle := oracleForTest(t)
	paths, err := LoadCases(casesDir(t))
	if err != nil {
		t.Fatalf("descobrindo casos: %v", err)
	}

	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			first := runOracle(t, oracle, path)
			second := runOracle(t, oracle, path)

			result, err := Compare(filepath.Base(path), first, second)
			if err != nil {
				t.Fatalf("comparando duas execuções do oráculo: %v", err)
			}
			if result.Status != StatusPass {
				t.Fatalf("o oráculo não é determinístico — %s:\n%s\n\n"+
					"paridade diferencial não tem sentido sem isso",
					result.Reason, result.Diff)
			}
		})
	}
}

// TestComparisonDetectsDivergence guards against a vacuous comparator.
//
// A comparison that always answers "equal" would make the parity number climb
// to 100% on its own, which is the one failure mode of this harness that would
// look like success. So the comparator is tested with a divergence it must find.
func TestComparisonDetectsDivergence(t *testing.T) {
	oracle := oracleForTest(t)
	paths, err := LoadCases(casesDir(t))
	if err != nil || len(paths) == 0 {
		t.Fatalf("descobrindo casos: %v (%d casos)", err, len(paths))
	}
	report := runOracle(t, oracle, paths[0])

	t.Run("estado divergente", func(t *testing.T) {
		mutated := cloneReport(report)
		last := len(mutated.Frames) - 1
		state, ok := mutated.Frames[last].State.(map[string]any)
		if !ok {
			t.Fatalf("estado do quadro %d não é objeto: %T", last, mutated.Frames[last].State)
		}
		changed := maps.Clone(state)
		changed["dirty_count"] = json.Number("99")
		mutated.Frames[last].State = changed

		result, err := Compare("mutado", report, mutated)
		if err != nil {
			t.Fatalf("comparação recusada: %v", err)
		}
		if result.Status != StatusFail {
			t.Fatal("um estado divergente foi comparado como igual")
		}
		if result.Diff == "" {
			t.Error("falha sem diff: o relatório não diria o que mudou")
		}
		if !strings.Contains(result.Reason, "passo") {
			t.Errorf("razão %q não aponta o passo divergente", result.Reason)
		}
	})

	t.Run("passo divergente", func(t *testing.T) {
		mutated := cloneReport(report)
		mutated.Frames[0].Step = "chord ctrl+inexistente"

		result, err := Compare("mutado", report, mutated)
		if err != nil {
			t.Fatalf("comparação recusada: %v", err)
		}
		if result.Status != StatusFail {
			t.Fatal("um passo divergente foi comparado como igual")
		}
	})
}

// TestComparisonRejectsMismatchedShapes checks that a harness fault is reported
// as an error rather than as a product divergence. Collapsing the two would let
// a broken suite masquerade as progress.
func TestComparisonRejectsMismatchedShapes(t *testing.T) {
	t.Run("schema diferente", func(t *testing.T) {
		want := &Report{Schema: 1, Steps: 1, Frames: []Frame{{AfterStep: 1, Step: "x", State: map[string]any{}}}}
		got := &Report{Schema: 2, Steps: 1, Frames: []Frame{{AfterStep: 1, Step: "x", State: map[string]any{}}}}

		if _, err := Compare("c", want, got); err == nil {
			t.Fatal("schemas diferentes compararam sem erro")
		}
	})

	t.Run("contagem de quadros diferente", func(t *testing.T) {
		want := &Report{Schema: 1, Steps: 2, Frames: []Frame{
			{AfterStep: 1, Step: "a", State: map[string]any{}},
			{AfterStep: 2, Step: "b", State: map[string]any{}},
		}}
		got := &Report{Schema: 1, Steps: 1, Frames: []Frame{
			{AfterStep: 1, Step: "a", State: map[string]any{}},
		}}

		if _, err := Compare("c", want, got); err == nil {
			t.Fatal("contagens de quadros diferentes compararam sem erro")
		}
	})
}

func TestCaseRejectsUnknownStepKind(t *testing.T) {
	_, err := LoadCase(writeTempCase(t, `{"steps":[{"kind":"teleport","action":"save"}]}`))
	if err == nil {
		t.Fatal("kind desconhecido foi aceito")
	}
	if !strings.Contains(err.Error(), "teleport") {
		t.Errorf("erro %q não nomeia o kind inválido", err)
	}
}

func TestCaseRejectsAmbiguousStep(t *testing.T) {
	_, err := LoadCase(writeTempCase(t, `{"steps":[{"kind":"action","action":"save","chord":"ctrl+s"}]}`))
	if err == nil {
		t.Fatal("passo com dois payloads foi aceito")
	}
}

func TestCaseRejectsStepsThatAssertNothing(t *testing.T) {
	if _, err := LoadCase(writeTempCase(t, `{"steps":[]}`)); err == nil {
		t.Fatal("caso sem passos foi aceito")
	}
	if _, err := LoadCase(writeTempCase(t, `{"steps":[{"kind":"text"}]}`)); err == nil {
		t.Fatal("passo de texto vazio foi aceito")
	}
}

func TestReportRejectsInconsistentFrames(t *testing.T) {
	raw := []byte(`{"schema":1,"steps":3,"frames":[{"after_step":1,"step":"a","state":{}}]}`)
	if _, err := DecodeReport(raw); err == nil {
		t.Fatal("relatório com quadros faltando foi aceito")
	}

	outOfOrder := []byte(`{"schema":1,"steps":2,"frames":[
		{"after_step":2,"step":"a","state":{}},
		{"after_step":1,"step":"b","state":{}}]}`)
	if _, err := DecodeReport(outOfOrder); err == nil {
		t.Fatal("relatório com quadros fora de ordem foi aceito")
	}
}

func TestParityReportRefusesToInventANumber(t *testing.T) {
	empty := ParityReport{Results: []Result{{Case: "a", Status: StatusSkipped}}}
	if _, comparable := empty.Parity(); comparable {
		t.Fatal("uma suíte inteiramente pulada não tem paridade a reportar")
	}
	if !strings.Contains(empty.String(), "sem caso comparável") {
		t.Errorf("resumo não admite a ausência de medição:\n%s", empty.String())
	}

	mixed := ParityReport{Results: []Result{
		{Case: "a", Status: StatusPass},
		{Case: "b", Status: StatusFail, Reason: "passo 1", Diff: "-1\n+2"},
		{Case: "c", Status: StatusSkipped},
	}}
	parity, comparable := mixed.Parity()
	if !comparable || parity != 0.5 {
		t.Fatalf("paridade = %v (comparable=%v), esperado 0.5", parity, comparable)
	}
	if mixed.OK() {
		t.Error("OK() verdadeiro com uma falha presente")
	}
}

// cloneReport deep-copies enough of a report to mutate it in isolation.
func cloneReport(report *Report) *Report {
	clone := *report
	clone.Frames = make([]Frame, len(report.Frames))
	for i, frame := range report.Frames {
		clone.Frames[i] = frame
		if state, ok := frame.State.(map[string]any); ok {
			clone.Frames[i].State = maps.Clone(state)
		}
	}
	return &clone
}

func writeTempCase(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "case.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("escrevendo caso: %v", err)
	}
	return path
}
