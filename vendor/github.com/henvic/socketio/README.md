# golang socket.io [![GoDoc](https://godoc.org/github.com/henvic/socketio?status.svg)](https://godoc.org/github.com/henvic/socketio)

**This is a heavily modified fork of [github.com/graarh/golang-socketio](https://github.com/graarh/golang-socketio).**

The [socket.io](https://socket.io) protocol is only partially implemented.

See the example directory to see how you can use it. There is a lack of specification for the socket.io protocol, so reverse engineering is the easiest way to do it.

You must run the server.js code with node and the socket.io library to run the tests. Use `npm install socket.io` to install it.

If you need to improve this library you should consider using these tools:

* [Charles Proxy](https://www.charlesproxy.com)
* [Wireshark WebSocket wiki page](https://wiki.wireshark.org/WebSocket)

This library is used by the [WeDeploy](https://wedeploy.com) Command-Line Interface tool to open a shell and execute commands on Docker container services.