// Package lsp speaks the Language Server Protocol over stdio.
//
// A language server is a third-party process that may be missing, slow, or
// crash. None of that may take the editor down, so every failure here becomes a
// value the caller turns into a status line.
package lsp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
)

// Protocol errors.
var (
	ErrEOF             = errors.New("fim da conexão antes dos cabeçalhos")
	ErrHeadersTooLarge = errors.New("cabeçalhos maiores que o limite")
	ErrMissingLength   = errors.New("cabeçalho Content-Length ausente")
	ErrMalformedLength = errors.New("Content-Length não é um número")
)

// maxHeaderBytes bounds the header section. A server that never sends the blank
// line would otherwise grow the buffer without limit.
const maxHeaderBytes = 64 * 1024

// WriteMessage writes one JSON-RPC message with Content-Length framing.
func WriteMessage(w io.Writer, body any) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("serializando mensagem: %w", err)
	}
	if _, err := fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(encoded)); err != nil {
		return fmt.Errorf("escrevendo cabeçalho: %w", err)
	}
	if _, err := w.Write(encoded); err != nil {
		return fmt.Errorf("escrevendo corpo: %w", err)
	}
	return nil
}

// ReadMessage reads one framed message.
//
// Read byte by byte until the blank line rather than scanning with a buffered
// reader, because the body follows immediately and a buffered reader would
// swallow the first bytes of it.
func ReadMessage(r io.Reader) (json.RawMessage, error) {
	header := make([]byte, 0, 128)
	buffer := make([]byte, 1)

	for {
		n, err := r.Read(buffer)
		if n == 0 {
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("lendo cabeçalho: %w", err)
			}
			return nil, ErrEOF
		}
		header = append(header, buffer[0])

		if len(header) >= 4 && string(header[len(header)-4:]) == "\r\n\r\n" {
			break
		}
		if len(header) > maxHeaderBytes {
			return nil, ErrHeadersTooLarge
		}
	}

	length, err := contentLength(string(header))
	if err != nil {
		return nil, err
	}

	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, fmt.Errorf("lendo corpo: %w", err)
	}
	return json.RawMessage(body), nil
}

// contentLength finds the length in a header block.
//
// Scanned case-insensitively instead of assuming the exact spelling: servers
// differ, and a header the client failed to recognise looks like a protocol
// violation rather than a parse gap.
func contentLength(header string) (int, error) {
	for _, line := range strings.Split(header, "\r\n") {
		name, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			continue
		}
		var length int
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &length); err != nil {
			return 0, fmt.Errorf("%w: %q", ErrMalformedLength, value)
		}
		if length < 0 {
			return 0, fmt.Errorf("%w: %d", ErrMalformedLength, length)
		}
		return length, nil
	}
	return 0, ErrMissingLength
}

// Request builds a request envelope.
func Request(id uint64, method string, params any) map[string]any {
	return map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}
}

// Notification builds a notification envelope.
func Notification(method string, params any) map[string]any {
	return map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
}

// PathToURI renders a filesystem path as a file URI.
//
// A POSIX absolute path needs only the scheme prefix, because it already begins
// with a slash. A Windows path does not, and is routed through the URL encoder so
// that `C:\dir\file` becomes `file:///C:/dir/file` rather than the invalid
// `file://C:\dir\file` the reference produces — a defect with no measured effect
// yet, recorded in the parity ledger.
func PathToURI(path string) string {
	absolute := path
	if resolved, err := filepath.Abs(path); err == nil {
		absolute = resolved
	}
	slash := filepath.ToSlash(absolute)

	if strings.HasPrefix(slash, "/") {
		return "file://" + escapePath(slash)
	}
	return (&url.URL{Scheme: "file", Path: "/" + slash}).String()
}

// escapePath percent-encodes the characters that are not valid in a URI path,
// leaving the separators alone.
func escapePath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

// URIToPath converts a file URI back to a filesystem path.
func URIToPath(uri string) (string, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("uri inválida %q: %w", uri, err)
	}
	if parsed.Scheme != "file" {
		return "", fmt.Errorf("uri não é file: %q", uri)
	}
	return filepath.FromSlash(parsed.Path), nil
}
