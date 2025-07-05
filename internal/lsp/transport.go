package lsp

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// Transport handles the low-level JSON-RPC communication over stdin/stdout
type Transport struct {
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	requestID int64
	responses map[int64]chan json.RawMessage
	mu        sync.RWMutex
	done      chan struct{}
}

// NewTransport creates a new transport instance
func NewTransport(stdin io.WriteCloser, stdout io.ReadCloser) *Transport {
	return &Transport{
		stdin:     stdin,
		stdout:    stdout,
		responses: make(map[int64]chan json.RawMessage),
		done:      make(chan struct{}),
	}
}

// Start starts reading responses from the transport
func (t *Transport) Start() {
	go t.readResponses()
}

// Close closes the transport
func (t *Transport) Close() error {
	close(t.done)
	if t.stdin != nil {
		return t.stdin.Close()
	}
	return nil
}

// lspResponse represents a JSON-RPC response
type lspResponse struct {
	ID     json.RawMessage `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  json.RawMessage `json:"error"`
}

// readResponses reads responses from the stdout stream
func (t *Transport) readResponses() {
	defer close(t.done)

	for {
		// Read Content-Length header byte by byte until we find \r\n\r\n
		var contentLength int
		var header []byte

		for {
			b := make([]byte, 1)
			if _, err := t.stdout.Read(b); err != nil {
				return
			}
			header = append(header, b[0])

			if len(header) >= 4 && string(header[len(header)-4:]) == "\r\n\r\n" {
				headerStr := string(header)
				if _, err := fmt.Sscanf(headerStr, "Content-Length: %d\r\n\r\n", &contentLength); err != nil {
					continue
				}
				break
			}
		}

		// Read the JSON response body
		body := make([]byte, contentLength)
		if _, err := io.ReadFull(t.stdout, body); err != nil {
			return
		}

		t.handleResponse(body)
	}
}

// handleResponse handles a JSON-RPC response
func (t *Transport) handleResponse(content []byte) {
	var resp lspResponse
	if err := json.Unmarshal(content, &resp); err != nil {
		return
	}

	if resp.ID == nil {
		return // notification
	}

	var id int64
	if err := json.Unmarshal(resp.ID, &id); err != nil {
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
func (t *Transport) SendRequest(method string, params any) (json.RawMessage, error) {
	id := atomic.AddInt64(&t.requestID, 1)

	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
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
		return nil, err
	}

	// Wait for response with timeout
	select {
	case response := <-ch:
		return response, nil
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response to method %s", method)
	}
}

// SendNotification sends a JSON-RPC notification (no response expected)
func (t *Transport) SendNotification(method string, params any) error {
	notification := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	return t.writeMessage(data)
}

// writeMessage writes a message with the LSP Content-Length header
func (t *Transport) writeMessage(data []byte) error {
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	if _, err := t.stdin.Write([]byte(header)); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	if _, err := t.stdin.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}