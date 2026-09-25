package conformance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ori-team/oride/internal/action"
	"github.com/ori-team/oride/internal/app"
	"github.com/ori-team/oride/internal/config"
	"github.com/ori-team/oride/internal/editor"
	"github.com/ori-team/oride/internal/fs"
	"github.com/ori-team/oride/internal/keymap"
	"github.com/ori-team/oride/internal/search"
)

// NotYetImplemented names the state the Go side does not produce yet, with the
// slice that will produce it.
//
// The comparison skips exactly these keys and nothing else. An exclusion without
// a reason is how a parity suite starts lying: the number keeps climbing while
// the coverage quietly shrinks.
var NotYetImplemented = map[string]string{
	"scm":          "M2 — nenhum caso estabelece um repositório git ainda",
	"split":        "M3 — painéis",
	"vim":          "M3 — motor modal",
	"diagnostics":  "M2 — diagnósticos LSP",
	"lsp_failures": "M2 — LSP",
	"status":       "M3 — texto da barra de status",
}

// ExcludedKeys returns the sorted keys the comparison skips.
func ExcludedKeys() []string {
	out := make([]string, 0, len(NotYetImplemented))
	for key := range NotYetImplemented {
		out = append(out, key)
	}
	return out
}

// GoRun executes a case against the Go implementation.
//
// It mirrors the oracle's contract: the caller hands over a directory the case
// materialises, and the report has one frame per step. The state is serialised to
// JSON and decoded back so both sides go through the same representation — a
// comparison between an in-memory Go value and a JSON tree would compare types,
// not content.
func GoRun(_ context.Context, c *Case, workspace string) (*Report, error) {
	if err := materializeForGo(c, workspace); err != nil {
		return nil, err
	}

	application, err := buildApp(c, workspace)
	if err != nil {
		return nil, err
	}

	report := &Report{
		Schema:      int(app.SchemaVersion),
		Description: c.Description,
		Steps:       len(c.Steps),
		Frames:      make([]Frame, 0, len(c.Steps)),
	}

	for index, step := range c.Steps {
		if err := applyGoStep(application, step); err != nil {
			return nil, fmt.Errorf("passo %d (%s): %w", index+1, describeStep(step), err)
		}
		raw, err := application.DumpState()
		if err != nil {
			return nil, fmt.Errorf("passo %d: dump: %w", index+1, err)
		}
		var state any
		if err := decodeInto(raw, &state); err != nil {
			return nil, fmt.Errorf("passo %d: %w", index+1, err)
		}
		report.Frames = append(report.Frames, Frame{
			AfterStep: index + 1,
			Step:      describeStep(step),
			State:     state,
		})
	}

	return report, nil
}

func describeStep(step Step) string {
	switch step.Kind {
	case StepChord:
		return "chord " + step.Chord
	case StepText:
		return fmt.Sprintf("text %q", step.Text)
	case StepAction:
		return "action " + step.Action
	case StepSearch:
		// Mirrors the oracle's format exactly. The step description travels in
		// the frame, so a difference here fails the comparison as if the state
		// had diverged.
		return fmt.Sprintf("search %q case=%t word=%t re=%t",
			step.Query,
			boolOr(step.CaseSensitive, false),
			boolOr(step.WholeWord, false),
			boolOr(step.UseRegex, false))
	default:
		return step.Kind
	}
}

// materializeForGo writes the case's files, refusing paths that escape.
func materializeForGo(c *Case, workspace string) error {
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		return fmt.Errorf("criando workspace: %w", err)
	}
	for _, file := range c.Files {
		target, err := resolveInside(workspace, file.Path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("criando %s: %w", filepath.Dir(target), err)
		}
		if err := os.WriteFile(target, []byte(file.Content), 0o644); err != nil {
			return fmt.Errorf("escrevendo %s: %w", target, err)
		}
	}
	return nil
}

// resolveInside joins a case-relative path onto the workspace, rejecting an
// escape. A case is data, and data does not get to write outside its sandbox.
func resolveInside(workspace, relative string) (string, error) {
	base := filepath.Clean(workspace)
	clean := filepath.Clean(filepath.Join(base, relative))
	if clean != base && !strings.HasPrefix(clean, base+string(os.PathSeparator)) {
		return "", fmt.Errorf("caminho fora do workspace: %q", relative)
	}
	return clean, nil
}

