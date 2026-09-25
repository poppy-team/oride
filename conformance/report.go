package conformance

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// DumpSchemaVersion is the shape version of the state dump.
//
// It is duplicated here on purpose: the Rust side owns the constant, and this
// copy is what makes a fixture written for one shape fail loudly against the
// other instead of comparing fields that no longer exist.
const DumpSchemaVersion = 1

// Report is one case's outcome: the observable state after every step.
type Report struct {
	Schema      int     `json:"schema"`
	Description string  `json:"description,omitempty"`
	Steps       int     `json:"steps"`
	Frames      []Frame `json:"frames"`
}

// Frame is the state captured immediately after one step.
type Frame struct {
	AfterStep int    `json:"after_step"`
	Step      string `json:"step"`
	// State is the state dump, left untyped on purpose.
	//
	// Typing every field in Go would mean a field the oracle adds is silently
	// dropped here — and a comparison that ignores new information is worse than
	// no comparison, because it reports parity while the shapes drift apart.
	State any `json:"state"`
}

// DecodeReport parses a report from raw JSON and checks its shape.
func DecodeReport(raw []byte) (*Report, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	// Numbers stay exact: a byte count that round-trips through float64 would
	// still compare equal, but its diff would print something nobody wrote.
	decoder.UseNumber()

	var report Report
	if err := decoder.Decode(&report); err != nil {
		return nil, fmt.Errorf("relatório: %w", err)
	}
	if err := report.Validate(); err != nil {
		return nil, err
	}
	return &report, nil
}

// Validate rejects a report the harness could not compare faithfully.
//
// These are contract checks, not sanity checks: if a frame is missing or out of
// order, the comparison would pass over a gap and report parity that was never
// measured.
func (r Report) Validate() error {
	if r.Schema != DumpSchemaVersion {
		return fmt.Errorf("dump no schema %d, esperado %d: fixture e implementação estão em versões diferentes",
			r.Schema, DumpSchemaVersion)
	}
	if len(r.Frames) != r.Steps {
		return fmt.Errorf("%d quadros para %d passos: o relatório perdeu passos",
			len(r.Frames), r.Steps)
	}
	for i, frame := range r.Frames {
		if frame.AfterStep != i+1 {
			return fmt.Errorf("quadro %d diz after_step=%d, esperado %d: quadros fora de ordem",
				i, frame.AfterStep, i+1)
		}
		if frame.State == nil {
			return fmt.Errorf("quadro %d (após passo %d) sem estado", i, frame.AfterStep)
		}
	}
	return nil
}
