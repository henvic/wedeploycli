package websocket

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	ws "github.com/gorilla/websocket"
)

const (
	// PingInterval for the connection
	PingInterval = 25 * time.Second

	// PingTimeout for the connection
	PingTimeout = 60 * time.Second

	// ReadTimeout for the connection
	ReadTimeout = 60 * time.Second

	// SendTimeout for the connection
	SendTimeout = 60 * time.Second

	// BufferSize for the connection
	BufferSize = 1024 * 32
)

var (
	// ErrUnsupportedBinaryMessage is returned when trying to send an unsupported binary message
	ErrUnsupportedBinaryMessage = errors.New("sending binary messages is not supported")

	// ErrBadBuffer is used when there is an error while reading the buffer
	ErrBadBuffer = errors.New("error while reading buffer")

	// ErrPacketType is used when a packet comes with an unexpected format
	ErrPacketType = errors.New("wrong packet type")
)

// Connection to websocket
type Connection struct {
	socket    *ws.Conn
	transport *Transport
}

// GetMessage on connection
func (c *Connection) GetMessage() (data []byte, err error) {
	c.socket.SetReadDeadline(time.Now().Add(c.transport.ReadTimeout))

	msgType, reader, err := c.socket.NextReader()

	if err != nil {
		return data, err
	}

	if msgType != ws.TextMessage {
		return data, ErrUnsupportedBinaryMessage
	}

	data, err = ioutil.ReadAll(reader)

	if err != nil {
		return data, ErrBadBuffer
	}

	if len(data) == 0 {
		return data, ErrPacketType
	}

	return data, nil
}

// WriteMessage to the socket
func (c *Connection) WriteMessage(message string) error {
	c.socket.SetWriteDeadline(time.Now().Add(c.transport.SendTimeout))
	writer, err := c.socket.NextWriter(ws.TextMessage)

	if err != nil {
		return err
	}

	if _, err := writer.Write([]byte(message)); err != nil {
		return err
	}

	return writer.Close()
}

// Close the connection
func (c *Connection) Close() {
	c.socket.Close()
}

// PingParams gets the ping and pong interval and timeout
func (c *Connection) PingParams() (interval, timeout time.Duration) {
	return c.transport.PingInterval, c.transport.PingTimeout
}

// Transport for the websocket
type Transport struct {
	PingInterval time.Duration
	PingTimeout  time.Duration
	ReadTimeout  time.Duration
	SendTimeout  time.Duration

	BufferSize int

	Dialer        ws.Dialer
	RequestHeader http.Header
}

// NewTransport creates a new WebSocket connection transport
func NewTransport() *Transport {
	t := &Transport{
		PingInterval:  PingInterval,
		PingTimeout:   PingTimeout,
		ReadTimeout:   ReadTimeout,
		SendTimeout:   SendTimeout,
		BufferSize:    BufferSize,
		RequestHeader: http.Header{},
	}

	t.RequestHeader.Add("User-Agent", "socketio client; (+https://github.com/wedeploy/gosocket.io)")

	return t
}

// Connect to web socket with default gorilla websocket dialer
func (wst *Transport) Connect(url string) (conn *Connection, err error) {
	return wst.ConnectDialer(*ws.DefaultDialer, url)
}

// ConnectDialer to web socket
func (wst *Transport) ConnectDialer(dialer ws.Dialer, url string) (conn *Connection, err error) {
	socket, _, err := dialer.Dial(url, wst.RequestHeader)

	if err != nil {
		return nil, err
	}

	return &Connection{socket, wst}, nil
}