// buildApp constructs the headless application from a case.
func buildApp(c *Case, workspace string) (*app.App, error) {
	cfg := config.Default()
	applyCaseConfig(&cfg, c.Config)

	// A case that rebinds a chord must go through the same path a user's config
	// does, or the case would test a keymap the product never builds.
	keys, err := keymap.FromBindings(cfg.Keys)
	if err != nil {
		return nil, fmt.Errorf("keymap do caso: %w", err)
	}

	// Canonicalised for the same reason the oracle canonicalises: the reported
	// `<workspace>` prefix must come from the same form of the path on both
	// sides, and on macOS `/tmp` is a symlink to `/private/tmp`.
	base, err := filepath.EvalSymlinks(workspace)
	if err != nil {
		return nil, fmt.Errorf("canonicalizando workspace: %w", err)
	}

	store := editor.NewStore()
	if c.Open != "" {
		target, err := resolveInside(base, c.Open)
		if err != nil {
			return nil, err
		}
		if _, err := store.OpenPath(target); err != nil {
			return nil, err
		}
	} else {
		store.OpenEmpty()
	}

	application := app.New(store, cfg, keys)
	application.Workspace = base
	// The oracle opens the tree whenever the panel is visible, so a case that
	// leaves it visible is comparing tree contents and not just a flag.
	if application.ShowTree {
		if tree, err := fs.Open(base, cfg.Tree.ShowHidden); err == nil {
			application.Tree = tree
		}
	}
	application.SCM = c.Config.ShowScm != nil && *c.Config.ShowScm
	if c.Config.ShowTree != nil {
		application.ShowTree = *c.Config.ShowTree
	}
	return application, nil
}

// applyCaseConfig applies the case's overrides over the defaults.
func applyCaseConfig(cfg *config.Config, overrides CaseConfig) {
	if overrides.Theme != nil {
		cfg.Theme = *overrides.Theme
	}
	if overrides.Locale != nil {
		cfg.Locale = *overrides.Locale
	}
	if overrides.ShowLineNumbers != nil {
		cfg.ShowLineNumbers = *overrides.ShowLineNumbers
	}
	if overrides.SoftWrap != nil {
		cfg.SoftWrap = *overrides.SoftWrap
	}
	if overrides.Mouse != nil {
		cfg.Mouse = *overrides.Mouse
	}
	if overrides.ModalMode != nil {
		cfg.Editor.ModalMode = *overrides.ModalMode
	}
	if overrides.TabSize != nil {
		cfg.Editor.TabSize = *overrides.TabSize
	}
	if overrides.InsertSpaces != nil {
		cfg.Editor.InsertSpaces = *overrides.InsertSpaces
	}
	if overrides.TreeWidth != nil {
		cfg.Tree.Width = *overrides.TreeWidth
	}
	if overrides.ShowHidden != nil {
		cfg.Tree.ShowHidden = *overrides.ShowHidden
	}
	if cfg.Keys == nil {
		cfg.Keys = map[string]string{}
	}
	for chord, actionID := range overrides.Keys {
		cfg.Keys[chord] = actionID
	}
}

func applyGoStep(application *app.App, step Step) error {
	switch step.Kind {
	case StepChord:
		return application.ApplyKey(step.Chord)
	case StepText:
		return application.ApplyText(step.Text)
	case StepAction:
		bound, err := action.Parse(step.Action)
		if err != nil {
			return err
		}
		return application.Apply(bound)
	case StepSearch:
		options := search.Options{
			CaseSensitive: boolOr(step.CaseSensitive, false),
			// serde's default for an absent bool is false, so the oracle clears
			// the flag when a case omits it. Matching that is matching the
			// contract.
			IgnoreAccents: boolOr(step.IgnoreAccents, false),
			WholeWord:     boolOr(step.WholeWord, false),
			UseRegex:      boolOr(step.UseRegex, false),
		}
		return application.ApplySearch(step.Query, options, step.Replace)
	default:
		return fmt.Errorf("kind desconhecido: %q", step.Kind)
	}
}

// boolOr reads an optional flag, defaulting to what the product does when the
// field is absent.
func boolOr(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

// decodeInto decodes JSON preserving number precision, matching how the oracle's
// report is decoded so the two compare as the same kind of value.
func decodeInto(raw []byte, target *any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	return decoder.Decode(target)
}
