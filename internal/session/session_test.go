package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDigestGoldenValues is the cross-language contract.
//
// These three values are pinned in the Rust test of the same name. The digest
// names a file on disk, so if the two implementations disagree, a session written
// by one is invisible to the other — the user's session silently disappears.
func TestDigestGoldenValues(t *testing.T) {
	cases := map[string]uint64{
		"/tmp/oride":          0xD667_B877_1CBE_1AA3,
		"":                    0xCBF2_9CE4_8422_2325,
		"/tmp/test_dir_alpha": 0x980F_B79D_966D_0797,
	}
	for input, want := range cases {
		if got := Digest(input); got != want {
			t.Errorf("Digest(%q) = 0x%016X, esperado 0x%016X", input, got, want)
		}
	}
}

// TestDigestEmptyIsTheOffsetBasis proves the algorithm is plain FNV-1a rather
// than something that merely resembles it.
func TestDigestEmptyIsTheOffsetBasis(t *testing.T) {
	if got := Digest(""); got != fnvOffsetBasis {
		t.Errorf("Digest(\"\") = 0x%016X, esperado o offset basis 0x%016X", got, fnvOffsetBasis)
	}
}

func TestDigestIsStableAndDistinguishing(t *testing.T) {
	first := Digest("/tmp/projeto-a")
	if first != Digest("/tmp/projeto-a") {
		t.Error("o mesmo caminho produziu digests diferentes")
	}
	if first == Digest("/tmp/projeto-b") {
		t.Error("caminhos diferentes colidiram")
	}
}

func TestFileNameIsSixteenHexDigits(t *testing.T) {
	name := FileName("/tmp/oride")
	if name != "d667b8771cbe1aa3.toml" {
		t.Errorf("FileName = %q", name)
	}
	if len(name) != len("d667b8771cbe1aa3.toml") {
		t.Errorf("nome com tamanho inesperado: %q", name)
	}
}

// TestPathPrefersTheWorkspaceDirectory: a project that has a `.oride` directory
// keeps its session inside itself, so the state travels with the project.
func TestPathPrefersTheWorkspaceDirectory(t *testing.T) {
	workspace := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspace, ".oride"), 0o755); err != nil {
		t.Fatalf("criando .oride: %v", err)
	}

	path, ok := PathForWorkspace(workspace)
	if !ok {
		t.Fatal("nenhum caminho devolvido")
	}
	if !strings.HasPrefix(path, workspace) {
		t.Errorf("path = %q, esperado dentro de %q", path, workspace)
	}
	if filepath.Base(path) != "session.toml" {
		t.Errorf("nome = %q", filepath.Base(path))
	}
}

func TestPathFallsBackToTheDataDirectory(t *testing.T) {
	workspace := t.TempDir()
	path, ok := PathForWorkspace(workspace)
	if !ok {
		t.Skip("sem diretório de dados nesta máquina")
	}
	if strings.HasPrefix(path, workspace) {
		t.Errorf("path = %q, não deveria ficar dentro do workspace sem .oride", path)
	}
	if !strings.HasSuffix(path, ".toml") {
		t.Errorf("path = %q, esperado um .toml", path)
	}
}

