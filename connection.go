package wstest

import (
	"io"
	"net"
	"time"
)

// conn is a connection for testing, implementing the net.Conn interface
type conn struct {
	name   string
	in     *buffer
	out    *buffer
	local  net.Addr
	remote net.Addr

	// set Log to a Println function in order to print debug information of the connection
	Log func(v ...interface{})
}

// newConnPair returns two connections, paired by channels.
// any message written into the first will be read in the second
// and vice-versa.
func newConnPair() (server, client *conn) {
	var (
		s2c   = newBuffer()
		c2s   = newBuffer()
		cAddr = &address{"tcp", "127.0.0.1:12345"}
		sAddr = &address{"tcp", "8.8.8.8:12346"}
	)

	server = &conn{name: "server", in: c2s, out: s2c, local: sAddr, remote: cAddr}
	client = &conn{name: "client", in: s2c, out: c2s, local: cAddr, remote: sAddr}
	return
}

func (c *conn) Read(b []byte) (n int, err error) {
	c.in.Lock()
	defer c.in.Unlock()
	for {
		n, err = c.in.Read(b)
		if err != io.EOF {
			break
		}

		// nothing to read, wait for a signal from the other side writer
		if c.Log != nil {
			c.Log(c.name, "waiting read")
		}
		c.in.Wait()
	}
	if c.Log != nil {
		c.Log(c.name, err, "<", string(b[:n]))
	}
	return
}

func (c *conn) Write(b []byte) (n int, err error) {
	c.out.Lock()
	defer c.out.Unlock()

	n, err = c.out.Write(b)
	if c.Log != nil {
		c.Log(c.name, err, ">", string(b[:n]))
	}

	// signal other side reader for new content in buffer
	c.out.Signal()
	return
}

func (c *conn) Close() error {
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
