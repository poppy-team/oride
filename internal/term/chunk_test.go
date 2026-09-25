package term

import "testing"

// apply is a convenience for the many cases that only need the resulting text.
func apply(t *testing.T, scrollback string, col int, text string) (string, int) {
	t.Helper()
	ApplyChunk(&scrollback, &col, text)
	return scrollback, col
}

// TestCarriageReturnAndBackspace reproduces the reference's own test: an
// overwriting progress bar, a backspace, and a prompt redrawn over itself with a
// trailing space.
func TestCarriageReturnAndBackspace(t *testing.T) {
	var scrollback string
	var col int

	ApplyChunk(&scrollback, &col, "loading 10%\rloading 50%\rloading 100%\n")
	if scrollback != "loading 100%\n" {
		t.Errorf("barra de progresso: %q", scrollback)
	}

	ApplyChunk(&scrollback, &col, "sh-5.2$ el\x08cho oi\r\n")
	if scrollback != "loading 100%\nsh-5.2$ echo oi\n" {
		t.Errorf("backspace: %q", scrollback)
	}

	ApplyChunk(&scrollback, &col, "%        \r \ruser@arch$ cargo")
	if scrollback != "loading 100%\nsh-5.2$ echo oi\nuser@arch$ cargo" {
		t.Errorf("prompt sobrescrito: %q", scrollback)
	}
}

// TestPromptWithANSISequences reproduces the reference's second test: the erase
// and cursor-movement codes zsh emits to redraw a prompt, and colour sequences,
// which must change nothing about the text.
func TestPromptWithANSISequences(t *testing.T) {
	var scrollback string
	var col int

	ApplyChunk(&scrollback, &col, "%\x1b[0m               \r \r\x1b[Juser@arch:~$ \x1b[K")
	if scrollback != "user@arch:~$ " {
		t.Errorf("prompt zsh: %q", scrollback)
	}
	if col != 13 {
		t.Errorf("coluna = %d, esperado 13", col)
	}

	// Typing, with the syntax highlighting the shell rewrites as it goes.
	ApplyChunk(&scrollback, &col, "echo\x1b[4D\x1b[32me\x1b[32mc\x1b[32mh\x1b[32mo\x1b[39m")
	if scrollback != "user@arch:~$ echo" {
		t.Errorf("digitação: %q", scrollback)
	}
	if col != 17 {
		t.Errorf("coluna = %d, esperado 17", col)
	}
}

func TestCarriageReturnWithoutNewlineOverwrites(t *testing.T) {
	var scrollback string
	var col int

	ApplyChunk(&scrollback, &col, "abcdef\rxy")
	if scrollback != "xycdef" {
		t.Errorf("%q, esperado \"xycdef\" — o retorno de carro não apaga", scrollback)
	}
}

func TestNewlineSettlesTheLine(t *testing.T) {
	var scrollback string
	var col int

	ApplyChunk(&scrollback, &col, "um\ndois")
	if scrollback != "um\ndois" {
		t.Errorf("%q", scrollback)
	}
	if col != 4 {
		t.Errorf("coluna = %d, esperado 4", col)
	}
}

func TestTabAdvancesToTheNextStopAndOverwrites(t *testing.T) {
	// A tab from column zero reaches column eight.
	got, col := apply(t, "", 0, "ab\tc")
	if got != "ab      c" {
		t.Errorf("%q, esperado \"ab      c\"", got)
	}
	if col != 9 {
		t.Errorf("coluna = %d, esperado 9", col)
	}

	// Tab stops are a grid, not a fixed count of spaces: from column seven the
	// next stop is column eight.
	got, col = apply(t, "", 0, "abcdefg\th")
	if got != "abcdefg h" {
		t.Errorf("%q, esperado \"abcdefg h\"", got)
	}
	if col != 9 {
		t.Errorf("coluna = %d, esperado 9", col)
	}
}

