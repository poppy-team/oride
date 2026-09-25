// Package conformance is the differential parity harness between the Rust
// oracle and the Go implementation.
//
// The oracle is the frozen Rust binary. It replays a scripted case and prints
// the observable state after every step as canonical JSON; this package runs
// the same case against the Go implementation and compares the two reports
// without tolerance. Parity is therefore a number a program produces, not a
// claim a person makes.
//
// The case format is documented in conformance/README.md, and every deliberate
// divergence is recorded in docs/migration/parity-ledger.md. A divergence that
// is not in the ledger is a regression.
package conformance

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Step kinds understood by the oracle. A case that names anything else is
// rejected rather than silently skipped, because a step that does nothing looks
// exactly like a step that passed.
const (
	StepAction = "action"
	StepChord  = "chord"
	StepText   = "text"
	// StepSearch sets the in-buffer search state and recomputes it.
	StepSearch = "search"
)

// Case is one scripted scenario.
type Case struct {
	// Description states what invariant the case guards. It travels into the
	// report so a failing diff explains itself.
	Description string `json:"description,omitempty"`
	// Files are materialised in the workspace before the run.
	Files []CaseFile `json:"files,omitempty"`
	// Open is the workspace-relative file opened at start. Empty means an
	// empty buffer.
	Open string `json:"open,omitempty"`
	// Config overrides the product defaults.
	Config CaseConfig `json:"config,omitempty"`
	// Git enables git queries. Off by default: a case that does not test git
	// must not depend on the binary being installed.
	Git bool `json:"git,omitempty"`
	// Steps run in order; state is captured after each one.
	Steps []Step `json:"steps"`
	// Divergence declares that this case is expected to differ from the oracle,
	// because the ledger says the oracle is wrong here.
	//
	// A case that diverges without declaring it is a regression. A case that
	// declares a divergence and stops diverging is also worth knowing about: it
	// means the ledger item can be closed.
	Divergence *Divergence `json:"divergence,omitempty"`
}

// Divergence records an expected disagreement with the oracle.
type Divergence struct {
	// Ledger names the item that justifies it, e.g. "B2".
	Ledger string `json:"ledger"`
	// Note says what diverges, in one line.
	Note string `json:"note,omitempty"`
}

// CaseFile is one file the case declares.
type CaseFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// CaseConfig overrides the product defaults for one case.
//
// Every field is a pointer because "absent" and "false" are different answers:
// a case may legitimately turn something off, and the harness must not turn it
// off by accident while marshalling.
type CaseConfig struct {
	Theme           *string           `json:"theme,omitempty"`
	Locale          *string           `json:"locale,omitempty"`
	ShowLineNumbers *bool             `json:"show_line_numbers,omitempty"`
	SoftWrap        *bool             `json:"soft_wrap,omitempty"`
	Mouse           *bool             `json:"mouse,omitempty"`
	ModalMode       *bool             `json:"modal_mode,omitempty"`
	TabSize         *int              `json:"tab_size,omitempty"`
	InsertSpaces    *bool             `json:"insert_spaces,omitempty"`
	TreeWidth       *int              `json:"tree_width,omitempty"`
	ShowTree        *bool             `json:"show_tree,omitempty"`
	ShowScm         *bool             `json:"show_scm,omitempty"`
	ShowHidden      *bool             `json:"show_hidden,omitempty"`
	Keys            map[string]string `json:"keys,omitempty"`
}

