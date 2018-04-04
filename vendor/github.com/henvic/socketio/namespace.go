package socketio

import (
	"context"
	"encoding/json"

	"github.com/henvic/socketio/ack"
	"github.com/henvic/socketio/internal/protocol"
)

// NewNamespace creates a namespace
func NewNamespace(c *Client, namespace string) *Namespace {
	return &Namespace{
		name: namespace,

		getHandlers:  c.getHandlers,
		getAck:       c.getAck,
		writeMessage: c.writeMessage,
	}
}

// Namespace for the connection
type Namespace struct {
	name string

	getHandlers  func() *handlers
	getAck       func() *ack.Waiter
	writeMessage func(message string) error
}

// On registers a listener
func (n *Namespace) On(method string, f interface{}) error {
	h, err := NewHandler(f)

	if err != nil {
		return err
	}

	l := location{
		namespace: n.name,
		method:    method,
	}

	n.getHandlers().Set(l, h)
	return nil
}

// Off unregisters a listener
func (n *Namespace) Off(method string) {
	l := location{
		namespace: n.name,
		method:    method,
	}

	n.getHandlers().Delete(l)
}

// Listeners on a namespace
func (n *Namespace) Listeners() (list []string) {
	handlers := n.getHandlers()

	return handlers.List(n.name)
}

// Emit message
func (n *Namespace) Emit(method string, args ...interface{}) error {
	msg := &protocol.Message{
		Type:   protocol.MessageTypeEmit,
		Method: method,
	}

	return n.send(msg, args...)
}

// Ack packet based on given data and send it and receive response
func (n *Namespace) Ack(ctx context.Context, method string, args interface{}, v interface{}) error {
	msg := &protocol.Message{
		Type:   protocol.MessageTypeAckRequest,
		AckID:  n.getAck().Next(),
		Method: method,
	}

	waiter := make(chan string)
	n.getAck().Set(msg.AckID, waiter)

	if err := n.send(msg, args); err != nil {
		n.getAck().Delete(msg.AckID)
	}

	select {
	case ret := <-waiter:
		ret = ret[1 : len(ret)-1]
		return json.Unmarshal([]byte(ret), v)
	case <-ctx.Done():
		n.getAck().Delete(msg.AckID)
		return ctx.Err()
	}
}

func (n *Namespace) send(msg *protocol.Message, args ...interface{}) (err error) {
	msg.Namespace = n.name
	command, err := protocol.Encode(msg, args...)

	if err != nil {
		return err
	}

	return n.writeMessage(command)
}
