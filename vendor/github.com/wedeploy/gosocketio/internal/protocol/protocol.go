package protocol

import (
	"errors"
)

// Message types.
const (
	// OnConnection for "connection" messages.
	OnConnection = "connection"

	// OnDisconnect for "disconnect" messages.
	OnDisconnect = "disconnect"

	// OnError for "error" messages.
	OnError = "error"

	MessageTypeOpen        = "0"
	MessageTypeClose       = "1"
	MessageTypePing        = "2"
	MessageTypePong        = "3"
	MessageTypeEmpty       = "empty"
	MessageTypeEmit        = "emit"
	MessageTypeAckRequest  = "ack_request"
	MessageTypeAckResponse = "ack_response"
	MessageTypeNamespace   = "namespace"
	MessageTypeError       = "error"
)

// Message to emit or receive
type Message struct {
	Namespace string
	Method    string

	Type  string
	AckID int

	Data   []byte
	Source string
}

const (
	// EmptyMessage code.
	EmptyMessage = "40"

	// NamespaceClose code.
	NamespaceClose = "41"

	// CommonMessage code.
	CommonMessage = "42"

	// AckMessage code.
	AckMessage = "43"

	// ErrorMessage code.
	ErrorMessage = "44"

	// OpenMessage is the opening message.
	OpenMessage = "0"

	// CloseMessage is the close signal.
	CloseMessage = "1"

	// PingMessage is the ping signal.
	PingMessage = "2"

	// PongMessage is the pong signal.
	PongMessage = "3"

	// RegularMessage is a regular message.
	RegularMessage = "4"
)

var (
	// ErrorWrongMessageType is used for wrong message type.
	ErrorWrongMessageType = errors.New("wrong message type")

	// ErrorWrongPacket is used for wrong packet.
	ErrorWrongPacket = errors.New("wrong packet")
)
