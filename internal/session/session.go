// Package session persists what was open between runs.
//
// A session is small and disposable: which files were open, which was active,
// where the view was scrolled, and how the panels were arranged. Losing it costs
// a few keystrokes; corrupting it costs trust, so a session that cannot be read
// is ignored rather than fatal, and one that cannot be written is reported.
package session

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Split is the pane arrangement, when the editor was split.
type Split struct {
	Orientation   string `toml:"orientation"`
	SecondaryFile string `toml:"secondary_file,omitempty"`
	RatioPercent  int    `toml:"ratio_percent"`
}

// Session is what to restore on the next run.
type Session struct {
	Workspace string   `toml:"workspace"`
	Files     []string `toml:"files"`
	// ActiveIndex indexes into Files, not a document id: a document id is a
	// runtime counter and would mean nothing after a restart.
	ActiveIndex int `toml:"active_index"`
	ScrollY     int `toml:"scroll_y,omitempty"`
	// The optional fields are absent from a session written by an older version,
	// and absent is not the same as zero: a tree width of zero would collapse
	// the panel.
	TreeWidth *int   `toml:"tree_width,omitempty"`
	ShowTree  *bool  `toml:"show_tree,omitempty"`
	ShowSCM   *bool  `toml:"show_scm,omitempty"`
	Split     *Split `toml:"split,omitempty"`
}

// FNV-1a 64 parameters.
const (
	fnvOffsetBasis uint64 = 0xcbf29ce484222325
	fnvPrime       uint64 = 0x100000001b3
)

// Digest identifies a workspace by its canonical path.
//
// Not a hash from the standard library: Go's map hash and Rust's `DefaultHasher`
// are both unstable across versions, and this value names a file on disk. A
// digest that changes between releases makes saved sessions disappear, which is
// exactly the defect the reference had — ledger B18. FNV-1a is stable, needs no
// dependency, and is reproducible in any language.
func Digest(canonicalWorkspace string) uint64 {
	hash := fnvOffsetBasis
	for _, b := range []byte(canonicalWorkspace) {
		hash ^= uint64(b)
		hash *= fnvPrime
	}
	return hash
}

// FileName is the session file name for a workspace.
func FileName(canonicalWorkspace string) string {
	return fmt.Sprintf("%016x.toml", Digest(canonicalWorkspace))
}

// PathForWorkspace returns where a workspace's session lives.
//
// A workspace that has a `.oride` directory keeps the session inside it, so the
// state travels with the project. Otherwise it goes to the user's data
// directory, named by the workspace digest.
func PathForWorkspace(workspace string) (string, bool) {
	canonical := canonicalize(workspace)

	localDir := filepath.Join(canonical, ".oride")
	localSession := filepath.Join(localDir, "session.toml")
	if _, err := os.Stat(localSession); err == nil {
		return localSession, true
	}
	if info, err := os.Stat(localDir); err == nil && info.IsDir() {
		return localSession, true
	}

	base, err := os.UserCacheDir()
	if err != nil {
		return "", false
	}
	return filepath.Join(base, "oride", "sessions", FileName(canonical)), true
}

// GlobalPath returns the fallback session, used when no workspace is known.
func GlobalPath() (string, bool) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", false
	}
	return filepath.Join(base, "oride", "session.toml"), true
}

// LoadForWorkspace reads a workspace's session.
//
// A missing or unreadable session is not an error: there is nothing to restore,
// which is the normal first run.
func LoadForWorkspace(workspace string) (Session, bool) {
	path, ok := PathForWorkspace(workspace)
	if !ok {
		return Session{}, false
	}
	return loadFile(path)
}

// Load reads the session for the current directory, then the global one.
func Load() (Session, bool) {
	if cwd, err := os.Getwd(); err == nil {
		if session, ok := LoadForWorkspace(cwd); ok {
			return session, true
		}
	}
	path, ok := GlobalPath()
	if !ok {
		return Session{}, false
	}
	return loadFile(path)
}

func loadFile(path string) (Session, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Session{}, false
	}
	var session Session
	if err := toml.Unmarshal(data, &session); err != nil {
		return Session{}, false
	}
	return session, true
}

// Save writes the session next to the workspace, or globally.
func (s Session) Save() error {
	path, ok := PathForWorkspace(s.Workspace)
	if !ok {
		path, ok = GlobalPath()
		if !ok {
			return nil
		}
	}
	return s.SaveTo(path)
}

// SaveTo writes the session to an explicit path.
func (s Session) SaveTo(path string) error {
	if parent := filepath.Dir(path); parent != "" {
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return fmt.Errorf("criando %s: %w", parent, err)
		}
	}
	data, err := toml.Marshal(s)
	if err != nil {
		return fmt.Errorf("serializando sessão: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("escrevendo %s: %w", path, err)
	}
	return nil
}

// FromWorkspace builds a session from the current state.
func FromWorkspace(workspace string, files []string, activeIndex int) Session {
	return Session{
		Workspace:   workspace,
		Files:       files,
		ActiveIndex: activeIndex,
	}
}

// ActiveFile returns the file that should be focused on restore.
func (s Session) ActiveFile() (string, bool) {
	if s.ActiveIndex < 0 || s.ActiveIndex >= len(s.Files) {
		return "", false
	}
	return s.Files[s.ActiveIndex], true
}

// IntValue reads an optional integer, falling back when absent.
func IntValue(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

// BoolValue reads an optional flag, falling back when absent.
func BoolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

// canonicalize resolves symlinks and makes a path absolute, falling back to the
// input when that fails — a workspace that does not exist yet still needs a name.
func canonicalize(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		if absolute, err := filepath.Abs(resolved); err == nil {
			return absolute
		}
		return resolved
	}
	if absolute, err := filepath.Abs(path); err == nil {
		return absolute
	}
	return path
}
