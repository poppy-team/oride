package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

// Errors a caller may need to distinguish.
var (
	ErrClosed    = errors.New("cliente encerrado")
	ErrNoProcess = errors.New("nenhum processo de servidor")
	// ErrNotReading means the outgoing queue is full: the server accepted the
	// connection and stopped reading it.
	ErrNotReading = errors.New("o servidor parou de ler")
)

// DiagnosticEvent is a batch of diagnostics for one document.
type DiagnosticEvent struct {
	URI         string
	Diagnostics []Diagnostic
}

// ServerMessage is a notification the client does not interpret.
type ServerMessage struct {
	Method string
	Params json.RawMessage
}

// Event is something the server said outside a response.
//
// Exactly one field is set. A tagged struct rather than an interface: the
// consumer is a single event loop, and a type switch over three shapes is
// clearer than three implementations of a one-method interface.
type Event struct {
	Diagnostics   *DiagnosticEvent
	ServerMessage *ServerMessage
	// Exited reports that the connection ended, cleanly or not.
	Exited bool
}

type response struct {
	result json.RawMessage
	err    error
}

// Client speaks to a language server.
//
// The transport is a pair of streams rather than a process, so a test can drive
// the whole client through a pipe with no server installed. The process, when
// there is one, is an implementation detail of Spawn.
type Client struct {
	in  io.Writer
	out io.Reader
	cmd *exec.Cmd

	mu      sync.Mutex
	nextID  uint64
	pending map[uint64]chan response
	closed  bool

	// outgoing is a bounded queue drained by a writer goroutine.
	//
	// Writing from the caller would block it for as long as the server refuses
	// to read, and no context deadline can interrupt a blocked write. A bounded
	// queue turns that into a prompt error instead of a hung editor.
	outgoing  chan any
	events    chan Event
	closeOnce sync.Once
	done      chan struct{}
}

// NewWithTransport builds a client over arbitrary streams.
func NewWithTransport(in io.Writer, out io.Reader) *Client {
	client := &Client{
		in:       in,
		out:      out,
		pending:  map[uint64]chan response{},
		outgoing: make(chan any, outgoingQueue),
		events:   make(chan Event, 64),
		done:     make(chan struct{}),
	}
	go client.writeLoop()
	go client.readLoop()
	return client
}

// outgoingQueue bounds how far ahead of the server the client may get.
const outgoingQueue = 64

// writeLoop drains the outgoing queue.
func (c *Client) writeLoop() {
	for {
		select {
		case message := <-c.outgoing:
			if err := WriteMessage(c.in, message); err != nil {
				c.fail(err)
				return
			}
		case <-c.done:
			return
		}
	}
}

// fail ends the client because the transport broke.
func (c *Client) fail(err error) {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.closed = true
		pending := c.pending
		c.pending = map[uint64]chan response{}
		c.mu.Unlock()

		for _, waiter := range pending {
			waiter <- response{err: err}
		}
		close(c.done)
	})
}

// Spawn starts a server process and speaks to it over its stdio.
//
// The command is passed as program plus arguments and never through a shell: the
// command comes from configuration, and configuration is not a place to hide a
// shell pipeline.
func Spawn(ctx context.Context, command []string, dir string) (*Client, error) {
	if len(command) == 0 {
		return nil, errors.New("comando do servidor vazio")
	}

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	if dir != "" {
		cmd.Dir = dir
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin do servidor: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout do servidor: %w", err)
	}
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("iniciando %s: %w", command[0], err)
	}

	client := NewWithTransport(stdin, stdout)
	client.cmd = cmd
	return client, nil
}

// Events returns the stream of server notifications.
func (c *Client) Events() <-chan Event { return c.events }

