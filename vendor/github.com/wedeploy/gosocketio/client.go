package gosocketio

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/wedeploy/gosocketio/ack"
	"github.com/wedeploy/gosocketio/internal/protocol"
	"github.com/wedeploy/gosocketio/websocket"
)

const (
	// OnConnection for "connection" messages.
	OnConnection = protocol.OnConnection

	// OnDisconnect for "disconnect" messages.
	OnDisconnect = protocol.OnDisconnect

	// OnError for "error" messages.
	OnError = protocol.OnError

	// default namespace is always empty.
	defaultNamespace = ""
)

// Connect dials and waits for the "connection" event.
// It blocks for the timeout duration. If the connection is not established in time,
// it closes the connection and returns an error.
func Connect(u url.URL, tr *websocket.Transport) (c *Client, err error) {
	c, err = dial(u, tr)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), tr.PingTimeout)
	defer cancel()

	handshake := make(chan struct{}, 1)
	ec := make(chan error, 1)

	if err := c.On(OnConnection, func() {
		handshake <- struct{}{}
	}); err != nil {
		cancel()
		return nil, err
	}

	if err := c.On(OnError, func(err error) {
		ec <- err
	}); err != nil {
		cancel()
		return nil, err
	}

	select {
	case <-handshake:
		c.Off(OnConnection)
		c.Off(OnError)
	case e := <-ec:
		c = nil
		err = e
		c.Off(OnConnection)
		c.Off(OnError)
	case <-ctx.Done():
		c = nil
		err = fmt.Errorf("socket.io connection timeout (%v)", tr.PingTimeout)
	}

	cancel()
	return c, err
}

// DialOnly connects to the host and initializes the socket.io protocol.
// It doesn't wait for socket.io connection handshake.
// You probably want to use Connect instead. Only exposed for debugging.
func DialOnly(u url.URL, tr *websocket.Transport) (c *Client, err error) {
	return dial(u, tr)
}

func dial(u url.URL, tr *websocket.Transport) (c *Client, err error) {
	c = &Client{}
	c.init()

	var query = u.Query()
	query.Add("EIO", "3")
	query.Add("transport", "websocket")
	u.RawQuery = query.Encode()

	if !strings.HasSuffix(u.Path, "/socket.io") || !strings.HasSuffix(u.Path, "/socket.io/") {
		u.Path = u.Path + "/socket.io/"
	}

	c.connLocker.Lock()
	c.conn, err = tr.ConnectDialer(tr.Dialer, u.String())
	c.connLocker.Unlock()

	if err != nil {
		return nil, err
	}

	go c.inLoop()
	go c.outLoop()

	return c, nil
}

// Header of engine.io to send and receive packets
type Header struct {
	Sid          string   `json:"sid"`
	Upgrades     []string `json:"upgrades"`
	PingInterval int      `json:"pingInterval"`
	PingTimeout  int      `json:"pingTimeout"`
}

// Client to handle socket.io connections
type Client struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	header Header

	conn       *websocket.Connection
	connLocker sync.RWMutex

	namespaces       map[string]*Namespace
	namespacesLocker sync.RWMutex

	ack       *ack.Waiter
	ackLocker sync.RWMutex

	handlers       *handlers
	handlersLocker sync.RWMutex

	out chan *msgWriter
}

type handlers struct {
	m      map[location]*Handler
	locker sync.RWMutex
}

func (h *handlers) Get(l location) (*Handler, bool) {
	h.locker.RLock()
	value, ok := h.m[l]
	h.locker.RUnlock()

	return value, ok
}

func (h *handlers) Set(l location, handler *Handler) {
	h.locker.Lock()
	h.m[l] = handler
	h.locker.Unlock()
}

func (h *handlers) Delete(l location) {
	h.locker.Lock()
	delete(h.m, l)
	h.locker.Unlock()
}

func (h *handlers) Reset() {
	h.locker.Lock()
	h.m = map[location]*Handler{}
	h.locker.Unlock()
}

func (h *handlers) List(namespace string) (methods []string) {
	for handler := range h.m {
		if handler.namespace == namespace {
			methods = append(methods, handler.method)
		}
	}

	return methods
}

type location struct {
	namespace string
	method    string
}

func (c *Client) getConn() *websocket.Connection {
	c.connLocker.RLock()
	conn := c.conn
	c.connLocker.RUnlock()
	return conn
}

