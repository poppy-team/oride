// Package git reports repository status.
//
// It shells out to the `git` binary rather than linking a library. The reference
// implementation made the same choice: `git status --porcelain` is a stable,
// documented contract, while a library binding is a dependency that must track
// git's own changes.
package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Status is the simplified status the tree and the SCM panel show.
type Status string

// Status values.
const (
	Modified  Status = "Modified"
	Added     Status = "Added"
	Deleted   Status = "Deleted"
	Untracked Status = "Untracked"
	Renamed   Status = "Renamed"
	Conflict  Status = "Conflict"
)

// Badge is the single character shown next to a path.
func (s Status) Badge() string {
	switch s {
	case Modified:
		return "M"
	case Added:
		return "A"
	case Deleted:
		return "D"
	case Untracked:
		return "?"
	case Renamed:
		return "R"
	case Conflict:
		return "U"
	default:
		return "?"
	}
}

// rank orders statuses by how much attention they deserve, so that a path with
// two entries keeps the more alarming one.
//
// A file can appear twice — staged and modified, for instance — and showing the
// milder status would hide the reason it needs a look.
func rank(s Status) int {
	switch s {
	case Conflict:
		return 5
	case Deleted:
		return 4
	case Modified:
		return 3
	case Renamed:
		return 2
	case Added:
		return 1
	case Untracked:
		return 0
	default:
		return -1
	}
}

func worse(a, b Status) Status {
	if rank(b) > rank(a) {
		return b
	}
	return a
}

// Entry is one status map entry.
type Entry struct {
	Status Status
	Path   string
}

// StatusMap returns path → status for a repository.
//
// A missing `git`, a directory that is not a repository, or a failing command
// all produce an empty map rather than an error: an editor that refuses to open
// a folder because git is unhappy is worse than an editor without status badges.
func StatusMap(cwd string) map[string]Status {
	command := exec.Command("git", "status", "--porcelain", "-z")
	command.Dir = cwd

	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		return map[string]Status{}
	}
	return ParsePorcelainZ(stdout.Bytes())
}

// ParsePorcelainZ reads `git status --porcelain -z` output.
//
// Records are NUL-separated. A rename or copy is followed by a second record
// holding the original path, which is consumed here so it is not mistaken for a
// file of its own.
func ParsePorcelainZ(stdout []byte) map[string]Status {
	out := map[string]Status{}
	records := bytes.Split(stdout, []byte{0})

	for index := 0; index < len(records); index++ {
		record := records[index]
		if len(record) < 3 {
			continue
		}

		xy := record[:2]
		path := strings.TrimSpace(string(record[3:]))
		if path == "" {
			continue
		}
		status := classify(xy)

		if isRenameOrCopy(xy) {
			index++
		}
		if existing, ok := out[path]; ok {
			out[path] = worse(existing, status)
		} else {
			out[path] = status
		}
	}
	return out
}

func isRenameOrCopy(xy []byte) bool {
	return xy[0] == 'R' || xy[1] == 'R' || xy[0] == 'C' || xy[1] == 'C'
}

// classify maps the two porcelain status letters to one status.
//
// Order matters and mirrors the reference: a conflict outranks everything, and
// an untracked file has no index entry to compare against.
func classify(xy []byte) Status {
	x, y := xy[0], xy[1]

	if x == 'U' || y == 'U' || (x == 'A' && y == 'A') || (x == 'D' && y == 'D') {
		return Conflict
	}
	if x == '?' || y == '?' {
		return Untracked
	}
	if x == 'R' || y == 'R' {
		return Renamed
	}
	if x == 'A' || y == 'A' {
		return Added
	}
	if x == 'D' || y == 'D' {
		return Deleted
	}
	return Modified
}

// Entries returns statuses sorted by path.
//
// Sorted because this feeds the SCM panel and the state dump, and an order that
// depends on map iteration makes two runs of the same program look different.
func Entries(cwd string) []Entry {
	statuses := StatusMap(cwd)
	out := make([]Entry, 0, len(statuses))
	for path, status := range statuses {
		out = append(out, Entry{Status: status, Path: path})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

// EntriesRelative returns statuses with paths relative to a workspace.
//
// Git reports paths relative to the repository root, which is not necessarily
// the workspace: a workspace inside a larger repository would otherwise show
// paths that do not match anything in its tree.
func EntriesRelative(cwd, workspace string) []Entry {
	entries := Entries(cwd)
	root := repositoryRoot(cwd)
	if root == "" {
		return entries
	}
	out := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		absolute := filepath.Join(root, entry.Path)
		relative, err := filepath.Rel(workspace, absolute)
		if err != nil {
			out = append(out, entry)
			continue
		}
		out = append(out, Entry{Status: entry.Status, Path: filepath.ToSlash(relative)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

// GitError reports a git command that failed.
type GitError struct {
	Code    int
	Message string
}

func (e GitError) Error() string {
	return fmt.Sprintf("git falhou (%d): %s", e.Code, e.Message)
}

// ErrEmptyCommitMessage rejects an empty commit message before git does.
var ErrEmptyCommitMessage = errors.New("a mensagem de commit não pode ser vazia")

func repositoryRoot(cwd string) string {
	command := exec.Command("git", "rev-parse", "--show-toplevel")
	command.Dir = cwd

	var stdout bytes.Buffer
	command.Stdout = &stdout
	if err := command.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(stdout.String())
}

// Branch returns the current branch, or an empty string outside a repository.
func Branch(cwd string) string {
	command := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	command.Dir = cwd

	var stdout bytes.Buffer
	command.Stdout = &stdout
	if err := command.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(stdout.String())
}

// AheadBehind returns how many commits the branch is ahead and behind upstream.
func AheadBehind(cwd string) (int, int, bool) {
	command := exec.Command("git", "rev-list", "--left-right", "--count", "@{upstream}...HEAD")
	command.Dir = cwd

	var stdout bytes.Buffer
	command.Stdout = &stdout
	if err := command.Run(); err != nil {
		return 0, 0, false
	}
	return parseAheadBehind(stdout.String())
}

// parseAheadBehind reads `git rev-list --left-right --count` output, which is
// "behind\tahead".
func parseAheadBehind(raw string) (int, int, bool) {
	fields := strings.Fields(strings.TrimSpace(raw))
	if len(fields) != 2 {
		return 0, 0, false
	}
	var behind, ahead int
	if _, err := fmt.Sscanf(fields[0], "%d", &behind); err != nil {
		return 0, 0, false
	}
	if _, err := fmt.Sscanf(fields[1], "%d", &ahead); err != nil {
		return 0, 0, false
	}
	return ahead, behind, true
}

// Stage adds a path to the index.
func Stage(cwd, path string) error {
	return run(cwd, "add", "--", path)
}

// Unstage removes a path from the index.
func Unstage(cwd, path string) error {
	return run(cwd, "restore", "--staged", "--", path)
}

// Commit records the staged changes.
func Commit(cwd, message string) (string, error) {
	if strings.TrimSpace(message) == "" {
		return "", ErrEmptyCommitMessage
	}
	command := exec.Command("git", "commit", "-m", message)
	command.Dir = cwd

	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return "", GitError{Code: exitCode(err), Message: strings.TrimSpace(stderr.String())}
	}
	return strings.TrimSpace(stdout.String()), nil
}

func run(cwd string, args ...string) error {
	command := exec.Command("git", args...)
	command.Dir = cwd

	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return GitError{Code: exitCode(err), Message: strings.TrimSpace(stderr.String())}
	}
	return nil
}

func exitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}