// TestRoundTripPreservesEveryField: a field that survives the write but not the
// read is a setting the user discovers is not remembered.
func TestRoundTripPreservesEveryField(t *testing.T) {
	workspace := t.TempDir()
	treeWidth := 30
	showTree := true
	showSCM := false

	original := Session{
		Workspace:   workspace,
		Files:       []string{"/tmp/a.md", "/tmp/b.rs"},
		ActiveIndex: 1,
		ScrollY:     12,
		TreeWidth:   &treeWidth,
		ShowTree:    &showTree,
		ShowSCM:     &showSCM,
		Split: &Split{
			Orientation:   "vertical",
			SecondaryFile: "/tmp/b.rs",
			RatioPercent:  60,
		},
	}

	path := filepath.Join(t.TempDir(), "session.toml")
	if err := original.SaveTo(path); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}
	loaded, ok := loadFile(path)
	if !ok {
		t.Fatal("a sessão salva não foi lida")
	}

	if loaded.Workspace != original.Workspace {
		t.Errorf("workspace = %q", loaded.Workspace)
	}
	if len(loaded.Files) != 2 || loaded.Files[0] != "/tmp/a.md" {
		t.Errorf("files = %v", loaded.Files)
	}
	if loaded.ActiveIndex != 1 || loaded.ScrollY != 12 {
		t.Errorf("active_index = %d, scroll_y = %d", loaded.ActiveIndex, loaded.ScrollY)
	}
	if loaded.TreeWidth == nil || *loaded.TreeWidth != 30 {
		t.Errorf("tree_width = %v", loaded.TreeWidth)
	}
	if loaded.ShowTree == nil || !*loaded.ShowTree {
		t.Errorf("show_tree = %v", loaded.ShowTree)
	}
	if loaded.ShowSCM == nil || *loaded.ShowSCM {
		t.Errorf("show_scm = %v", loaded.ShowSCM)
	}
	if loaded.Split == nil || loaded.Split.RatioPercent != 60 {
		t.Errorf("split = %+v", loaded.Split)
	}
}

// TestAbsentOptionalFieldsStayAbsent: `false` and "not set" are different
// answers, and collapsing them would override the product default with a zero.
func TestAbsentOptionalFieldsStayAbsent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.toml")
	minimal := Session{Workspace: "/tmp/ws", Files: []string{"/tmp/a.txt"}, ActiveIndex: 0}
	if err := minimal.SaveTo(path); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("lendo: %v", err)
	}
	for _, absent := range []string{"tree_width", "show_tree", "show_scm", "split"} {
		if strings.Contains(string(data), absent) {
			t.Errorf("campo opcional %q foi escrito apesar de não estar definido:\n%s", absent, data)
		}
	}

	loaded, _ := loadFile(path)
	if loaded.TreeWidth != nil || loaded.ShowTree != nil || loaded.ShowSCM != nil || loaded.Split != nil {
		t.Error("campo opcional ausente voltou preenchido")
	}
}

func TestLoadMissingSessionIsNotAnError(t *testing.T) {
	if _, ok := LoadForWorkspace(filepath.Join(t.TempDir(), "nao-existe")); ok {
		t.Error("sessão inexistente foi reportada como presente")
	}
}

func TestCorruptSessionIsIgnored(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.toml")
	if err := os.WriteFile(path, []byte("isto = não é toml válido ][\n"), 0o644); err != nil {
		t.Fatalf("escrevendo: %v", err)
	}
	if _, ok := loadFile(path); ok {
		t.Error("sessão corrompida foi aceita")
	}
}

func TestActiveFile(t *testing.T) {
	session := Session{Files: []string{"/a", "/b"}, ActiveIndex: 1}
	if got, ok := session.ActiveFile(); !ok || got != "/b" {
		t.Errorf("ActiveFile = %q, %v", got, ok)
	}

	for _, index := range []int{-1, 5} {
		outOfRange := Session{Files: []string{"/a"}, ActiveIndex: index}
		if _, ok := outOfRange.ActiveFile(); ok {
			t.Errorf("índice %d foi aceito", index)
		}
	}
}

func TestIntAndBoolValueFallBack(t *testing.T) {
	if got := IntValue(nil, 28); got != 28 {
		t.Errorf("IntValue(nil) = %d", got)
	}
	value := 0
	if got := IntValue(&value, 28); got != 0 {
		t.Errorf("IntValue(&0) = %d, esperado 0 — zero é um valor, não ausência", got)
	}
	if got := BoolValue(nil, true); !got {
		t.Error("BoolValue(nil, true) = false")
	}
	flag := false
	if got := BoolValue(&flag, true); got {
		t.Error("BoolValue(&false, true) = true")
	}
}

func TestFromWorkspace(t *testing.T) {
	session := FromWorkspace("/tmp/ws", []string{"/tmp/a"}, 0)
	if session.Workspace != "/tmp/ws" || session.ActiveIndex != 0 {
		t.Errorf("sessão = %+v", session)
	}
}
