// Package term embeds an interactive PTY terminal panel.
//
// It is not a terminal emulator, and neither is the reference: it keeps a text
// scrollback and applies the handful of control sequences a shell prompt actually
// emits — carriage returns for progress bars, backspace for editing, the erase and
// cursor-movement codes zsh and fish use to redraw a prompt. This package
// reproduces that behaviour rather than inventing a screen model, so the two
// implementations display the same thing.
package term

import "strings"

// tabWidth is the tab stop interval.
const tabWidth = 8

// ApplyChunk applies one chunk of output to a scrollback and a cursor column.
//
// Chunk boundaries are not line boundaries: a shell writes a prompt in several
// writes, and a CSI sequence can be split across two of them. The reference has
// the same limitation, and it is why the state lives in the caller — the
// scrollback and the column persist between chunks.
func ApplyChunk(scrollback *string, cursorCol *int, text string) {
	if text == "" {
		return
	}

	// Split at the last line break: everything before it is settled, and only
	// the line in progress can still be edited.
	prefix := ""
	current := *scrollback
	if index := strings.LastIndexByte(*scrollback, '\n'); index >= 0 {
		prefix = (*scrollback)[:index+1]
		current = (*scrollback)[index+1:]
	}

	line := []rune(current)
	col := min(*cursorCol, len(line))

	var out strings.Builder
	out.WriteString(prefix)

	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		c := runes[i]

		switch c {
		case '\r':
			if i+1 < len(runes) && runes[i+1] == '\n' {
				i++
				out.WriteString(string(line))
				out.WriteByte('\n')
				line = line[:0]
				col = 0
			} else {
				// A carriage return alone moves to column zero without erasing:
				// that is what makes an overwriting progress bar work.
				col = 0
			}

		case '\n':
			out.WriteString(string(line))
			out.WriteByte('\n')
			line = line[:0]
			col = 0

		case '\b':
			col = max(0, col-1)

		case '\t':
			next := (col + tabWidth) &^ (tabWidth - 1)
			for col < next {
				if col < len(line) {
					line[col] = ' '
				} else {
					line = append(line, ' ')
				}
				col++
			}

		case 0x1b:
			consumed, newLine, newCol := applyEscape(runes[i+1:], line, col)
			i += consumed
			line, col = newLine, newCol

		default:
			if c < 32 {
				// Other ASCII control characters carry no meaning here.
				continue
			}
			if col < len(line) {
				line[col] = c
			} else {
				for len(line) < col {
					line = append(line, ' ')
				}
				line = append(line, c)
			}
			col++
		}
	}

	out.WriteString(string(line))
	*scrollback = out.String()
	*cursorCol = col
}

// applyEscape handles a sequence starting just after ESC and reports how many of
// the following runes it consumed, the resulting line, and the resulting column.
//
// It takes and returns the line because the erase sequences shorten it, and a
// slice cannot be truncated through a copy.
// applyEscape handles a sequence starting just after ESC and reports how many of
// the following runes it consumed, the resulting line, and the resulting column.
//
// The count is one past the last rune of the sequence, because the caller adds it
// to the index of the ESC and then advances — so returning the index of the
// command instead would leave the command to be written as text.
//
// It takes and returns the line because the erase sequences shorten it, and a
// slice cannot be truncated through a copy.
func applyEscape(rest []rune, line []rune, col int) (int, []rune, int) {
	if len(rest) == 0 {
		return 0, line, col
	}

	switch rest[0] {
	case ']':
		// OSC: ESC ] ... terminated by BEL or by ESC backslash.
		for i := 1; i < len(rest); i++ {
			if rest[i] == '\x07' {
				return i + 1, line, col
			}
			if rest[i] == '\x1b' && i+1 < len(rest) && rest[i+1] == '\\' {
				return i + 2, line, col
			}
		}
		return len(rest), line, col

	case '[':
		// CSI: ESC [ params command
		index := 1
		params := strings.Builder{}
		for index < len(rest) && isCSIParameter(rest[index]) {
			params.WriteRune(rest[index])
			index++
		}
		if index >= len(rest) {
			// The command byte has not arrived. The sequence is dropped rather
			// than buffered, matching the reference.
			return len(rest), line, col
		}

		newLine, newCol := applyCSI(rest[index], params.String(), line, col)
		return index + 1, newLine, newCol

	case '=', '>', '<':
		// One of these is the whole sequence. `<` is not, but the reference
		// consumes it here and parity means reproducing that.
		return 1, line, col

	case '(', ')':
		// Charset designation: one designator follows.
		if len(rest) > 1 {
			return 2, line, col
		}
		return 1, line, col
	}

	// An unrecognised escape is not consumed past the ESC: the reference leaves
	// the following character in the stream to be processed as ordinary text.
	return 0, line, col
}

// isCSIParameter reports whether a rune belongs to a CSI parameter string.
func isCSIParameter(r rune) bool {
	return (r >= '0' && r <= '9') || r == ';' || r == '?'
}

// applyCSI applies a control sequence and returns the resulting line and column.
func applyCSI(command rune, params string, line []rune, col int) ([]rune, int) {
	switch command {
	case 'K':
		// Erase in line.
		switch wholeParam(params, 0) {
		case 0:
			return line[:min(col, len(line))], col
		case 1:
			for i := 0; i <= col && i < len(line); i++ {
				line[i] = ' '
			}
			return line, col
		case 2:
			return line[:0], 0
		}
		return line, col

	case 'J':
		// Erase in display. In a line buffer only the current line exists.
		switch wholeParam(params, 0) {
		case 0:
			return line[:min(col, len(line))], col
		case 2:
			return line[:0], 0
		}
		return line, col

	case 'D':
		return line, max(0, col-atLeastOne(params))

	case 'C':
		return line, col + atLeastOne(params)

	case 'G', '`':
		return line, max(0, atLeastOne(params)-1)

	case 'H', 'f':
		// ESC [ row ; col H. Only the column has meaning in a line buffer.
		if _, column, found := strings.Cut(params, ";"); found {
			return line, max(0, atLeastOne(column)-1)
		}
		if params != "" {
			return line, max(0, atLeastOne(params)-1)
		}
		return line, col
	}

	// Colour and mode sequences ('m', '?2004h', …) change nothing here.
	return line, col
}

// wholeParam reads a parameter only when the entire string is a number.
//
// With a `;` or a `?` the reference's parse fails and it falls back, rather than
// taking the first number — so `ESC [ 1 ; 2 K` erases as little as `ESC [ K`.
func wholeParam(params string, fallback int) int {
	if params == "" {
		return fallback
	}
	value := 0
	for _, r := range params {
		if r < '0' || r > '9' {
			return fallback
		}
		value = value*10 + int(r-'0')
	}
	return value
}

// atLeastOne reads a parameter that defaults to one and is never smaller.
func atLeastOne(params string) int {
	return max(1, wholeParam(params, 1))
}
