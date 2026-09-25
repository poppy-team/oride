package lsp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestUTF16ColumnAroundANonBMPCharacter is the reference's own example. A
// character outside the Basic Multilingual Plane occupies two UTF-16 units, and
// getting this wrong corrupts text at the first emoji.
func TestUTF16ColumnAroundANonBMPCharacter(t *testing.T) {
	line := "a😀b"

	if got := UTF16Column(line, 2); got != 3 {
		t.Errorf("UTF16Column(%q, 2) = %d, esperado 3 (a=1, 😀=2)", line, got)
	}
	if got, ok := CharacterColumn(line, 3); !ok || got != 2 {
		t.Errorf("CharacterColumn(%q, 3) = %d, %v; esperado 2, true", line, got, ok)
	}
	// Column 2 is inside the surrogate pair, and there is no scalar that
	// corresponds to it.
	if got, ok := CharacterColumn(line, 2); ok {
		t.Errorf("CharacterColumn(%q, 2) = %d, true; esperado false", line, got)
	}
}

func TestUTF16ConversionRoundTrips(t *testing.T) {
	lines := []string{
		"",
		"abc",
		"ção",
		"日本語",
		"a😀b",
		"🙂🙂🙂",
		"e\u0301",
		"fim🙂",
	}
	for _, line := range lines {
		scalars := 0
		for range line {
			scalars++
		}
		for column := 0; column <= scalars; column++ {
			units := UTF16Column(line, column)
			back, ok := CharacterColumn(line, units)
			if !ok {
				t.Errorf("%q: coluna %d → %d unidades → recusado", line, column, units)
				continue
			}
			if back != column {
				t.Errorf("%q: coluna %d → %d unidades → coluna %d", line, column, units, back)
			}
		}
	}
}

func TestFramingRoundTrip(t *testing.T) {
	var buffer bytes.Buffer
	message := map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize"}

	if err := WriteMessage(&buffer, message); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}
	if !strings.HasPrefix(buffer.String(), "Content-Length: ") {
		t.Errorf("cabeçalho ausente: %q", buffer.String())
	}

	raw, err := ReadMessage(&buffer)
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("corpo não é JSON: %v", err)
	}
	if decoded["method"] != "initialize" {
		t.Errorf("method = %v", decoded["method"])
	}
}

// TestFramingReadsConsecutiveMessages: a buffered reader would swallow the first
// bytes of the next message, so the reader must stop exactly at the body
// boundary.
func TestFramingReadsConsecutiveMessages(t *testing.T) {
	var buffer bytes.Buffer
	for i := 1; i <= 3; i++ {
		if err := WriteMessage(&buffer, map[string]any{"id": i}); err != nil {
			t.Fatalf("WriteMessage: %v", err)
		}
	}

	for i := 1; i <= 3; i++ {
		raw, err := ReadMessage(&buffer)
		if err != nil {
			t.Fatalf("mensagem %d: %v", i, err)
		}
		var decoded struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatalf("mensagem %d: %v", i, err)
		}
		if decoded.ID != i {
			t.Errorf("mensagem %d trouxe id %d", i, decoded.ID)
		}
	}
}

// TestContentLengthIsFoundWhateverTheSpelling: a header the client failed to
// recognise looks like a protocol violation rather than a parse gap.
func TestContentLengthIsFoundWhateverTheSpelling(t *testing.T) {
	for _, header := range []string{
		"Content-Length: 42\r\n\r\n",
		"content-length: 42\r\n\r\n",
		"CONTENT-LENGTH: 42\r\n\r\n",
		"Content-Type: application/vscode-jsonrpc; charset=utf-8\r\nContent-Length: 42\r\n\r\n",
	} {
		length, err := contentLength(header)
		if err != nil {
			t.Errorf("%q: %v", header, err)
			continue
		}
		if length != 42 {
			t.Errorf("%q → %d, esperado 42", header, length)
		}
	}
}

func TestFramingRejectsBrokenHeaders(t *testing.T) {
	if _, err := contentLength("Content-Type: text/plain\r\n\r\n"); !errors.Is(err, ErrMissingLength) {
		t.Errorf("erro = %v, esperado ErrMissingLength", err)
	}
	if _, err := contentLength("Content-Length: abc\r\n\r\n"); !errors.Is(err, ErrMalformedLength) {
		t.Errorf("erro = %v, esperado ErrMalformedLength", err)
	}
	if _, err := ReadMessage(strings.NewReader("")); !errors.Is(err, ErrEOF) {
		t.Errorf("erro = %v, esperado ErrEOF", err)
	}
}

