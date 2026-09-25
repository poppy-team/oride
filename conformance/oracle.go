package conformance

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// EnvOracle names the environment variable that overrides oracle discovery.
const EnvOracle = "ORIDE_ORACLE"

// ErrOracleNotFound is returned when no Rust binary is available.
var ErrOracleNotFound = errors.New("oráculo Rust não encontrado")

// Oracle drives the frozen Rust implementation.
//
// It is a subprocess rather than a library because that is the only way to run
// two implementations of the same product in one test: the Rust side needs a
// scripted-input mode precisely because it cannot be linked into a Go test.
type Oracle struct {
	// Binary is the path to the Rust `oride`.
	Binary string
}

// DiscoverOracle locates the Rust binary.
//
// Order: the ORIDE_ORACLE override, then a debug build, then a release build.
// Debug comes first because that is what a working tree has; a developer who
// wants the release binary sets the variable.
func DiscoverOracle(repoRoot string) (Oracle, error) {
	if override := os.Getenv(EnvOracle); override != "" {
		if !isExecutable(override) {
			return Oracle{}, fmt.Errorf("%s aponta para %q, que não é executável", EnvOracle, override)
		}
		return Oracle{Binary: override}, nil
	}

	for _, candidate := range []string{
		filepath.Join("target", "debug", binaryName()),
		filepath.Join("target", "release", binaryName()),
	} {
		path := filepath.Join(repoRoot, candidate)
		if isExecutable(path) {
			return Oracle{Binary: path}, nil
		}
	}

	return Oracle{}, fmt.Errorf("%w em %s: rode `cargo build -p oride` ou defina %s",
		ErrOracleNotFound, repoRoot, EnvOracle)
}

// Run replays one case and returns the oracle's report.
//
// The caller owns workspace and must hand over an empty directory: the oracle
// materialises the case's files but never clears what is already there, so a
// reused directory would let leftovers from a previous case leak into this one.
func (o Oracle) Run(ctx context.Context, casePath, workspace string) (*Report, error) {
	if o.Binary == "" {
		return nil, ErrOracleNotFound
	}

	command := exec.CommandContext(ctx, o.Binary,
		"conformance", "run",
		"--case", casePath,
		"--workspace", workspace,
		"--json",
	)

	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		// The oracle reserves stdout for JSON and puts diagnostics on stderr,
		// so the reason a case failed is in stderr and must travel with the error.
		reason := strings.TrimSpace(stderr.String())
		if reason == "" {
			reason = "sem saída em stderr"
		}
		return nil, fmt.Errorf("oráculo falhou em %s: %w — %s", filepath.Base(casePath), err, reason)
	}

	// A binary that exits 0 without printing JSON is almost always the wrong
	// binary — a mistyped ORIDE_ORACLE, a wrapper, a shell builtin. Saying so is
	// worth more than the decoder's "EOF", which points at the symptom.
	if strings.TrimSpace(stdout.String()) == "" {
		return nil, fmt.Errorf(
			"oráculo %s não imprimiu JSON em %s: confira se é mesmo o binário do oride",
			o.Binary, filepath.Base(casePath))
	}

	return DecodeReport(stdout.Bytes())
}

func binaryName() string {
	if runtime.GOOS == "windows" {
		return "oride.exe"
	}
	return "oride"
}

// isExecutable reports whether path is a file the harness can run.
//
// The execute bit only means something on Unix; on Windows the extension is
// what decides, and a directory named `oride` must never be mistaken for one.
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode().Perm()&0o111 != 0
}
