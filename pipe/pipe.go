// Package pipe provides a function that creates two paired in-memory
// net connections: objects that implements the `net.Conn` interface.
//
// The standard library `net.Pipe() (c1, c2 net.Conn)` function
// creates two in-memory paired connections, which will return a
// not-implemented error up on SetDeadline, SetReadDeadline or SetWriteDeadline
// calls. This functionality is sometimes needed, and it is provided by
// this package pipe implementation.
package pipe

import "net"

// New creates a new pipe.
// log is a log.Println-like function that can be set for debugging the
// connection. It can be set to nil if no debug is needed.
func New(log Println) (c1, c2 net.Conn) {
	var (
		// an in-memory buffer that holds server to client communication
		s2c = newBuffer()
		// an in-memory buffer that holds client to server communication
		c2s   = newBuffer()
		cAddr = &address{"tcp", "1.2.3.4:12345"}
		sAddr = &address{"tcp", "5.6.7.8:12346"}
	)

	// cross the s2c and c2s buffers between two connections, so everything written to the
	// outgoing buffer of the first could be read by the ingoing buffer of the second,
	// and vice-versa.
	c1 = &conn{name: "client", in: s2c, out: c2s, local: cAddr, remote: sAddr, log: log}
	c2 = &conn{name: "server", in: c2s, out: s2c, local: sAddr, remote: cAddr, log: log}

	return
}
