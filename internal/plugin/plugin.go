// Package plugin extends the editor from outside it.
//
// An extension is a manifest and a process, never code loaded into the editor.
// A theme does not need a shell; a formatter does not need the buffer in memory.
// Keeping that boundary is what makes installing a plugin different from
// granting it access — a distinction the reference documented wrongly as
// "no external plugins at all" while shipping them — ledger D2.
package plugin

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Manifest is a `plugin.toml`.
type Manifest struct {
	Plugin   Header    `toml:"plugin"`
	Commands []Command `toml:"commands"`
	Hooks    *Hooks    `toml:"hooks"`
}

// Header identifies a plugin.
type Header struct {
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Description string `toml:"description"`
}

// Command is one executable the palette can run.
type Command struct {
	ID          string   `toml:"id"`
	Label       string   `toml:"label"`
	Description string   `toml:"description"`
	Executable  string   `toml:"executable"`
	Args        []string `toml:"args"`
}

// Hooks are the lifecycle commands a plugin can register.
type Hooks struct {
	OnOpen *HookCommand `toml:"on_open"`
	OnSave *HookCommand `toml:"on_save"`
}

// HookCommand is an executable run at a lifecycle point.
type HookCommand struct {
	Executable string   `toml:"executable"`
	Args       []string `toml:"args"`
}

// Context is what a command is expanded against.
type Context struct {
	// File is the active document's path, empty when it has none.
	File string
	// Workspace is the project root.
	Workspace string
}

// Variables a command's arguments may reference.
const (
	VarFile      = "$FILE"
	VarDir       = "$DIR"
	VarWorkspace = "$WORKSPACE"
)

// ErrNoExecutable rejects a manifest that would run nothing.
var ErrNoExecutable = errors.New("o manifesto não declara `executable`")

// ParseManifest reads a manifest from TOML.
func ParseManifest(content string) (Manifest, error) {
	var manifest Manifest
	if err := toml.Unmarshal([]byte(content), &manifest); err != nil {
		return Manifest{}, fmt.Errorf("manifesto inválido: %w", err)
	}
	if strings.TrimSpace(manifest.Plugin.Name) == "" {
		// Without a name the plugin cannot be addressed in the palette, and two
		// anonymous plugins would collide.
		return Manifest{}, fmt.Errorf("manifesto sem `plugin.name`")
	}
	for i, command := range manifest.Commands {
		if strings.TrimSpace(command.ID) == "" {
			return Manifest{}, fmt.Errorf("comando %d sem `id`", i+1)
		}
		if strings.TrimSpace(command.Executable) == "" {
			return Manifest{}, fmt.Errorf("comando %q: %w", command.ID, ErrNoExecutable)
		}
	}
	// Checked in a fixed order: iterating a map would make the reported error
	// depend on the run.
	for _, named := range []struct {
		name string
		hook *HookCommand
	}{
		{"on_open", manifest.Hooks.onOpen()},
		{"on_save", manifest.Hooks.onSave()},
	} {
		if named.hook != nil && strings.TrimSpace(named.hook.Executable) == "" {
			return Manifest{}, fmt.Errorf("hook %q: %w", named.name, ErrNoExecutable)
		}
	}
	return manifest, nil
}

func (h *Hooks) onOpen() *HookCommand {
	if h == nil {
		return nil
	}
	return h.OnOpen
}

func (h *Hooks) onSave() *HookCommand {
	if h == nil {
		return nil
	}
	return h.OnSave
}

// LoadManifest reads a `plugin.toml`.
func LoadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("lendo %s: %w", path, err)
	}
	manifest, err := ParseManifest(string(data))
	if err != nil {
		return Manifest{}, fmt.Errorf("%s: %w", path, err)
	}
	return manifest, nil
}

// Plugin is a discovered plugin and where it came from.
type Plugin struct {
	Manifest Manifest
	Dir      string
}