func TestEraseInLine(t *testing.T) {
	cases := []struct {
		name       string
		input      string
		want       string
		wantColumn int
	}{
		{"K sem parâmetro apaga da coluna até o fim", "abcdef\x1b[3D\x1b[K", "abc", 3},
		{"K=0 é o mesmo que K", "abcdef\x1b[3D\x1b[0K", "abc", 3},
		{"K=2 apaga a linha e volta à coluna zero", "abcdef\x1b[2K", "", 0},
		// `col + 1` characters, so the one under the cursor goes too.
		{"K=1 apaga até a coluna atual, inclusive", "abcdef\x1b[2D\x1b[1K", "     f", 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, col := apply(t, "", 0, tc.input)
			if got != tc.want {
				t.Errorf("texto = %q, esperado %q", got, tc.want)
			}
			if col != tc.wantColumn {
				t.Errorf("coluna = %d, esperado %d", col, tc.wantColumn)
			}
		})
	}
}

func TestEraseInDisplay(t *testing.T) {
	// J=2 clears the line and resets the column.
	got, col := apply(t, "", 0, "abcdef\x1b[2J")
	if got != "" || col != 0 {
		t.Errorf("%q, coluna %d; esperado vazio na coluna 0", got, col)
	}

	// J without a parameter erases from the cursor to the end and leaves the
	// cursor where it was.
	got, col = apply(t, "", 0, "abcdef\x1b[3D\x1b[J")
	if got != "abc" {
		t.Errorf("%q, esperado \"abc\"", got)
	}
	if col != 3 {
		t.Errorf("coluna = %d, esperado 3", col)
	}
}

func TestCursorMovement(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  int
	}{
		{"D recua uma coluna por padrão", "abcdef\x1b[D", 5},
		{"D com contagem", "abcdef\x1b[3D", 3},
		{"D não passa de zero", "ab\x1b[9D", 0},
		{"C avança uma coluna por padrão", "ab\x1b[C", 3},
		{"C com contagem", "ab\x1b[4C", 6},
		{"G vai para a coluna absoluta, 1-based", "abcdef\x1b[3G", 2},
		{"H usa a coluna depois do ponto e vírgula", "abcdef\x1b[1;3H", 2},
		{"H sem parâmetros não move", "abc\x1b[H", 3},
		{"crase é o mesmo que G", "abcdef\x1b[2`", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, col := apply(t, "", 0, tc.input); col != tc.want {
				t.Errorf("coluna = %d, esperado %d", col, tc.want)
			}
		})
	}
}

// TestParameterParsingMatchesTheReference: the reference parses the whole
// parameter string as one number and falls back when that fails, rather than
// taking the first. Reproducing it keeps the two identical on the prompts that
// emit multi-parameter sequences.
func TestParameterParsingMatchesTheReference(t *testing.T) {
	// `1;2` does not parse, so it erases as much as `K` — from the cursor on.
	if got, _ := apply(t, "", 0, "abcdef\x1b[3D\x1b[1;2K"); got != "abc" {
		t.Errorf("%q — um parâmetro múltiplo cai no padrão, não no primeiro número", got)
	}

	// A private-mode marker also fails to parse.
	if got, _ := apply(t, "", 0, "abcdef\x1b[3D\x1b[?K"); got != "abc" {
		t.Errorf("%q", got)
	}
}

// TestOSCSequenceTerminatesOnBELOrST: a window-title sequence must be swallowed
// whole, or its payload would appear as text in the panel.
func TestOSCSequenceTerminatesOnBELOrST(t *testing.T) {
	got, col := apply(t, "", 0, "a\x1b]0;título\x07b")
	if got != "ab" {
		t.Errorf("BEL: %q, esperado \"ab\"", got)
	}
	if col != 2 {
		t.Errorf("coluna = %d, esperado 2", col)
	}

	got, _ = apply(t, "", 0, "a\x1b]0;título\x1b\\b")
	if got != "ab" {
		t.Errorf("ST: %q, esperado \"ab\"", got)
	}
}

// TestColourSequencesChangeNothing: the panel is text, so SGR must not alter it.
func TestColourSequencesChangeNothing(t *testing.T) {
	plain, plainCol := apply(t, "", 0, "aviso")
	coloured, colouredCol := apply(t, "", 0, "\x1b[31m\x1b[1maviso\x1b[0m")
	if coloured != plain || colouredCol != plainCol {
		t.Errorf("%q coluna %d; esperado %q coluna %d", coloured, colouredCol, plain, plainCol)
	}
}

