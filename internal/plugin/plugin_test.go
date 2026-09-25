package plugin

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const sampleManifest = `
[plugin]
name = "formatter"
version = "0.1.0"
description = "Formata automaticamente"

[[commands]]
id = "fmt"
label = "Plugin: Formatar Arquivo"
description = "Formata o buffer com prettier"
executable = "echo"
args = ["formatando", "$FILE"]

[hooks.on_save]
executable = "echo"
args = ["salvo:", "$FILE"]
`

func TestParseManifestMatchesTheReferenceExample(t *testing.T) {
	manifest, err := ParseManifest(sampleManifest)
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	if manifest.Plugin.Name != "formatter" {
		t.Errorf("name = %q", manifest.Plugin.Name)
	}
	if manifest.Plugin.Version != "0.1.0" {
		t.Errorf("version = %q", manifest.Plugin.Version)
	}
	if len(manifest.Commands) != 1 {
		t.Fatalf("commands = %d", len(manifest.Commands))
	}
	command := manifest.Commands[0]
	if command.ID != "fmt" || command.Label == "" || command.Description == "" {
		t.Errorf("comando = %+v", command)
	}
	if command.Executable != "echo" {
		t.Errorf("executable = %q", command.Executable)
	}
	if len(command.Args) != 2 || command.Args[1] != "$FILE" {
		t.Errorf("args = %v", command.Args)
	}
	if manifest.Hooks == nil || manifest.Hooks.OnSave == nil {
		t.Fatal("o hook on_save não foi lido")
	}
	if manifest.Hooks.OnOpen != nil {
		t.Error("um hook não declarado foi inventado")
	}
}

// TestManifestWithoutANameIsRejected: two anonymous plugins would collide in the
// palette and neither could be addressed.
func TestManifestWithoutANameIsRejected(t *testing.T) {
	_, err := ParseManifest("[plugin]\nversion = \"1\"\n")
	if err == nil {
		t.Fatal("manifesto sem nome foi aceito")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("erro %q não menciona o campo", err)
	}
}

// TestCommandWithoutAnExecutableIsRejected: a command that runs nothing looks
// exactly like a command that worked.
func TestCommandWithoutAnExecutableIsRejected(t *testing.T) {
	_, err := ParseManifest(`
[plugin]
name = "vazio"

[[commands]]
id = "nada"
label = "Nada"
`)
	if err == nil {
		t.Fatal("comando sem executable foi aceito")
	}
}

// TestExpandSubstitutesEveryVariable pins the three names a manifest may use.
func TestExpandSubstitutesEveryVariable(t *testing.T) {
	ctx := Context{File: "/projeto/src/main.rs", Workspace: "/projeto"}

	cases := map[string]string{
		"$FILE":           "/projeto/src/main.rs",
		"$DIR":            "/projeto/src",
		"$WORKSPACE":      "/projeto",
		"--write $FILE":   "--write /projeto/src/main.rs",
		"$WORKSPACE/$DIR": "/projeto//projeto/src",
		"sem variavel":    "sem variavel",
	}
	for input, want := range cases {
		if got := Expand(input, ctx); got != want {
			t.Errorf("Expand(%q) = %q, esperado %q", input, got, want)
		}
	}
}

// TestExpandWithNoActiveFile keeps a variable from becoming the string "$FILE"
// or a bare directory when there is no document open.
func TestExpandWithNoActiveFile(t *testing.T) {
	ctx := Context{Workspace: "/projeto"}
	if got := Expand("$FILE", ctx); got != "" {
		t.Errorf("$FILE sem arquivo = %q, esperado vazio", got)
	}
	if got := Expand("$DIR", ctx); got != "." {
		t.Errorf("$DIR sem arquivo = %q, esperado \".\"", got)
	}
}

func TestDiscoverFindsThePluginsLayout(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "plugins", "formatter")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("criando: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.toml"), []byte(sampleManifest), 0o644); err != nil {
		t.Fatalf("escrevendo: %v", err)
	}

	plugins := Discover([]string{dir})
	if len(plugins) != 1 {
		t.Fatalf("plugins = %d, esperado 1", len(plugins))
	}
	if plugins[0].Manifest.Plugin.Name != "formatter" {
		t.Errorf("nome = %q", plugins[0].Manifest.Plugin.Name)
	}
	if plugins[0].Dir != pluginDir {
		t.Errorf("dir = %q, esperado %q", plugins[0].Dir, pluginDir)
	}
}

// TestDiscoverAlsoAcceptsTheFlatLayout: a workspace's `.oride` holds other
// things besides plugins, so the name directory sits directly under it.
func TestDiscoverAlsoAcceptsTheFlatLayout(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "meu-plugin")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("criando: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.toml"), []byte(sampleManifest), 0o644); err != nil {
		t.Fatalf("escrevendo: %v", err)
	}

	if plugins := Discover([]string{dir}); len(plugins) != 1 {
		t.Fatalf("plugins = %d, esperado 1", len(plugins))
	}
}