// Close stops the client and the process behind it.
func (c *Client) Close() error {
	var err error
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.closed = true
		pending := c.pending
		c.pending = map[uint64]chan response{}
		c.mu.Unlock()

		// Anything still waiting is told the client is gone, rather than left
		// blocked until its context expires.
		for _, waiter := range pending {
			waiter <- response{err: ErrClosed}
		}

		close(c.done)

		if closer, ok := c.in.(io.Closer); ok {
			_ = closer.Close()
		}
		if c.cmd != nil {
			if killErr := c.cmd.Process.Kill(); killErr != nil && !errors.Is(killErr, os.ErrProcessDone) {
				err = killErr
			}
			// Reaped so the process does not linger as a zombie, and so the
			// editor can exit cleanly with a server running.
			_ = c.cmd.Wait()
		}
	})
	return err
}

// readLoop dispatches messages until the stream ends.
func (c *Client) readLoop() {
	defer func() {
		select {
		case c.events <- Event{Exited: true}:
		default:
		}
	}()

	for {
		raw, err := ReadMessage(c.out)
		if err != nil {
			return
		}
		c.dispatch(raw)
	}
}

func (c *Client) dispatch(raw json.RawMessage) {
	var envelope struct {
		ID     *uint64         `json:"id"`
		Method string          `json:"method"`
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		Params json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return
	}

	// A message with an id is a response; one with a method is a notification.
	// A server may send a request of its own, which has both — it is answered
	// with an error rather than ignored, so the server is not left waiting.
	if envelope.ID != nil && envelope.Method != "" {
		c.sendErrorResponse(*envelope.ID, "o cliente não implementa requisições do servidor")
		return
	}

	if envelope.ID != nil {
		c.mu.Lock()
		waiter, ok := c.pending[*envelope.ID]
		delete(c.pending, *envelope.ID)
		c.mu.Unlock()

		if !ok {
			return
		}
		if envelope.Error != nil {
			waiter <- response{err: fmt.Errorf("erro do servidor %d: %s", envelope.Error.Code, envelope.Error.Message)}
			return
		}
		waiter <- response{result: envelope.Result}
		return
	}

	c.emitNotification(envelope.Method, envelope.Params)
}

func (c *Client) emitNotification(method string, params json.RawMessage) {
	event := Event{}

	switch method {
	case "textDocument/publishDiagnostics":
		var payload struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(params, &payload); err != nil {
			return
		}
		diagnostics, err := ParseDiagnostics(params)
		if err != nil {
			return
		}
		event.Diagnostics = &DiagnosticEvent{URI: payload.URI, Diagnostics: diagnostics}
	default:
		event.ServerMessage = &ServerMessage{Method: method, Params: params}
	}

	select {
	case c.events <- event:
	case <-c.done:
	default:
		// The consumer is behind. Dropping a notification is better than
		// blocking the reader, which would stall every pending request too.
	}
}

// send queues a message for the writer goroutine.
//
// The queue is bounded, so a server that stopped reading produces an error
// rather than a call that never returns.
func (c *Client) send(message any) error {
	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()
	if closed {
		return ErrClosed
	}
	if c.in == nil {
		return ErrNoProcess
	}

	select {
	case c.outgoing <- message:
		return nil
	case <-c.done:
		return ErrClosed
	default:
		return ErrNotReading
	}
}

func (c *Client) sendErrorResponse(id uint64, message string) {
	_ = c.send(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error":   map[string]any{"code": -32601, "message": message},
	})
}

// call sends a request and waits for its response.
func (c *Client) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, ErrClosed
	}
	c.nextID++
	id := c.nextID
	waiter := make(chan response, 1)
	c.pending[id] = waiter
	c.mu.Unlock()

	if err := c.send(Request(id, method, params)); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, err
	}

	select {
	case result := <-waiter:
		return result.result, result.err
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, fmt.Errorf("%s: %w", method, ctx.Err())
	case <-c.done:
		return nil, ErrClosed
	}
}

// ---------------------------------------------------------------------------
// Protocol methods
// ---------------------------------------------------------------------------

