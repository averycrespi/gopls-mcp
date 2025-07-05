package types

import "encoding/json"

// Transport defines transport layer interface
type Transport interface {
	Start() error
	Stop() error

	SendRequest(method string, params any) (json.RawMessage, error)
	SendNotification(method string, params any) error
}
