package pipe

import (
	"net"
	"testing"

	"github.com/golang/net/nettest"
)

// TestPipe tests the pipe implementation with the nettest.TestConn tests suite.
func TestPipe(t *testing.T) {
	t.Parallel()

	nettest.TestConn(t, func() (c1, c2 net.Conn, stop func(), err error) {
		c1, c2 = New(nil)
		stop = func() {
			c1.Close()
			c2.Close()
		}
		return
	})
}
