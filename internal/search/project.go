package search

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Backend names how a project search ran.
type Backend string

// Backends, in the order they are tried.
const (
	// BackendRipgrep shells out to `rg`, which brings full gitignore semantics
	// and is fast enough that no in-process scanner competes with it.
	BackendRipgrep Backend = "ripgrep"
	// BackendWalk is the fallback. Its ignore handling is deliberately simpler,
	// and it says so rather than pretending to be equivalent.
	BackendWalk Backend = "walk"
)

// Query is a project search.
type Query struct {
	Text          string
	CaseSensitive bool
	UseRegex      bool
	// Globs filters which files are searched. A leading `!` excludes. Empty
	// means every file the backend would consider.
	Globs []string
	// MaxHits bounds the result. Zero means the default.
	MaxHits int
}

// Hit is one match.
type Hit struct {
	// Path is relative to the search root, with forward slashes.
	Path string
	// Line is 1-based, as every editor and compiler reports it.
	Line int
	// Column is 1-based, counted in bytes from the start of the line.
	Column int
	// Text is the whole line, without its line break.
	Text string
}

// DefaultMaxHits bounds a search that would otherwise return a whole repository.
const DefaultMaxHits = 2000

// ErrEmptyQuery rejects a search that would match everything.
var ErrEmptyQuery = errors.New("consulta vazia")

// FindProject searches a directory tree.
//
// The backend that ran is returned alongside the hits, because the two do not
// ignore the same files: a caller that wants to tell the user why a file was
// skipped needs to know which one answered.
func FindProject(root string, query Query) ([]Hit, Backend, error) {
	if strings.TrimSpace(query.Text) == "" {
		return nil, "", ErrEmptyQuery
	}
	if query.MaxHits <= 0 {
		query.MaxHits = DefaultMaxHits
	}

	if hasRipgrep() {
		hits, err := searchWithRipgrep(root, query)
		if err == nil {
			return hits, BackendRipgrep, nil
		}
		// A failing rg is not fatal: the fallback still answers, and refusing to
		// search because a helper is unhappy would be worse.
	}

	hits, err := searchWithWalk(root, query)
	return hits, BackendWalk, err
}

// hasRipgrep reports whether `rg` is on the path.
func hasRipgrep() bool {
	_, err := exec.LookPath("rg")
	return err == nil
}

// ripgrepEvent is the subset of `rg --json` this package reads.
type ripgrepEvent struct {
	Type string `json:"type"`
	Data struct {
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
		Lines struct {
			Text string `json:"text"`
		} `json:"lines"`
		LineNumber     int `json:"line_number"`
		AbsoluteOffset int `json:"absolute_offset"`
		Submatches     []struct {
			Start int `json:"start"`
		} `json:"submatches"`
	} `json:"data"`
}

// searchWithRipgrep runs `rg --json`.
//
// The JSON format rather than `path:line:text`, because a file name may contain
// a colon and a naive split would report a nonsense path — a bug that only
// appears on somebody else's machine.
func searchWithRipgrep(root string, query Query) ([]Hit, error) {
	args := []string{"--json", "--line-number", "--color", "never"}

	if !query.CaseSensitive {
		args = append(args, "--ignore-case")
	}
	if !query.UseRegex {
		args = append(args, "--fixed-strings")
	}
	for _, glob := range query.Globs {
		trimmed := strings.TrimSpace(glob)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "!") {
			args = append(args, "--glob", "!"+strings.TrimPrefix(trimmed, "!"))
			continue
		}
		args = append(args, "--glob", trimmed)
	}
	// `--` so a query beginning with a dash is a pattern, not a flag.
	args = append(args, "--", query.Text, ".")

	command := exec.Command("rg", args...)
	command.Dir = root

	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	// Exit code 1 means "no matches", which is an answer rather than a failure.
	if err := command.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
			return nil, fmt.Errorf("rg: %w (%s)", err, strings.TrimSpace(stderr.String()))
		}
	}

	return parseRipgrepJSON(stdout.Bytes(), query.MaxHits)
}

// parseRipgrepJSON reads `rg --json` output.
func parseRipgrepJSON(data []byte, maxHits int) ([]Hit, error) {
	hits := make([]Hit, 0, 64)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)

	for scanner.Scan() {
		var event ripgrepEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			// One unreadable line should not discard the rest of the answer.
			continue
		}
		if event.Type != "match" {
			continue
		}

		column := 1
		if len(event.Data.Submatches) > 0 {
			column = event.Data.Submatches[0].Start + 1
		}
		hits = append(hits, Hit{
			Path:   filepath.ToSlash(filepath.Clean(event.Data.Path.Text)),
			Line:   event.Data.LineNumber,
			Column: column,
			Text:   strings.TrimRight(event.Data.Lines.Text, "\r\n"),
		})
		if len(hits) >= maxHits {
			break
		}
	}
	return hits, scanner.Err()
}

