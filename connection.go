package wstest

import (
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
	Log    log
}

type log func(...interface{})

// Read from in buffer
func (c *conn) Read(b []byte) (n int, err error) {
	n, err = c.in.Read(b)
	c.log(c.name, err, "<", string(b[:n]))
	return
}

// Write to out buffer
func (c *conn) Write(b []byte) (n int, err error) {
	n, err = c.out.Write(b)
	c.log(c.name, err, ">", string(b[:n]))
	return
}

// Close the out buffer
func (c *conn) Close() error {
	return c.out.Close()
}

// log debug messages, if Log was defined
func (c *conn) log(i ...interface{}) {
	if c.Log != nil {
		c.Log(i...)
	}
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
