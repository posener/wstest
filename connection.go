package wstest

import (
	"net"
	"time"
)

// conn is a connection for testing, implementing the net.Conn interface
type conn struct {
	in     <-chan []byte
	out    chan<- []byte
	local  net.Addr
	remote net.Addr
}

// newConnPair returns two connections, paired by channels.
// any message written into the first will be read in the second
// and vice-versa.
func newConnPair() (server, client net.Conn) {
	var (
		s2c   = make(chan []byte)
		c2s   = make(chan []byte)
		cAddr = &address{"tcp", "127.0.0.1:12345"}
		sAddr = &address{"tcp", "8.8.8.8:12346"}
	)

	server = &conn{in: c2s, out: s2c, local: sAddr, remote: cAddr}
	client = &conn{in: s2c, out: c2s, local: cAddr, remote: sAddr}
	return
}

func (c *conn) Read(b []byte) (n int, err error) {
	read := <-c.in
	n = copy(b, read)
	return
}

func (c *conn) Write(b []byte) (n int, err error) {
	c.out <- b
	return len(b), nil
}

func (c *conn) Close() error {
	close(c.out)
	return nil
}

func (c *conn) LocalAddr() net.Addr { return c.local }

func (c *conn) RemoteAddr() net.Addr { return c.remote }

func (c *conn) SetDeadline(t time.Time) error { return nil }

func (c *conn) SetReadDeadline(t time.Time) error { return nil }

func (c *conn) SetWriteDeadline(t time.Time) error { return nil }

type address struct {
	network string
	address string
}

func (a *address) Network() string { return a.network }

func (a *address) String() string { return a.address }