// ignoreDirectories are skipped by the fallback at every level.
//
// The fallback does not implement gitignore semantics — `rg` does, and it is the
// path taken whenever it is installed. What is skipped here is what no project
// wants searched, so a machine without `rg` still gets a useful answer instead of
// a slow one over build output.
var ignoreDirectories = map[string]bool{
	".git":         true,
	"node_modules": true,
	"target":       true,
	"vendor":       true,
	".oride":       true,
}

// searchWithWalk is the fallback: a bounded recursive walk.
func searchWithWalk(root string, query Query) ([]Hit, error) {
	matcher, err := compileMatcher(query)
	if err != nil {
		return nil, err
	}

	hits := make([]Hit, 0, 64)
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// An unreadable directory is skipped, not fatal: a permission
			// problem somewhere should not abort a repository-wide search.
			if entry != nil && entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		name := entry.Name()
		// The root is never skipped, even when it is itself a hidden directory:
		// a workspace under `~/.config` would otherwise search nothing.
		if entry.IsDir() {
			if path == root {
				return nil
			}
			if ignoreDirectories[name] || strings.HasPrefix(name, ".") {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(name, ".") {
			return nil
		}

		relative, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		relative = filepath.ToSlash(relative)
		if !globAllows(query.Globs, relative) {
			return nil
		}

		found, err := searchFile(path, relative, matcher, query.MaxHits-len(hits))
		if err != nil {
			return nil
		}
		hits = append(hits, found...)

		if len(hits) >= query.MaxHits {
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Path != hits[j].Path {
			return hits[i].Path < hits[j].Path
		}
		return hits[i].Line < hits[j].Line
	})
	return hits, nil
}

// compileMatcher builds the per-line matcher.
func compileMatcher(query Query) (*regexp.Regexp, error) {
	pattern := query.Text
	if !query.UseRegex {
		pattern = regexp.QuoteMeta(pattern)
	}
	if !query.CaseSensitive {
		pattern = "(?i)" + pattern
	}
	return regexp.Compile(pattern)
}

// searchFile scans one file.
func searchFile(path, relative string, matcher *regexp.Regexp, budget int) ([]Hit, error) {
	if budget <= 0 {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}
	// A NUL byte in the first block means a binary file, which has no lines to
	// report and would otherwise produce hits full of control characters.
	if bytes.IndexByte(data[:min(len(data), 8000)], 0) >= 0 {
		return nil, nil
	}

	hits := make([]Hit, 0, 8)
	for index, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		location := matcher.FindStringIndex(line)
		if location == nil {
			continue
		}
		hits = append(hits, Hit{
			Path:   relative,
			Line:   index + 1,
			Column: location[0] + 1,
			Text:   line,
		})
		if len(hits) >= budget {
			break
		}
	}
	return hits, nil
}

// globAllows applies the include and exclude patterns to a path.
//
// Patterns are matched against the relative path and against the base name, so
// `*.rs` works the way someone writing it means, and `src/**` works too.
func globAllows(globs []string, relative string) bool {
	if len(globs) == 0 {
		return true
	}

	base := filepath.Base(relative)
	included := false
	hasInclude := false

	for _, glob := range globs {
		pattern := strings.TrimSpace(glob)
		if pattern == "" {
			continue
		}

		if strings.HasPrefix(pattern, "!") {
			if matchesGlob(strings.TrimPrefix(pattern, "!"), relative, base) {
				return false
			}
			continue
		}

		hasInclude = true
		if matchesGlob(pattern, relative, base) {
			included = true
		}
	}

	return included || !hasInclude
}

func matchesGlob(pattern, relative, base string) bool {
	if matched, err := filepath.Match(pattern, base); err == nil && matched {
		return true
	}
	if matched, err := filepath.Match(pattern, relative); err == nil && matched {
		return true
	}
	// A `dir/**` prefix matches everything under it.
	if prefix, ok := strings.CutSuffix(pattern, "/**"); ok {
		return strings.HasPrefix(relative, strings.TrimSuffix(prefix, "/")+"/")
	}
	// A `**/` prefix matches at any depth.
	if suffix, ok := strings.CutPrefix(pattern, "**/"); ok {
		if matched, err := filepath.Match(suffix, base); err == nil && matched {
			return true
		}
	}
	return false
}