// TestHeadersAreBounded: a server that never sends the blank line must not grow
// the buffer without limit.
func TestHeadersAreBounded(t *testing.T) {
	flood := strings.NewReader(strings.Repeat("X", maxHeaderBytes+16))
	if _, err := ReadMessage(flood); !errors.Is(err, ErrHeadersTooLarge) {
		t.Errorf("erro = %v, esperado ErrHeadersTooLarge", err)
	}
}

func TestPathToURIIsAFullFileURL(t *testing.T) {
	uri := PathToURI("/tmp/projeto/notas.txt")
	if uri != "file:///tmp/projeto/notas.txt" {
		t.Errorf("uri = %q", uri)
	}

	// A space must be encoded, or the URI is not a URI.
	withSpace := PathToURI("/tmp/meu projeto/a.txt")
	if strings.Contains(withSpace, " ") {
		t.Errorf("uri %q contém espaço não escapado", withSpace)
	}
}

func TestURIRoundTrip(t *testing.T) {
	path := "/tmp/projeto/notas.txt"
	back, err := URIToPath(PathToURI(path))
	if err != nil {
		t.Fatalf("URIToPath: %v", err)
	}
	if back != path {
		t.Errorf("volta = %q, esperado %q", back, path)
	}
}

// ---------------------------------------------------------------------------
// Stub server over a pipe
// ---------------------------------------------------------------------------

// stub drives a client without any process, by answering framed messages.
type stub struct {
	toClient   *io.PipeWriter
	fromClient *io.PipeReader
	client     *Client

	// mu guards seen. The stub records on its own goroutine while the test
	// reads, so an unguarded map is a data race — and a race that the detector
	// only catches under load is still a real defect, not a flaky test.
	mu   sync.Mutex
	seen map[string]int
}

// record counts one request, from the stub's goroutine.
func (s *stub) record(method string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen[method]++
}

// count reports how many times a method was requested.
func (s *stub) count(method string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.seen[method]
}

// newStub wires a client to a reader that answers with the given handler.
func newStub(t *testing.T, handler func(method string, params json.RawMessage) (any, bool)) *stub {
	t.Helper()

	serverFromClient, clientToServer := io.Pipe()
	clientFromServer, serverToClient := io.Pipe()

	client := NewWithTransport(clientToServer, clientFromServer)
	t.Cleanup(func() {
		_ = client.Close()
		_ = serverFromClient.Close()
		_ = serverToClient.Close()
	})

	server := &stub{toClient: serverToClient, fromClient: serverFromClient, client: client, seen: map[string]int{}}

	go func() {
		for {
			raw, err := ReadMessage(serverFromClient)
			if err != nil {
				return
			}
			var envelope struct {
				ID     *uint64         `json:"id"`
				Method string          `json:"method"`
				Params json.RawMessage `json:"params"`
			}
			if err := json.Unmarshal(raw, &envelope); err != nil {
				continue
			}
			server.record(envelope.Method)

			result, respond := handler(envelope.Method, envelope.Params)
			if !respond || envelope.ID == nil {
				continue
			}
			_ = WriteMessage(serverToClient, map[string]any{
				"jsonrpc": "2.0",
				"id":      *envelope.ID,
				"result":  result,
			})
		}
	}()

	return server
}

