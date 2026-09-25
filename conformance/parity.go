package conformance

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
)

// Status is the outcome of comparing one case.
type Status int

const (
	// StatusPass means the two reports are identical.
	StatusPass Status = iota
	// StatusFail means they differ somewhere the ledger does not excuse.
	StatusFail
	// StatusSkipped means the case did not run.
	StatusSkipped
	// StatusDiverged means they differ, and the case declared it — the ledger
	// says the oracle is wrong here.
	StatusDiverged
)

// String renders the status the way the parity report prints it.
func (s Status) String() string {
	switch s {
	case StatusPass:
		return "pass"
	case StatusFail:
		return "FAIL"
	case StatusSkipped:
		return "skip"
	case StatusDiverged:
		return "diverge"
	default:
		return "?"
	}
}

// Result is the outcome for one case.
type Result struct {
	Case   string
	Status Status
	// Reason states which step diverged, or why the case was skipped.
	Reason string
	// Diff is the readable difference. Empty when the case passes.
	Diff string
	// Skipped names the state keys this comparison excluded, and why.
	//
	// Carried on the result rather than kept out of band so the parity report
	// cannot claim coverage it does not have.
	Skipped []string
	// Ledger names the item that justifies an expected divergence, if any.
	Ledger string
}

// CompareMeasured compares two reports over the state this migration produces.
//
// Keys listed in NotYetImplemented are removed from both sides before comparing,
// so a slice that has not landed yet does not fail the suite — while the
// exclusion stays visible, with a reason per key, on the result. Comparing the
// whole state would fail on every unimplemented field, and a suite that is
// expected to fail is a suite nobody reads.
func CompareMeasured(name string, want, got *Report) (Result, error) {
	result, err := Compare(name, stripUnimplemented(want), stripUnimplemented(got))
	if err != nil {
		return Result{}, err
	}
	result.Skipped = ExcludedKeys()
	return result, nil
}

// stripUnimplemented copies a report without the excluded state keys.
func stripUnimplemented(report *Report) *Report {
	out := *report
	out.Frames = make([]Frame, len(report.Frames))
	for i, frame := range report.Frames {
		out.Frames[i] = frame
		state, ok := frame.State.(map[string]any)
		if !ok {
			continue
		}
		trimmed := make(map[string]any, len(state))
		for key, value := range state {
			if _, skip := NotYetImplemented[key]; skip {
				continue
			}
			trimmed[key] = value
		}
		out.Frames[i].State = trimmed
	}
	return &out
}

// Compare compares two reports for the same case, without tolerance.
//
// It reports the *first* divergent frame rather than the whole report: one case
// is one scenario, so the first wrong step is the cause and every later frame is
// cascade. A diff of the entire report would bury the cause in the noise it
// produced.
//
// The error return is for a comparison that cannot be made at all — mismatched
// schema or step count. That is a harness fault, not a product divergence, and
// collapsing the two into one status would hide a broken suite behind a number
// that looks like progress.
func Compare(name string, want, got *Report) (Result, error) {
	if want.Schema != got.Schema {
		return Result{}, fmt.Errorf(
			"%s: schema %d contra %d — oráculo e implementação falam línguas diferentes",
			name, want.Schema, got.Schema)
	}
	if len(want.Frames) != len(got.Frames) {
		return Result{}, fmt.Errorf(
			"%s: %d quadros contra %d — os dois lados não executaram os mesmos passos",
			name, len(want.Frames), len(got.Frames))
	}

	for i := range want.Frames {
		expected, actual := want.Frames[i], got.Frames[i]
		if diff := cmp.Diff(expected, actual); diff != "" {
			return Result{
				Case:   name,
				Status: StatusFail,
				Reason: fmt.Sprintf("passo %d (%s)", expected.AfterStep, expected.Step),
				Diff:   diff,
			}, nil
		}
	}

	return Result{Case: name, Status: StatusPass}, nil
}

// ParityReport aggregates the results of a run.
type ParityReport struct {
	Results []Result
}

// Passed counts the cases whose reports were identical.
func (p ParityReport) Passed() int { return p.count(StatusPass) }

// Failed counts the cases that diverged without declaring it.
func (p ParityReport) Failed() int { return p.count(StatusFail) }

// Skipped counts the cases that did not run.
func (p ParityReport) Skipped() int { return p.count(StatusSkipped) }

// Diverged counts the cases whose divergence the ledger excuses.
func (p ParityReport) Diverged() int { return p.count(StatusDiverged) }

// Total counts every case considered.
func (p ParityReport) Total() int { return len(p.Results) }

// OK reports whether nothing diverged unexpectedly.
func (p ParityReport) OK() bool { return p.Failed() == 0 }

// Parity returns the share of comparable cases that matched, and whether
// anything was comparable at all.
//
// The second return value exists because a suite where everything was skipped
// has no parity to report. Answering 0% there would read as total failure and
// answering 100% would read as success; both would be lies.
func (p ParityReport) Parity() (float64, bool) {
	comparable := p.Passed() + p.Failed()
	if comparable == 0 {
		return 0, false
	}
	return float64(p.Passed()) / float64(comparable), true
}

func (p ParityReport) count(status Status) int {
	total := 0
	for _, result := range p.Results {
		if result.Status == status {
			total++
		}
	}
	return total
}

// String renders the report the way it is written to
// docs/migration/parity-report.md.
func (p ParityReport) String() string {
	var out strings.Builder

	parity, comparable := p.Parity()
	switch {
	case !comparable:
		fmt.Fprintf(&out, "Paridade: sem caso comparável (%d de %d pulados)\n",
			p.Skipped(), p.Total())
	default:
		// Counts first, percentage second. A bare "100%" next to five declared
		// divergences reads as "everything is perfect", which is the opposite of
		// what it says: the percentage only covers the cases where the two sides
		// are expected to agree.
		fmt.Fprintf(&out, "Paridade: %d idêntico(s) · %d falha(s) · %d divergência(s) declarada(s) — de %d casos\n",
			p.Passed(), p.Failed(), p.Diverged(), p.Total())
		fmt.Fprintf(&out, "  (%.1f%% dos %d casos em que os dois lados devem concordar)\n",
			parity*100, p.Passed()+p.Failed())
		if skipped := p.Skipped(); skipped > 0 {
			fmt.Fprintf(&out, "  (%d pulados)\n", skipped)
		}
	}

	for _, result := range p.Results {
		if result.Status == StatusPass {
			continue
		}
		label := result.Status.String()
		if result.Ledger != "" {
			label = fmt.Sprintf("%s (%s)", label, result.Ledger)
		}
		fmt.Fprintf(&out, "\n%s %s — %s\n", label, result.Case, result.Reason)
		if result.Diff != "" {
			out.WriteString(indent(result.Diff, "    "))
			out.WriteString("\n")
		}
	}

	return out.String()
}

func indent(text, prefix string) string {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}