func (c *Client) getAck() *ack.Waiter {
	c.ackLocker.RLock()
	ack := c.ack
	c.ackLocker.RUnlock()
	return ack
}

func (c *Client) getHandlers() *handlers {
	c.handlersLocker.RLock()
	handlers := c.handlers
	c.handlersLocker.RUnlock()
	return handlers
}

func (c *Client) init() {
	c.ctx, c.ctxCancel = context.WithCancel(context.Background())
	c.namespaces = map[string]*Namespace{}
	c.ack = &ack.Waiter{}
	c.handlers = &handlers{}
	c.handlers.Reset()
	c.out = make(chan *msgWriter)
}

// ID of current socket connection
func (c *Client) ID() string {
	return c.header.Sid
}

// incoming messages loop, puts incoming messages to In channel
func (c *Client) inLoop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// gorilla's websocket (c *Conn) NextReader() is used internally by GetMessage
			// see notes there about breaking out of the loop on error
			pkg, err := c.getConn().GetMessage()

			if err == websocket.ErrUnsupportedBinaryMessage ||
				err == websocket.ErrBadBuffer ||
				err == websocket.ErrPacketType {
				c.callLoopEvent(defaultNamespace, protocol.OnError, err)
				continue
			}

			if err != nil {
				c.callLoopEvent(defaultNamespace, protocol.OnError, err)
				c.ctxCancel()
				return
			}

			msg, err := protocol.Decode(pkg)

			if err != nil {
				c.callLoopEvent(defaultNamespace, protocol.OnError, err)
				continue
			}

			if msg.Type == protocol.MessageTypeClose {
				return
			}

			c.incomingHandler(msg)
		}
	}
}

// outcoming messages loop
func (c *Client) outLoop() {
	// socket.io requires a ping strategy to identify that the connection is alive
	pingInterval, _ := c.getConn().PingParams()
	var ticker = time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case mw := <-c.out:
			writeMsg(mw, c.conn.WriteMessage)
		case <-ticker.C:
			if err := c.conn.WriteMessage(protocol.PingMessage); err != nil {
				c.callLoopEvent(defaultNamespace, OnError, err)
			}
		}
	}
}

func writeMsg(m *msgWriter, writer func(string) error) {
	defer m.wg.Done()
	m.err = writer(m.msg)
}

type msgWriter struct {
	msg string
	err error
	wg  sync.WaitGroup
}

func (c *Client) writeMessage(msg string) error {
	mw := &msgWriter{
		msg: msg,
	}

	mw.wg.Add(1)
	c.out <- mw
	mw.wg.Wait()
	return mw.err
}

// On registers a handler
func (c *Client) On(method string, f interface{}) error {
	def, err := c.Of(defaultNamespace)

	if err != nil {
		return err
	}

	return def.On(method, f)
}

// Off unregisters a listener
func (c *Client) Off(method string) {
	def, _ := c.Of(defaultNamespace)

	if def != nil {
		def.Off(method)
	}
}

// Listeners on the default namespace
func (c *Client) Listeners() (list []string) {
	def, _ := c.Of(defaultNamespace)

	if def != nil {
		list = def.Listeners()
	}

	return list
}

// Emit message
func (c *Client) Emit(method string, args ...interface{}) error {
	def, err := c.Of(defaultNamespace)

	if err != nil {
		return err
	}

	return def.Emit(method, args...)
}

// Ack packet based on given data and send it and receive response
func (c *Client) Ack(ctx context.Context, method string, args interface{}, ret interface{}) error {
	def, err := c.Of(defaultNamespace)

	if err != nil {
		return err
	}

	return def.Ack(ctx, method, args, ret)
}

// Of subscribes to a namespace
func (c *Client) Of(namespace string) (*Namespace, error) {
	c.namespacesLocker.RLock()
	n, ok := c.namespaces[namespace]
	c.namespacesLocker.RUnlock()

	if ok {
		return n, nil
	}

	if err := c.maybeOf(namespace); err != nil {
		return nil, err
	}

	n = NewNamespace(c, namespace)
	c.namespacesLocker.Lock()
	c.namespaces[namespace] = n
	c.namespacesLocker.Unlock()

	return n, nil
}

