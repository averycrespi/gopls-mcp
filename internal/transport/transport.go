package transport

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/averycrespi/gopls-mcp/pkg/types"
)

const (
	receiveTimeout = 10 * time.Second
)

var _ types.Transport = &JsonRpcTransport{}

// JsonRpcTransport handles low-level JSON-RPC communication
type JsonRpcTransport struct {
	writer    io.Writer
	reader    io.Reader
	requestID int64
	responses map[int64]chan json.RawMessage
	mu        sync.RWMutex
	done      chan struct{}
}

// NewJsonRpcTransport creates a new JSON-RPC transport
func NewJsonRpcTransport(writer io.Writer, reader io.Reader) *JsonRpcTransport {
	return &JsonRpcTransport{
		writer:    writer,
		reader:    reader,
		responses: make(map[int64]chan json.RawMessage),
		done:      make(chan struct{}),
	}
}

func (t *JsonRpcTransport) Start() error {
	slog.Debug("Starting JSON-RPC transport")
	go t.readResponses()
	return nil
}

func (t *JsonRpcTransport) Stop() error {
	if !t.isClosed() {
		slog.Debug("Stopping JSON-RPC transport")
		close(t.done)
	}
	return nil
}

func (t *JsonRpcTransport) isClosed() bool {
	select {
	case <-t.done:
		return true
	default:
		return false
	}
}

func (t *JsonRpcTransport) readResponses() {
	slog.Debug("Reading JSON-RPC responses")

	defer func() {
		_ = t.Stop()
	}()

	for {
		// Read one response at a time until the transport is closed
		if t.isClosed() {
			return
		}

		var contentLength int
		var header []byte

		for {
			// Read one byte at a time until we find the end of the header
			b := make([]byte, 1)
			if _, err := t.reader.Read(b); err != nil {
				slog.Error("Failed to read JSON-RPC response header", "error", err)
				return
			}
			header = append(header, b[0])

			// Extract the Content-Length from the header, then break
			if len(header) >= 4 && string(header[len(header)-4:]) == "\r\n\r\n" {
				headerStr := string(header)
				if _, err := fmt.Sscanf(headerStr, "Content-Length: %d\r\n\r\n", &contentLength); err != nil {
					slog.Error("Failed to scan JSON-RPC response header", "error", err, "header", headerStr)
					continue
				}
				break
			}
		}

		// Use the Content-Length to read the JSON response body
		body := make([]byte, contentLength)
		if _, err := io.ReadFull(t.reader, body); err != nil {
			slog.Error("Failed to read JSON-RPC response body", "error", err, "content_length", contentLength)
			return
		}
		t.handleResponse(body)
	}
}

func (t *JsonRpcTransport) handleResponse(content []byte) {
	var resp struct {
		ID     json.RawMessage `json:"id"`
		Result json.RawMessage `json:"result"`
		Error  json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(content, &resp); err != nil {
		slog.Error("Failed to unmarshal JSON-RPC response", "error", err, "content", string(content))
		return
	}

	if resp.ID == nil {
		return // ignore notifications
	}

	var id int64
	if err := json.Unmarshal(resp.ID, &id); err != nil {
		slog.Error("Failed to unmarshal JSON-RPC response ID", "error", err, "raw_id", string(resp.ID))
		return
	}

	t.mu.RLock()
	ch, ok := t.responses[id]
	t.mu.RUnlock()

	if ok {
		if resp.Error != nil {
			ch <- resp.Error
		} else {
			ch <- resp.Result
		}
	}
}

// SendRequest sends a JSON-RPC request and waits for the response
func (t *JsonRpcTransport) SendRequest(method string, params any) (json.RawMessage, error) {
	if t.isClosed() {
		return nil, fmt.Errorf("cannot send request: transport is closed")
	}

	id := atomic.AddInt64(&t.requestID, 1)
	startTime := time.Now()

	slog.Debug("Sending JSON-RPC request", "request_id", id, "method", method)

	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}

	ch := make(chan json.RawMessage, 1)
	t.mu.Lock()
	t.responses[id] = ch
	t.mu.Unlock()

	defer func() {
		t.mu.Lock()
		delete(t.responses, id)
		t.mu.Unlock()
	}()

	if err := t.writeMessage(data); err != nil {
		return nil, fmt.Errorf("failed to write JSON-RPC request: %w", err)
	}

	select {
	case response := <-ch:
		duration := time.Since(startTime)
		slog.Debug("Received JSON-RPC response",
			"request_id", id,
			"method", method,
			"duration_ms", duration.Milliseconds())
		return response, nil
	case <-time.After(receiveTimeout):
		duration := time.Since(startTime)
		slog.Error("Timeout waiting for JSON-RPC response",
			"request_id", id,
			"method", method,
			"timeout_ms", receiveTimeout.Milliseconds(),
			"duration_ms", duration.Milliseconds())
		return nil, fmt.Errorf("timeout waiting for response to method %s", method)
	}
}

// SendNotification sends a JSON-RPC notification (no response expected)
func (t *JsonRpcTransport) SendNotification(method string, params any) error {
	if t.isClosed() {
		return fmt.Errorf("cannot send notification: transport is closed")
	}

	slog.Debug("Sending JSON-RPC notification", "method", method)

	notification := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON-RPC notification: %w", err)
	}

	if err := t.writeMessage(data); err != nil {
		return fmt.Errorf("failed to write JSON-RPC notification: %w", err)
	}

	return nil
}

func (t *JsonRpcTransport) writeMessage(data []byte) error {
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	if _, err := t.writer.Write([]byte(header)); err != nil {
		return fmt.Errorf("failed to write JSON-RPC message header: %w", err)
	}

	if _, err := t.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write JSON-RPC message data: %w", err)
	}

	return nil
}
