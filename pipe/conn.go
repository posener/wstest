package pipe

import (
	"io"
	"net"
	"time"
)

// conn is in-memory struct that implements the net.Conn interface.
// It uses two buffer instances to read and write from.
type conn struct {
	name   string
	in     *buffer
	out    *buffer
	local  net.Addr
	remote net.Addr
	log    Println
}

// Read from in buffer
func (c *conn) Read(b []byte) (n int, err error) {
	n, err = c.in.Read(b)
	err = c.opError("read", err)

	c.log.Println(c.name, err, "read", len(b[:n]))
	return
}

// Write to out buffer
func (c *conn) Write(b []byte) (n int, err error) {
	n, err = c.out.Write(b)
	err = c.opError("write", err)

	c.log.Println(c.name, err, "write", len(b[:n]))
	return
}

// Close the out buffer
func (c *conn) Close() error {
	inErr := c.in.Close()
	err := c.out.Close()
	if err == nil {
		err = inErr
	}
	return c.opError("close", err)
}

// SetDeadLine sets the read and write deadlines
func (c *conn) SetDeadline(t time.Time) error {
	c.log.Println(c.name, "set deadline", t)
	c.in.SetReadDeadline(t)
	c.out.SetWriteDeadline(t)
	return nil
}

// SetReadDeadline sets the read deadline from the input buffer
func (c *conn) SetReadDeadline(t time.Time) error {
	c.log.Println(c.name, "set read deadline", t)
	c.in.SetReadDeadline(t)
	return nil
}

// SetWriteDeadline sets the write deadline to the output buffer
func (c *conn) SetWriteDeadline(t time.Time) error {
	c.log.Println(c.name, "set write deadline", t)
	c.out.SetWriteDeadline(t)
	return nil
}

func (c *conn) LocalAddr() net.Addr { return c.local }

func (c *conn) RemoteAddr() net.Addr { return c.remote }

// opError converts error to a net.OpError
func (c *conn) opError(op string, err error) error {
	if err == nil || err == io.EOF {
		return err
	}
	return &net.OpError{Op: op, Err: err, Source: c.local, Addr: c.remote, Net: "tcp"}
}