func TestPromptRedrawAcrossTwoChunks(t *testing.T) {
	var scrollback string
	var col int

	// A shell writes the prompt, then rewrites it a moment later. The cursor
	// column has to survive between the two writes, which is why it is state.
	ApplyChunk(&scrollback, &col, "user@host$ ")
	ApplyChunk(&scrollback, &col, "\r\x1b[Kuser@host:~$ ")
	if scrollback != "user@host:~$ " {
		t.Errorf("%q", scrollback)
	}
}

func TestOtherControlCharactersAreIgnored(t *testing.T) {
	got, col := apply(t, "", 0, "a\x01\x02b")
	if got != "ab" || col != 2 {
		t.Errorf("%q na coluna %d, esperado \"ab\" na coluna 2", got, col)
	}
}

func TestApplyChunkWithEmptyTextChangesNothing(t *testing.T) {
	got, col := apply(t, "abc", 3, "")
	if got != "abc" || col != 3 {
		t.Errorf("%q na coluna %d", got, col)
	}
}

// TestTrimScrollbackPreservesRuneBoundaries is the reference's own case: cutting
// at a byte offset would leave a replacement character where a split rune was.
func TestTrimScrollbackPreservesRuneBoundaries(t *testing.T) {
	scrollback := "a✨b✨c"

	trimScrollback(&scrollback, 6)

	if len(scrollback) > 6 {
		t.Errorf("%q tem %d bytes, maior que 6", scrollback, len(scrollback))
	}
	if scrollback != "b✨c" {
		t.Errorf("%q, esperado \"b✨c\"", scrollback)
	}
}

func TestTrimScrollbackLeavesShortTextAlone(t *testing.T) {
	scrollback := "curto"
	trimScrollback(&scrollback, 100)
	if scrollback != "curto" {
		t.Errorf("%q", scrollback)
	}
}

func TestVisibleLinesDropsTheTrailingEmptyLine(t *testing.T) {
	terminal := &Terminal{scrollback: "um\ndois\n"}

	lines := terminal.VisibleLines(10)
	if len(lines) != 2 {
		t.Fatalf("linhas = %q, esperado 2", lines)
	}
	if lines[0] != "um" || lines[1] != "dois" {
		t.Errorf("linhas = %q", lines)
	}
}

func TestVisibleLinesOnEmptyScrollback(t *testing.T) {
	terminal := &Terminal{}
	if lines := terminal.VisibleLines(10); lines != nil {
		t.Errorf("linhas = %q, esperado nada", lines)
	}
}

func TestVisibleLinesTakesTheTail(t *testing.T) {
	terminal := &Terminal{scrollback: "um\ndois\ntres\nquatro"}

	lines := terminal.VisibleLines(2)
	if len(lines) != 2 || lines[0] != "tres" || lines[1] != "quatro" {
		t.Errorf("linhas = %q, esperado as duas últimas", lines)
	}
}

// TestVisibleLinesAlwaysReturnsAtLeastOne: a zero or negative request is a
// caller bug, and showing nothing would look like a broken panel.
func TestVisibleLinesAlwaysReturnsAtLeastOne(t *testing.T) {
	terminal := &Terminal{scrollback: "um\ndois"}

	if lines := terminal.VisibleLines(0); len(lines) != 1 || lines[0] != "dois" {
		t.Errorf("linhas = %q", lines)
	}
}

func TestGrowShowsThePanelAndIsBounded(t *testing.T) {
	terminal := &Terminal{HeightLines: DefaultHeightLines}

	terminal.Grow(5)
	if terminal.HeightLines != DefaultHeightLines+5 {
		t.Errorf("altura = %d", terminal.HeightLines)
	}
	if !terminal.Visible {
		t.Error("Grow deveria mostrar o painel")
	}

	terminal.Grow(1000)
	if terminal.HeightLines != 40 {
		t.Errorf("altura = %d, esperado o teto de 40", terminal.HeightLines)
	}
}

func TestShrinkIsBounded(t *testing.T) {
	terminal := &Terminal{HeightLines: 12}

	terminal.Shrink(100)
	if terminal.HeightLines != 3 {
		t.Errorf("altura = %d, esperado o piso de 3", terminal.HeightLines)
	}
}

func TestToggleVisible(t *testing.T) {
	terminal := &Terminal{}
	terminal.ToggleVisible()
	if !terminal.Visible {
		t.Error("o painel deveria estar visível")
	}
	terminal.ToggleVisible()
	if terminal.Visible {
		t.Error("o painel deveria estar oculto")
	}
}