// Discover finds every plugin under the given directories.
//
// Two layouts are accepted: `<dir>/plugins/<name>/plugin.toml` and
// `<dir>/<name>/plugin.toml`, because a user directory is usually dedicated to
// plugins while a workspace's `.oride` also holds other things.
//
// Results are sorted by plugin name. Order is observable — it is the palette
// order — and a directory listing is not.
func Discover(dirs []string) []Plugin {
	found := map[string]Plugin{}

	for _, dir := range dirs {
		for _, candidate := range pluginManifestPaths(dir) {
			manifest, err := LoadManifest(candidate)
			if err != nil {
				continue
			}
			name := manifest.Plugin.Name
			if _, exists := found[name]; exists {
				continue
			}
			found[name] = Plugin{Manifest: manifest, Dir: filepath.Dir(candidate)}
		}
	}

	names := make([]string, 0, len(found))
	for name := range found {
		names = append(names, name)
	}
	sort.Strings(names)

	out := make([]Plugin, 0, len(names))
	for _, name := range names {
		out = append(out, found[name])
	}
	return out
}

// pluginManifestPaths lists candidate `plugin.toml` locations under a directory.
func pluginManifestPaths(dir string) []string {
	var out []string

	// `<dir>/plugins/<name>/plugin.toml`
	pluginsDir := filepath.Join(dir, "plugins")
	if entries, err := os.ReadDir(pluginsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				out = append(out, filepath.Join(pluginsDir, entry.Name(), "plugin.toml"))
			}
		}
	}

	// `<dir>/<name>/plugin.toml`
	if entries, err := os.ReadDir(dir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() != "plugins" {
				out = append(out, filepath.Join(dir, entry.Name(), "plugin.toml"))
			}
		}
	}
	return out
}

// Expand substitutes the context variables in one argument.
//
// `$FILE` is replaced first because it is a prefix of nothing else here, and the
// order is fixed so the same argument always expands the same way.
func Expand(arg string, ctx Context) string {
	expanded := strings.ReplaceAll(arg, VarFile, ctx.File)
	expanded = strings.ReplaceAll(expanded, VarDir, filepath.Dir(ctx.File))
	return strings.ReplaceAll(expanded, VarWorkspace, ctx.Workspace)
}

// Run executes a command, expanding its arguments.
//
// The executable and its arguments are passed separately, never assembled into a
// shell string: a file name is user data, and a name containing a semicolon must
// be a file name rather than a second command.
func Run(command Command, ctx Context) (string, error) {
	return execute(command.Executable, expandAll(command.Args, ctx), ctx.Workspace)
}

// RunHook executes a lifecycle hook.
func RunHook(hook HookCommand, ctx Context) (string, error) {
	return execute(hook.Executable, expandAll(hook.Args, ctx), ctx.Workspace)
}

func expandAll(args []string, ctx Context) []string {
	out := make([]string, 0, len(args))
	for _, arg := range args {
		out = append(out, Expand(arg, ctx))
	}
	return out
}

// execute runs a program and returns its combined output.
//
// Fail closed: a missing executable or a failing program becomes an error the
// caller turns into a status line, never a crash and never a silent success.
func execute(executable string, args []string, dir string) (string, error) {
	if strings.TrimSpace(executable) == "" {
		return "", ErrNoExecutable
	}

	command := exec.Command(executable, args...)
	if dir != "" {
		command.Dir = dir
	}

	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output

	if err := command.Run(); err != nil {
		message := strings.TrimSpace(output.String())
		if message == "" {
			message = err.Error()
		}
		return "", fmt.Errorf("%s falhou: %s", executable, message)
	}
	return strings.TrimSpace(output.String()), nil
}

// SearchDirs returns the directories a plugin can be installed in, in the order
// they are searched.
//
// User before workspace: a project overrides a user's plugin only by declaring
// the same name, which Discover resolves by keeping the first found.
func SearchDirs(workspace string) []string {
	out := make([]string, 0, 3)

	if base, err := os.UserConfigDir(); err == nil {
		out = append(out, filepath.Join(base, "oride", "plugins"))
	}
	if workspace != "" {
		out = append(out, filepath.Join(workspace, ".oride", "plugins"))
		out = append(out, filepath.Join(workspace, ".oride"))
	}
	return out
}
