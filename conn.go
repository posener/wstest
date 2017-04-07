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
	err = c.opError("read", err)

	c.log(c.name, err, "<", string(b[:n]))
	return
}

// Write to out buffer
func (c *conn) Write(b []byte) (n int, err error) {
	n, err = c.out.Write(b)
	err = c.opError("write", err)

	c.log(c.name, err, ">", string(b[:n]))
	return
}

// Close the out buffer
func (c *conn) Close() error {
	return c.opError("close", c.out.Close())
}

// SetDeadLine sets the read deadline from the input buffer
func (c *conn) SetDeadline(t time.Time) error {
	c.in.SetReadDeadline(t)
	return nil
}

// SetReadDeadline sets the read deadline from the input buffer
func (c *conn) SetReadDeadline(t time.Time) error {
	c.in.SetReadDeadline(t)
	return nil
}

// SetWriteDeadline to a connection.
// Write to a buffer is non blocking and will always happen, so
// setting a deadline is meaningless.
func (c *conn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (c *conn) LocalAddr() net.Addr { return c.local }

func (c *conn) RemoteAddr() net.Addr { return c.remote }

// log debug messages, if Log was defined
func (c *conn) log(i ...interface{}) {
	if c.Log != nil {
		c.Log(i...)
	}
}

// opError converts error to a net.OpError
func (c *conn) opError(op string, err error) error {
	if err == nil {
		return nil
	}
	return &net.OpError{Op: op, Err: err, Source: c.local, Addr: c.remote, Net: "tcp"}
}

type address struct {
	network string
	address string
}

func (a *address) Network() string { return a.network }

func (a *address) String() string { return a.address }
