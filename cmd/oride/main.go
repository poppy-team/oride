// Command oride is the Oride terminal IDE.
//
// This is the Go implementation. It is a composition root and nothing else:
// argument parsing, then delegation. Product decisions belong to the packages
// under internal/, which are usable without a terminal.
//
// The Rust implementation under crates/ is still the behavioural reference — it
// is the differential oracle for `go test ./conformance/...`. Until the Go
// editor reaches parity, the TUI path here reports that it is not wired yet
// instead of pretending to work.
package main

import (
	"fmt"
	"os"

	"github.com/ori-team/oride/internal/buildinfo"
)

const usage = `oride — terminal IDE

USAGE:
  oride --version      print the version
  oride --help         print this message

The Go editor is under migration from the Rust implementation. The TUI is not
wired yet; parity is tracked in docs/migration/parity-ledger.md and enforced by
the differential harness in conformance/.
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	for _, arg := range args {
		switch arg {
		case "--version", "-V", "version":
			fmt.Printf("oride %s\n", buildinfo.String())
			return nil
		case "--help", "-h", "help":
			fmt.Print(usage)
			return nil
		}
	}

	if len(args) > 0 {
		fmt.Print(usage)
		return fmt.Errorf("the Go editor is not wired yet, so %q has nothing to open; "+
			"the Rust implementation in crates/ still runs the TUI", args[0])
	}

	fmt.Print(usage)
	return nil
}
