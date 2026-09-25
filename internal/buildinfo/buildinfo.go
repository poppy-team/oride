// Package buildinfo carries the identity of the running binary.
//
// It exists as its own package so the release pipeline has a stable ldflags
// target: `-X github.com/ori-team/oride/internal/buildinfo.Version=…` keeps
// working when `main` moves, and `main` is not an importable path.
package buildinfo

// Version is the product version. Overridden at build time by the release
// pipeline; the literal below is what a `go build` from a working tree reports.
var Version = "0.2.0-dev"

// Commit is the source revision the binary was built from, when known.
var Commit = ""

// String renders the version the way `oride --version` prints it.
func String() string {
	if Commit == "" {
		return Version
	}
	return Version + " (" + Commit + ")"
}
