package conformance

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

// anyNumber builds the value the decoder produces for a JSON number, so a
// hand-built report compares against a decoded one.
func anyNumber(raw string) json.Number { return json.Number(raw) }

// TestGoMatchesTheOracle is the differential assertion this migration exists for.
//
// The same case is replayed against both implementations and their observable
// state is compared. Until this test existed the harness only compared the oracle
// against itself: it could prove the oracle was stable, never that the Go port
// agreed with it.
//
// State the Go side does not produce yet is excluded, and the exclusions carry a
// reason each (see NotYetImplemented). The count of compared cases is asserted,
// so a run where nothing was comparable fails instead of passing quietly.
func TestGoMatchesTheOracle(t *testing.T) {
	oracle := oracleForTest(t)
	paths, err := LoadCases(casesDir(t))
	if err != nil {
		t.Fatalf("descobrindo casos: %v", err)
	}

	compared := 0
	report := ParityReport{}

	for _, path := range paths {
		loaded, err := LoadCase(path)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}

		name := filepath.Base(path)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

		want, err := oracle.Run(ctx, path, t.TempDir())
		if err != nil {
			cancel()
			t.Fatalf("oráculo em %s: %v", name, err)
		}
		got, err := GoRun(ctx, loaded, t.TempDir())
		cancel()
		if err != nil {
			t.Fatalf("implementação Go em %s: %v", name, err)
		}

		result, err := CompareMeasured(name, want, got)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		if divergence := loaded.Divergence; divergence != nil {
			// The case says it must differ, because the ledger records the
			// oracle as wrong here. Agreement would mean the item can be closed,
			// so it is reported rather than ignored.
			if result.Status == StatusPass {
				t.Errorf("%s: divergência declarada (%s) mas o Go passou a concordar com o oráculo — o item do ledger pode ser fechado",
					name, divergence.Ledger)
			}
			result.Status = StatusDiverged
			result.Ledger = divergence.Ledger
			result.Reason = divergence.Note
			// The diff of an expected divergence is noise in the report; the
			// ledger carries the explanation.
			result.Diff = ""
		}

		report.Results = append(report.Results, result)
		compared++
	}

	if compared == 0 {
		t.Fatal("nenhum caso comparado: o teste passaria sem medir nada")
	}

	// Always reported, not only on failure: the number is the product, and a
	// number nobody sees is a number nobody trusts.
	t.Log("\n" + report.String())
	if len(NotYetImplemented) > 0 {
		t.Logf("estado ainda não produzido pelo Go: %d chave(s) — %v",
			len(NotYetImplemented), ExcludedKeys())
	}

	if !report.OK() {
		t.Fatalf("paridade medida:\n\n%s", report.String())
	}
}

// TestGoRunRejectsPathsOutsideTheWorkspace guards the sandbox. A case is data,
// and data does not get to write outside the directory the harness handed over.
func TestGoRunRejectsPathsOutsideTheWorkspace(t *testing.T) {
	c := &Case{
		Files: []CaseFile{{Path: "../escape.txt", Content: "nope"}},
		Steps: []Step{{Kind: StepAction, Action: "move_doc_end"}},
	}
	if err := c.Validate(); err != nil {
		// Validate may reject it first, which is also fine.
		return
	}
	if _, err := GoRun(context.Background(), c, t.TempDir()); err == nil {
		t.Fatal("caminho que escapa do workspace foi aceito")
	}
}

// TestGoRunRejectsUnknownActions: an action the Go side does not implement must
// fail loudly. Silently ignoring it would make a case report parity it never
// measured.
func TestGoRunRejectsUnknownActions(t *testing.T) {
	c := &Case{
		Files: []CaseFile{{Path: "a.txt", Content: "x"}},
		Open:  "a.txt",
		Steps: []Step{{Kind: StepAction, Action: "toggle_terminal"}},
	}
	if _, err := GoRun(context.Background(), c, t.TempDir()); err == nil {
		t.Fatal("ação não implementada foi aceita em silêncio")
	}
}

// TestExclusionsCarryAReason keeps the exclusion list honest: every key skipped
// must say which slice will produce it.
func TestExclusionsCarryAReason(t *testing.T) {
	for key, reason := range NotYetImplemented {
		if reason == "" {
			t.Errorf("chave %q está excluída sem motivo", key)
		}
	}
	if len(NotYetImplemented) == 0 {
		t.Log("nenhuma exclusão: a comparação cobre o estado inteiro")
	}
}

// TestMeasuredComparisonStillDetectsDivergence makes sure the exclusion
// machinery did not quietly neuter the comparator.
func TestMeasuredComparisonStillDetectsDivergence(t *testing.T) {
	want := &Report{
		Schema: 1,
		Steps:  1,
		Frames: []Frame{{
			AfterStep: 1,
			Step:      "action move_doc_end",
			State: map[string]any{
				"dirty_count": anyNumber("0"),
			},
		}},
	}
	got := &Report{
		Schema: 1,
		Steps:  1,
		Frames: []Frame{{
			AfterStep: 1,
			Step:      "action move_doc_end",
			State: map[string]any{
				"dirty_count": anyNumber("1"),
			},
		}},
	}

	result, err := CompareMeasured("mutado", want, got)
	if err != nil {
		t.Fatalf("comparação recusada: %v", err)
	}
	if result.Status != StatusFail {
		t.Fatal("uma divergência em campo medido foi comparada como igual")
	}
	if len(result.Skipped) == 0 {
		t.Error("o resultado não registra o que foi excluído")
	}
}

// TestExcludedFieldsDoNotFailTheComparison: a divergence only in an excluded key
// must not fail, or the suite would be red until every slice lands.
//
// The key is taken from the exclusion list rather than hardcoded, so that
// including a key later breaks this test loudly instead of leaving it asserting
// something that is no longer true.
func TestExcludedFieldsDoNotFailTheComparison(t *testing.T) {
	keys := ExcludedKeys()
	if len(keys) == 0 {
		t.Skip("nenhuma exclusão restante: nada a comparar aqui")
	}
	excluded := keys[0]

	makeReport := func(value string) *Report {
		return &Report{
			Schema: 1,
			Steps:  1,
			Frames: []Frame{{
				AfterStep: 1,
				Step:      "action move_doc_end",
				State: map[string]any{
					"dirty_count": anyNumber("0"),
					excluded:      map[string]any{"rows": anyNumber(value)},
				},
			}},
		}
	}

	result, err := CompareMeasured("chave excluída divergente", makeReport("1"), makeReport("99"))
	if err != nil {
		t.Fatalf("comparação recusada: %v", err)
	}
	if result.Status != StatusPass {
		t.Fatalf("a chave excluída %q derrubou a comparação: %s", excluded, result.Diff)
	}
}