// TestDiscoverIsSortedAndDeduplicated: the order is the palette order, and a
// directory listing is not an order.
func TestDiscoverIsSortedAndDeduplicated(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"zebra", "alfa", "meio"} {
		pluginDir := filepath.Join(dir, "plugins", name)
		if err := os.MkdirAll(pluginDir, 0o755); err != nil {
			t.Fatalf("criando: %v", err)
		}
		content := strings.Replace(sampleManifest, `name = "formatter"`, `name = "`+name+`"`, 1)
		if err := os.WriteFile(filepath.Join(pluginDir, "plugin.toml"), []byte(content), 0o644); err != nil {
			t.Fatalf("escrevendo: %v", err)
		}
	}

	plugins := Discover([]string{dir, dir})
	if len(plugins) != 3 {
		t.Fatalf("plugins = %d, esperado 3 (o mesmo diretório duas vezes não duplica)", len(plugins))
	}
	names := []string{plugins[0].Manifest.Plugin.Name, plugins[1].Manifest.Plugin.Name, plugins[2].Manifest.Plugin.Name}
	if names[0] != "alfa" || names[1] != "meio" || names[2] != "zebra" {
		t.Errorf("ordem = %v, esperada alfabética", names)
	}
}

// TestAMalformedManifestIsSkippedRatherThanFatal: one broken plugin must not
// take the editor's start with it.
func TestAMalformedManifestIsSkippedRatherThanFatal(t *testing.T) {
	dir := t.TempDir()
	broken := filepath.Join(dir, "plugins", "quebrado")
	good := filepath.Join(dir, "plugins", "bom")
	for _, target := range []string{broken, good} {
		if err := os.MkdirAll(target, 0o755); err != nil {
			t.Fatalf("criando: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(broken, "plugin.toml"), []byte("isto ][ não é toml"), 0o644); err != nil {
		t.Fatalf("escrevendo: %v", err)
	}
	if err := os.WriteFile(filepath.Join(good, "plugin.toml"), []byte(sampleManifest), 0o644); err != nil {
		t.Fatalf("escrevendo: %v", err)
	}

	plugins := Discover([]string{dir})
	if len(plugins) != 1 {
		t.Fatalf("plugins = %d, esperado só o válido", len(plugins))
	}
}

func TestDiscoverOnAMissingDirectory(t *testing.T) {
	if plugins := Discover([]string{filepath.Join(t.TempDir(), "nao-existe")}); len(plugins) != 0 {
		t.Errorf("plugins = %d, esperado nenhum", len(plugins))
	}
}

// TestRunExpandsArguments: the executable and its arguments are passed
// separately, never assembled into a shell string.
func TestRunExpandsArguments(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("o teste usa `echo` do POSIX")
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "notas.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("escrevendo: %v", err)
	}

	output, err := Run(Command{
		ID:         "echo",
		Executable: "echo",
		Args:       []string{"arquivo:", "$FILE", "pasta:", "$DIR"},
	}, Context{File: file, Workspace: dir})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(output, file) {
		t.Errorf("saída %q não contém o caminho expandido", output)
	}
	if !strings.Contains(output, dir) {
		t.Errorf("saída %q não contém o diretório", output)
	}
}

// TestRunFailsClosed: a missing executable is an error the caller turns into a
// status line, not a crash and not a silent success.
func TestRunFailsClosed(t *testing.T) {
	_, err := Run(Command{ID: "x", Executable: "programa-que-nao-existe-em-lugar-nenhum"}, Context{})
	if err == nil {
		t.Fatal("executável ausente não produziu erro")
	}
	if !strings.Contains(err.Error(), "programa-que-nao-existe") {
		t.Errorf("erro %q não nomeia o programa", err)
	}
}

func TestRunWithoutAnExecutable(t *testing.T) {
	if _, err := Run(Command{ID: "x"}, Context{}); err == nil {
		t.Fatal("comando sem executável foi aceito")
	}
}

// TestArgumentsAreNotShellInterpreted: a file name is user data. A name with a
// semicolon is a file name, not a second command.
func TestArgumentsAreNotShellInterpreted(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("o teste usa `echo` do POSIX")
	}

	output, err := Run(Command{
		ID:         "echo",
		Executable: "echo",
		Args:       []string{"$FILE"},
	}, Context{File: "arquivo;echo INJETADO"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if strings.Contains(output, "INJETADO\n") && !strings.Contains(output, "arquivo;echo INJETADO") {
		t.Errorf("o argumento foi interpretado pelo shell: %q", output)
	}
	if output != "arquivo;echo INJETADO" {
		t.Errorf("saída = %q, esperado o argumento literal", output)
	}
}

func TestSearchDirsPutsTheUserFirst(t *testing.T) {
	dirs := SearchDirs("/projeto")
	if len(dirs) == 0 {
		t.Skip("sem diretório de configuração nesta máquina")
	}
	if !strings.Contains(dirs[len(dirs)-1], "/projeto") {
		t.Errorf("o workspace deveria ser o último a ser procurado: %v", dirs)
	}
}
