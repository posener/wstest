package pipe

import (
	"testing"
)

func TestAddress(t *testing.T) {
	t.Parallel()

	network := "tcp"
	addr := "4.3.2.1:4321"

	a := &address{network: network, address: addr}

	if got, want := a.Network(), network; got != want {
		t.Errorf("a.Network() = %s, want %s", got, want)
	}

	if got, want := a.String(), addr; got != want {
		t.Errorf("a.String() = %s, want %s", got, want)
	}
}
