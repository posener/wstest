package pipe

type address struct {
	network string
	address string
}

func (a *address) Network() string { return a.network }

func (a *address) String() string { return a.address }
