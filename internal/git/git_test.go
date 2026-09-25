package git

import (
	"testing"
)

// TestParsePorcelainZ covers the format directly. It is fiddly enough that
// reading it wrong produces a plausible-but-wrong status rather than an error:
// records are NUL-separated, and a rename is followed by a second record holding
// the original path.
func TestParsePorcelainZ(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  map[string]Status
	}{
		{
			name:  "modified in the worktree",
			input: " M arquivo.txt\x00",
			want:  map[string]Status{"arquivo.txt": Modified},
		},
		{
			name:  "staged addition",
			input: "A  novo.txt\x00",
			want:  map[string]Status{"novo.txt": Added},
		},
		{
			name:  "untracked",
			input: "?? rascunho.txt\x00",
			want:  map[string]Status{"rascunho.txt": Untracked},
		},
		{
			name:  "deleted",
			input: " D removido.txt\x00",
			want:  map[string]Status{"removido.txt": Deleted},
		},
		{
			name:  "merge conflict outranks the rest",
			input: "UU conflito.txt\x00",
			want:  map[string]Status{"conflito.txt": Conflict},
		},
		{
			name:  "both added is a conflict",
			input: "AA ambos.txt\x00",
			want:  map[string]Status{"ambos.txt": Conflict},
		},
		{
			name:  "both deleted is a conflict",
			input: "DD sumiu.txt\x00",
			want:  map[string]Status{"sumiu.txt": Conflict},
		},
		{
			// The second record is the original path and must not become an
			// entry of its own.
			name:  "rename consumes the original path record",
			input: "R  novo.txt\x00antigo.txt\x00",
			want:  map[string]Status{"novo.txt": Renamed},
		},
		{
			// The source record is consumed either way, so a copy does not
			// produce a phantom entry for the file it came from.
			//
			// Classified as Modified, not Renamed: the reference only recognises
			// `R`, and parity means reproducing that rather than inventing a
			// better answer. A copy reading as "modified" is a cosmetic quirk,
			// not a defect worth a divergence.
			name:  "copy consumes the source record",
			input: "C  copia.txt\x00origem.txt\x00",
			want:  map[string]Status{"copia.txt": Modified},
		},
		{
			name:  "several records",
			input: " M a.txt\x00?? b.txt\x00A  c.txt\x00",
			want: map[string]Status{
				"a.txt": Modified,
				"b.txt": Untracked,
				"c.txt": Added,
			},
		},
		{
			name:  "a path with a space survives",
			input: " M com espaço.txt\x00",
			want:  map[string]Status{"com espaço.txt": Modified},
		},
		{
			name:  "empty output",
			input: "",
			want:  map[string]Status{},
		},
		{
			name:  "short record is ignored rather than crashing",
			input: "M\x00",
			want:  map[string]Status{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParsePorcelainZ([]byte(tc.input))
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for path, wantStatus := range tc.want {
				gotStatus, ok := got[path]
				if !ok {
					t.Errorf("path %q ausente em %v", path, got)
					continue
				}
				if gotStatus != wantStatus {
					t.Errorf("%q = %v, esperado %v", path, gotStatus, wantStatus)
				}
			}
		})
	}
}

// TestWorseKeepsTheMoreAlarmingStatus: a path can appear twice — staged and
// modified, for instance — and showing the milder status would hide the reason
// it needs attention.
func TestWorseKeepsTheMoreAlarmingStatus(t *testing.T) {
	order := []Status{Untracked, Added, Renamed, Modified, Deleted, Conflict}
	for i := 0; i < len(order)-1; i++ {
		milder, worseStatus := order[i], order[i+1]
		if got := worse(milder, worseStatus); got != worseStatus {
			t.Errorf("worse(%v, %v) = %v", milder, worseStatus, got)
		}
		if got := worse(worseStatus, milder); got != worseStatus {
			t.Errorf("worse(%v, %v) = %v", worseStatus, milder, got)
		}
	}
}

func TestBadgesAreDistinct(t *testing.T) {
	seen := map[string]Status{}
	for _, status := range []Status{Modified, Added, Deleted, Untracked, Renamed, Conflict} {
		badge := status.Badge()
		if len(badge) != 1 {
			t.Errorf("badge de %v = %q, esperado um caractere", status, badge)
		}
		if other, ok := seen[badge]; ok {
			t.Errorf("badge %q compartilhado por %v e %v", badge, other, status)
		}
		seen[badge] = status
	}
}

// TestStatusOutsideARepositoryIsEmptyRatherThanFatal: an editor that refuses to
// open a folder because git is unhappy is worse than an editor without badges.
func TestStatusOutsideARepositoryIsEmptyRatherThanFatal(t *testing.T) {
	statuses := StatusMap(t.TempDir())
	if len(statuses) != 0 {
		t.Errorf("esperado mapa vazio fora de um repositório, veio %v", statuses)
	}
	if entries := Entries(t.TempDir()); len(entries) != 0 {
		t.Errorf("esperado nenhuma entrada, veio %v", entries)
	}
	if branch := Branch(t.TempDir()); branch != "" {
		t.Errorf("esperado branch vazio, veio %q", branch)
	}
	if _, _, ok := AheadBehind(t.TempDir()); ok {
		t.Error("esperado sem upstream fora de um repositório")
	}
}

func TestParseAheadBehind(t *testing.T) {
	ahead, behind, ok := parseAheadBehind("2\t3\n")
	if !ok {
		t.Fatal("não parseou")
	}
	if ahead != 3 || behind != 2 {
		t.Errorf("ahead=%d behind=%d, esperado ahead=3 behind=2 (a saída é \"behind\\tahead\")", ahead, behind)
	}

	for _, bad := range []string{"", "só uma coluna", "a b"} {
		if _, _, ok := parseAheadBehind(bad); ok {
			t.Errorf("%q foi aceito como contagem válida", bad)
		}
	}
}

func TestCommitRejectsAnEmptyMessage(t *testing.T) {
	for _, message := range []string{"", "   ", "\n"} {
		if _, err := Commit(t.TempDir(), message); err == nil {
			t.Errorf("mensagem %q foi aceita", message)
		}
	}
}
