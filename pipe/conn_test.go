package pipe

import (
	"testing"
)

func TestConn(t *testing.T) {
	t.Parallel()

	b := newBuffer()

	local := &address{"tcp", "4.3.2.1:4321"}
	remote := &address{"tcp", "1.2.3.4:1234"}

	c := &conn{
		name:   "test",
		in:     b,
		out:    b,
		remote: remote,
		local:  local,
		log:    t.Log,
	}

	if got, want := c.RemoteAddr(), remote; got != want {
		t.Errorf("c.RemoteAddress = %s, want %s", got, want)
	}

	if got, want := c.LocalAddr(), local; got != want {
		t.Errorf("c.RemoteAddress = %s, want %s", got, want)
	}

	wrote := []byte("hello")
	n, err := c.out.Write(wrote)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := n, len(wrote); got != want {
		t.Errorf("n = %d, want %d", got, want)
	}

	read := make([]byte, n)
	n, err = c.in.Read(read)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := n, len(wrote); got != want {
		t.Errorf("n = %d, want %d", got, want)
	}
	if got, want := string(read), string(wrote); got != want {
		t.Errorf("read = %s, want %s", got, want)
	}
}