// TestClientHandshakeAndRequest drives the client end to end through a pipe.
func TestClientHandshakeAndRequest(t *testing.T) {
	server := newStub(t, func(method string, params json.RawMessage) (any, bool) {
		switch method {
		case "initialize":
			return map[string]any{"capabilities": map[string]any{}}, true
		case "textDocument/completion":
			return map[string]any{"items": []map[string]any{
				{"label": "println!", "detail": "macro"},
				{"label": "print", "detail": "fn"},
			}}, true
		case "textDocument/hover":
			return map[string]any{"contents": map[string]any{"kind": "markdown", "value": "```rust\nfn\n```"}}, true
		case "textDocument/definition":
			return []map[string]any{{
				"uri":   "file:///tmp/projeto/lib.rs",
				"range": map[string]any{"start": map[string]any{"line": 3, "character": 0}},
			}}, true
		case "shutdown":
			return nil, true
		}
		return nil, false
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.client.Initialize(ctx, "file:///tmp/projeto"); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	items, err := server.client.Completion(ctx, "/tmp/projeto/main.rs", 2, 4)
	if err != nil {
		t.Fatalf("Completion: %v", err)
	}
	if len(items) != 2 || items[0].Label != "println!" {
		t.Errorf("items = %+v", items)
	}

	hover, err := server.client.Hover(ctx, "/tmp/projeto/main.rs", 2, 4)
	if err != nil {
		t.Fatalf("Hover: %v", err)
	}
	if !strings.Contains(hover, "fn") {
		t.Errorf("hover = %q", hover)
	}

	locations, err := server.client.Definition(ctx, "/tmp/projeto/main.rs", 2, 4)
	if err != nil {
		t.Fatalf("Definition: %v", err)
	}
	if len(locations) != 1 || locations[0].Range.Start.Line != 3 {
		t.Errorf("locations = %+v", locations)
	}

	if err := server.client.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	if got := server.count("initialize"); got != 1 {
		t.Errorf("initialize chamado %d vezes", got)
	}
}

// TestDiagnosticsArriveAsEvents checks the notification path, which is how a
// server reports problems without being asked.
func TestDiagnosticsArriveAsEvents(t *testing.T) {
	server := newStub(t, func(method string, params json.RawMessage) (any, bool) {
		if method == "initialize" {
			return map[string]any{"capabilities": map[string]any{}}, true
		}
		return nil, false
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.client.Initialize(ctx, "file:///tmp"); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	notification := map[string]any{
		"jsonrpc": "2.0",
		"method":  "textDocument/publishDiagnostics",
		"params": map[string]any{
			"uri": "file:///tmp/projeto/main.rs",
			"diagnostics": []map[string]any{
				{"range": map[string]any{}, "severity": 1, "message": "não compila"},
				{"range": map[string]any{}, "message": "sem severidade"},
			},
		},
	}
	if err := WriteMessage(server.toClient, notification); err != nil {
		t.Fatalf("escrevendo notificação: %v", err)
	}

	select {
	case event := <-server.client.Events():
		if event.Diagnostics == nil {
			t.Fatalf("evento sem diagnósticos: %+v", event)
		}
		if len(event.Diagnostics.Diagnostics) != 2 {
			t.Fatalf("diagnósticos = %d", len(event.Diagnostics.Diagnostics))
		}
		if event.Diagnostics.Diagnostics[0].Severity != SeverityError {
			t.Errorf("severidade = %d, esperado Error", event.Diagnostics.Diagnostics[0].Severity)
		}
		// A diagnostic without a severity is still a problem; dropping it would
		// hide a real finding.
		if event.Diagnostics.Diagnostics[1].Severity == 0 {
			t.Error("diagnóstico sem severidade ficou sem classificação")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("nenhum evento de diagnóstico")
	}
}

// TestCloseUnblocksPendingRequests: a request waiting when the client closes
// must be told, not left until its context expires.
func TestCloseUnblocksPendingRequests(t *testing.T) {
	// A stub that never answers.
	_, clientToServer := io.Pipe()
	clientFromServer, _ := io.Pipe()

	client := NewWithTransport(clientToServer, clientFromServer)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := client.Completion(ctx, "/tmp/a.rs", 0, 0)
		done <- err
	}()

	// Wait for the request to actually be registered rather than sleeping for a
	// guessed duration: a fixed sleep is what makes a test flaky under load.
	waitForPending(t, client, 1)

	if err := client.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	select {
	case err := <-done:
		if err == nil {
			t.Error("a requisição pendente foi concluída sem erro após o fechamento")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("a requisição pendente não foi desbloqueada pelo fechamento")
	}
}

// waitForPending blocks until the client has the expected number of requests in
// flight, or fails.
func waitForPending(t *testing.T, client *Client, want int) {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		client.mu.Lock()
		count := len(client.pending)
		client.mu.Unlock()
		if count >= want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("a requisição não foi registrada: esperado %d pendente(s)", want)
}

func TestCloseIsIdempotent(t *testing.T) {
	_, clientToServer := io.Pipe()
	clientFromServer, _ := io.Pipe()

	client := NewWithTransport(clientToServer, clientFromServer)
	for range 3 {
		if err := client.Close(); err != nil {
			t.Fatalf("Close repetido: %v", err)
		}
	}
}

// TestRequestAfterCloseFailsClosed: a closed client must refuse rather than
// write into a closed pipe.
func TestRequestAfterCloseFailsClosed(t *testing.T) {
	_, clientToServer := io.Pipe()
	clientFromServer, _ := io.Pipe()

	client := NewWithTransport(clientToServer, clientFromServer)
	if err := client.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := client.Hover(ctx, "/tmp/a.rs", 0, 0); !errors.Is(err, ErrClosed) {
		t.Errorf("erro = %v, esperado ErrClosed", err)
	}
}

// TestRequestTimeout: a server that reads but never answers must not hang the
// editor.
func TestRequestTimeout(t *testing.T) {
	// The stub consumes whatever it is sent and answers nothing.
	server := newStub(t, func(string, json.RawMessage) (any, bool) { return nil, false })

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	if _, err := server.client.Hover(ctx, "/tmp/a.rs", 0, 0); err == nil {
		t.Error("uma requisição sem resposta não expirou")
	}
}

// TestServerThatStopsReadingFailsPromptly: writing from the caller would block
// for as long as the server refuses to read, and no context deadline interrupts a
// blocked write. The bounded queue turns that into an error.
func TestServerThatStopsReadingFailsPromptly(t *testing.T) {
	// A pipe whose reader side nobody drains.
	_, clientToServer := io.Pipe()
	clientFromServer, serverToClient := io.Pipe()

	client := NewWithTransport(clientToServer, clientFromServer)
	t.Cleanup(func() {
		_ = client.Close()
		_ = serverToClient.Close()
	})

	// The queue is driven directly: `call` waits for a response, so it would
	// never get far enough to fill anything.
	var lastErr error
	for i := range outgoingQueue + 8 {
		if err := client.send(Notification("probe", map[string]any{"n": i})); err != nil {
			lastErr = err
			break
		}
	}
	if lastErr == nil {
		t.Fatal("o cliente aceitou mensagens indefinidamente sem o servidor ler")
	}
	if !errors.Is(lastErr, ErrNotReading) && !errors.Is(lastErr, ErrClosed) {
		t.Errorf("erro = %v, esperado ErrNotReading ou ErrClosed", lastErr)
	}
}

func TestParseCompletionAcceptsBothShapes(t *testing.T) {
	bare := json.RawMessage(`[{"label":"a"},{"label":"b"}]`)
	items, err := ParseCompletion(bare)
	if err != nil || len(items) != 2 {
		t.Errorf("lista crua: %v, %+v", err, items)
	}

	wrapped := json.RawMessage(`{"isIncomplete":false,"items":[{"label":"c"}]}`)
	items, err = ParseCompletion(wrapped)
	if err != nil || len(items) != 1 || items[0].Label != "c" {
		t.Errorf("objeto com items: %v, %+v", err, items)
	}
}

func TestParseHoverAcceptsEveryShape(t *testing.T) {
	cases := map[string]string{
		`{"contents":{"kind":"markdown","value":"doc"}}`: "doc",
		`{"contents":"texto simples"}`:                   "texto simples",
		`{"contents":[{"value":"um"},{"value":"dois"}]}`: "um\ndois",
		`null`: "",
	}
	for raw, want := range cases {
		got, err := ParseHover(json.RawMessage(raw))
		if err != nil {
			t.Errorf("%s: %v", raw, err)
			continue
		}
		if got != want {
			t.Errorf("%s → %q, esperado %q", raw, got, want)
		}
	}
}

func TestParseLocationsAcceptsBothShapes(t *testing.T) {
	single, err := ParseLocations(json.RawMessage(`{"uri":"file:///a.rs","range":{}}`))
	if err != nil || len(single) != 1 {
		t.Errorf("única: %v, %+v", err, single)
	}

	list, err := ParseLocations(json.RawMessage(`[{"uri":"file:///a.rs","range":{}},{"uri":"file:///b.rs","range":{}}]`))
	if err != nil || len(list) != 2 {
		t.Errorf("lista: %v, %+v", err, list)
	}

	empty, err := ParseLocations(json.RawMessage(`null`))
	if err != nil || len(empty) != 0 {
		t.Errorf("nulo: %v, %+v", err, empty)
	}
}
