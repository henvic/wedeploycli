# golang socket.io [<img src="https://avatars3.githubusercontent.com/u/10002920" alt="WeDeploy logo" width="90" height="90" align="right">](https://wedeploy.com/)

[![GoDoc](https://godoc.org/github.com/wedeploy/gosocketio?status.svg)](https://godoc.org/github.com/wedeploy/gosocketio) [![Build Status](https://travis-ci.org/wedeploy/gosocketio.svg?branch=master)](https://travis-ci.org/wedeploy/gosocketio)

golang socket.io is an implementation for the [socket.io](https://socket.io) protocol in Go. There is a lack of specification for the socket.io protocol, so reverse engineering is the easiest way to find out how it works.

**This is a work in progress. Many features, such as middleware and binary support, are missing.**

**golang socket.io is an adapted work from [github.com/graarh/golang-socketio](https://github.com/graarh/golang-socketio).**

---

## on "connection", "error", and "disconnection"
socket.io has three special events it triggers on the client-side and you should not emit them on your own programs.

**Wait for the socket.io connection event before emitting messages or you risk losing them** due in an unpredictable fashion (due to concurrency: connection latency, server load, etc.). For the default namespace this is automatically handled on gosocketio.Connect.

However, before emitting a message on a custom namespace, you want to wait for the ready signal, like so:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
defer cancel()

select {
	case <-ctx.Done():
		return ctx.Err()
	case <-exampleNamespace.Ready():
		// don't need to do anything
}

if err := exampleNamespace.Emit("fleet", 100); err != nil {
	return err
}
```

The reason why you probably want to use a `select` receiving a second channel, such as context.Done() on all non-trivial programs is to avoid program loop, leak memory, or both in case of failure.

The default namespace is automatically ready after establishing the socket.io session. Therefore, `*gosocketio.Client` doesn't expose a `Ready()` method.

## Connecting to a socket.io server with a custom namespace
You can connect to a namespace and start emitting messages to it with:

```go
c, err := gosocketio.Connect(u, websocket.NewTransport())

if err != nil {
	return err
}

// ...

exampleNamespace, err := c.Of("example")

If err != nil {
	return err
}

<-exampleNamespace.Ready() // don't do this, use a select like shown above instead!

If err := exampleNamespace.Emit("list", "friends"); err != nil {
	return err
}
```

## Running the example

1. `npm install` to install the dependencies for the example server
2. `node server.js`
2. `go run example.go`

If you need to improve this library you should consider using these tools:

* [Charles Proxy](https://www.charlesproxy.com)
* [Wireshark WebSocket wiki page](https://wiki.wireshark.org/WebSocket)

This library is used by the [WeDeploy](https://wedeploy.com) Command-Line Interface tool to open a shell and execute commands on Docker container services.