// Step is one input in a case.
//
// The struct is flat rather than an interface with a tagged union because the
// JSON is flat: the oracle deserialises `kind` plus one payload field, and
// mirroring that shape keeps the two readers obviously equivalent.
type Step struct {
	Kind string `json:"kind"`
	// Action is the stable action id, e.g. "move_doc_end".
	Action string `json:"action,omitempty"`
	// Chord is a canonical keystroke, e.g. "ctrl+s". It resolves through the
	// effective keymap, so an unbound chord fails the case — that is how keymap
	// drift is caught.
	Chord string `json:"chord,omitempty"`
	// Text is typed character by character.
	Text string `json:"text,omitempty"`

	// The fields below belong to a `search` step. They are flat for the same
	// reason the rest of this struct is: the JSON is flat, and mirroring it keeps
	// the two readers obviously equivalent.
	//
	// Query may be empty — clearing a search is a legitimate step — so it is the
	// only payload that does not have to be non-empty.
	Query         string `json:"query,omitempty"`
	Replace       string `json:"replace,omitempty"`
	CaseSensitive *bool  `json:"case_sensitive,omitempty"`
	IgnoreAccents *bool  `json:"ignore_accents,omitempty"`
	WholeWord     *bool  `json:"whole_word,omitempty"`
	UseRegex      *bool  `json:"use_regex,omitempty"`
}

// LoadCase reads and validates one case file.
func LoadCase(path string) (*Case, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("lendo caso %s: %w", path, err)
	}
	var c Case
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("caso %s: %w", path, err)
	}
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("caso %s: %w", path, err)
	}
	return &c, nil
}

// LoadCases finds every case under root, in sorted order.
//
// Sorted because the parity report is read by people, and an order that depends
// on the filesystem makes two runs of the same suite look different.
func LoadCases(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("varrendo %s: %w", root, err)
	}
	sort.Strings(paths)
	return paths, nil
}

// Validate rejects a case the harness could not run faithfully.
func (c Case) Validate() error {
	if len(c.Steps) == 0 {
		return fmt.Errorf("nenhum passo: um caso sem passos não afirma nada")
	}
	for i, step := range c.Steps {
		if err := step.Validate(); err != nil {
			return fmt.Errorf("passo %d: %w", i+1, err)
		}
	}
	for i, file := range c.Files {
		if file.Path == "" {
			return fmt.Errorf("arquivo %d: path vazio", i+1)
		}
		if filepath.IsAbs(file.Path) {
			return fmt.Errorf("arquivo %d: %q é absoluto; caminhos do caso são relativos ao workspace", i+1, file.Path)
		}
	}
	if filepath.IsAbs(c.Open) {
		return fmt.Errorf("open %q é absoluto; caminhos do caso são relativos ao workspace", c.Open)
	}
	if c.Divergence != nil && c.Divergence.Ledger == "" {
		return fmt.Errorf("divergência declarada sem item do ledger: uma exceção sem justificativa é uma regressão com nome bonito")
	}
	return nil
}

// Validate rejects an ambiguous or unknown step.
func (s Step) Validate() error {
	payloads := map[string]string{
		StepAction: s.Action,
		StepChord:  s.Chord,
		StepText:   s.Text,
	}

	switch s.Kind {
	case StepAction, StepChord, StepText, StepSearch:
	default:
		return fmt.Errorf("kind %q desconhecido; use %s, %s, %s ou %s",
			s.Kind, StepAction, StepChord, StepText, StepSearch)
	}

	// A search with an empty query is meaningful — it clears the search — so the
	// non-empty rule applies to every other kind.
	if s.Kind != StepSearch && payloads[s.Kind] == "" {
		return fmt.Errorf("kind %q sem o campo %q preenchido; um passo vazio não afirma nada",
			s.Kind, s.Kind)
	}
	if s.Kind == StepSearch && (s.Action != "" || s.Chord != "" || s.Text != "") {
		return fmt.Errorf("kind %q também preenche outro payload; um passo carrega um payload só",
			s.Kind)
	}

	// A step carrying two payloads is ambiguous: the oracle ignores the extra
	// one, so the case would read as testing one thing while testing another.
	// Iterated in fixed order so the message does not vary between runs.
	if s.Kind == StepSearch {
		return nil
	}
	for _, kind := range []string{StepAction, StepChord, StepText} {
		if kind == s.Kind {
			continue
		}
		if payloads[kind] != "" {
			return fmt.Errorf("kind %q também preenche %q; um passo carrega um payload só",
				s.Kind, kind)
		}
	}
	return nil
}
