package types

import "encoding/json"

// Transport defines transport layer interface
type Transport interface {
	Listen()
	IsClosed() bool
	Close()
	SendRequest(method string, params any) (json.RawMessage, error)
	SendNotification(method string, params any) error
}