// Initialize performs the handshake.
func (c *Client) Initialize(ctx context.Context, rootURI string) error {
	params := map[string]any{
		"processId": nil,
		"rootUri":   rootURI,
		"capabilities": map[string]any{
			"textDocument": map[string]any{
				"completion":         map[string]any{"dynamicRegistration": false},
				"hover":              map[string]any{"dynamicRegistration": false},
				"definition":         map[string]any{"dynamicRegistration": false},
				"formatting":         map[string]any{"dynamicRegistration": false},
				"publishDiagnostics": map[string]any{},
				"synchronization": map[string]any{
					"dynamicRegistration": false,
					"didSave":             true,
				},
			},
		},
	}
	if _, err := c.call(ctx, "initialize", params); err != nil {
		return err
	}
	return c.send(Notification("initialized", map[string]any{}))
}

// Shutdown asks the server to stop and waits for it.
func (c *Client) Shutdown(ctx context.Context) error {
	_, err := c.call(ctx, "shutdown", nil)
	_ = c.send(Notification("exit", nil))
	return err
}

// DidOpen tells the server a document is open.
func (c *Client) DidOpen(path, languageID, text string) error {
	return c.send(Notification("textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{
			"uri":        PathToURI(path),
			"languageId": languageID,
			"version":    1,
			"text":       text,
		},
	}))
}

// DidChange reports the full text after an edit.
//
// Full text rather than an incremental range: the protocol allows either, and
// sending the whole document cannot desynchronise from the client's idea of what
// the server thinks it has.
func (c *Client) DidChange(path, text string, version int) error {
	return c.send(Notification("textDocument/didChange", map[string]any{
		"textDocument": map[string]any{
			"uri":     PathToURI(path),
			"version": version,
		},
		"contentChanges": []map[string]any{{"text": text}},
	}))
}

// DidSave reports that a document was written.
func (c *Client) DidSave(path, text string) error {
	params := map[string]any{
		"textDocument": map[string]any{"uri": PathToURI(path)},
	}
	if text != "" {
		params["text"] = text
	}
	return c.send(Notification("textDocument/didSave", params))
}

// DidClose reports that a document is no longer open.
func (c *Client) DidClose(path string) error {
	return c.send(Notification("textDocument/didClose", map[string]any{
		"textDocument": map[string]any{"uri": PathToURI(path)},
	}))
}

// positionParams builds a position request's parameters.
func positionParams(path string, line, character int) map[string]any {
	return map[string]any{
		"textDocument": map[string]any{"uri": PathToURI(path)},
		"position":     Position{Line: line, Character: character},
	}
}

// Completion asks for suggestions at a position.
func (c *Client) Completion(ctx context.Context, path string, line, character int) ([]CompletionItem, error) {
	raw, err := c.call(ctx, "textDocument/completion", positionParams(path, line, character))
	if err != nil {
		return nil, err
	}
	return ParseCompletion(raw)
}

// Hover asks for documentation at a position.
func (c *Client) Hover(ctx context.Context, path string, line, character int) (string, error) {
	raw, err := c.call(ctx, "textDocument/hover", positionParams(path, line, character))
	if err != nil {
		return "", err
	}
	return ParseHover(raw)
}

// Definition asks where a symbol is defined.
func (c *Client) Definition(ctx context.Context, path string, line, character int) ([]Location, error) {
	raw, err := c.call(ctx, "textDocument/definition", positionParams(path, line, character))
	if err != nil {
		return nil, err
	}
	return ParseLocations(raw)
}

// Formatting asks for edits that format the whole document.
func (c *Client) Formatting(ctx context.Context, path string, tabSize int, insertSpaces bool) ([]TextEdit, error) {
	raw, err := c.call(ctx, "textDocument/formatting", map[string]any{
		"textDocument": map[string]any{"uri": PathToURI(path)},
		"options": map[string]any{
			"tabSize":      tabSize,
			"insertSpaces": insertSpaces,
		},
	})
	if err != nil {
		return nil, err
	}
	var edits []TextEdit
	if err := json.Unmarshal(raw, &edits); err != nil {
		return nil, err
	}
	return edits, nil
}