func (c *Client) maybeOf(namespace string) error {
	// no need to authenticate default namespace
	// see https://github.com/socketio/socket.io/issues/474
	if namespace == "" {
		return nil
	}

	msg := &protocol.Message{
		Type:   protocol.MessageTypeNamespace,
		Method: namespace,
	}

	command, err := protocol.Encode(msg)

	if err != nil {
		return err
	}

	return c.writeMessage(command)
}

// Close client connection
func (c *Client) Close() {
	if len(c.ctx.Done()) != 0 {
		return
	}

	c.getConn().Close()
	c.ctxCancel()
	c.callLoopEvent(defaultNamespace, protocol.OnDisconnect)
}

// Find message processing function associated with given method
func (c *Client) getHandler(namespace, method string) (*Handler, bool) {
	l := location{
		namespace: namespace,
		method:    method,
	}

	return c.handlers.Get(l)
}

func (c *Client) callLoopEvent(namespace string, event string, args ...interface{}) {
	h, ok := c.getHandler(namespace, event)

	if !ok {
		return // event handler not found
	}

	if event == OnError && len(args) == 1 {
		var e = args[0].(error)
		_ = h.Call(&e)
		return
	}

	if args != nil {
		_ = h.Call(args...)
		return
	}

	_ = h.Call(&struct{}{})
}

func (c *Client) incomingHandler(msg *protocol.Message) {
	switch msg.Type {
	case protocol.MessageTypeOpen:
		if err := jsonUnmarshalUnpanic([]byte(msg.Source[1:]), &c.header); err != nil {
			c.callLoopEvent(defaultNamespace, OnError, err)
			return
		}

		def, _ := c.Of(defaultNamespace)
		def.setReady()

		c.callLoopEvent(msg.Namespace, protocol.OnConnection)
	case protocol.MessageTypePing:
		if err := c.writeMessage(protocol.PongMessage); err != nil {
			c.callLoopEvent(defaultNamespace, OnError, err)
		}
	case protocol.MessageTypePong:
	case protocol.MessageTypeError:
		err := fmt.Errorf("error on method %s on namespace %s", msg.Method, msg.Namespace)
		c.callLoopEvent(msg.Namespace, protocol.OnError, err)
	case protocol.MessageTypeEmit:
		c.handleIncomingEmit(msg)
	case protocol.MessageTypeAckRequest:
		c.handleIncomingAckRequest(msg)
	case protocol.MessageTypeAckResponse:
		c.handleIncomingAckResponse(msg)
	case protocol.MessageTypeEmpty:
		if msg.Namespace != defaultNamespace {
			def, _ := c.Of(msg.Namespace)
			def.setReady()
			c.handleIncomingNamespaceConnection(msg)
		}
	default:
		err := fmt.Errorf("message type %s is not implemented", msg.Type)
		c.callLoopEvent(msg.Namespace, OnError, err)
	}
}

func (c *Client) handleIncomingEmit(msg *protocol.Message) {
	h, ok := c.getHandler(msg.Namespace, msg.Method)

	if !ok {
		return
	}

	var args, err = h.getFunctionCallArgs(msg)

	if err != nil {
		c.callLoopEvent(msg.Namespace, OnError, err)
		return
	}

	_ = h.Call(args...)
}

func (c *Client) handleIncomingAckRequest(msg *protocol.Message) {
	h, ok := c.getHandler(msg.Namespace, msg.Method)

	if !ok || !h.Out {
		return
	}

	var args, err = h.getFunctionCallArgs(msg)

	if err != nil {
		c.callLoopEvent(msg.Namespace, OnError, err)
		return
	}

	result := h.Call(args...)

	ack := &protocol.Message{
		Type:  protocol.MessageTypeAckResponse,
		AckID: msg.AckID,
	}

	def, _ := c.Of(defaultNamespace)

	var ri = []interface{}{}

	for _, r := range result {
		ri = append(ri, r.Interface())
	}

	if err = def.send(ack, ri...); err != nil {
		c.callLoopEvent(msg.Namespace, OnError, err)
		return
	}
}

func (c *Client) handleIncomingAckResponse(msg *protocol.Message) {
	ack := c.getAck()

	if waiter, ok := ack.Load(msg.AckID); ok {
		waiter <- string(msg.Data)
	}

	// couldn't find incoming ack
}

func (c *Client) handleIncomingNamespaceConnection(msg *protocol.Message) {
	if h, ok := c.getHandler(msg.Namespace, msg.Method); ok {
		_ = h.Call(nil)
	}
}
