package wstest

import (
	"bytes"
	"sync"
)

// buffer is lockable conditional buffer
type buffer struct {
	bytes.Buffer
	*sync.Mutex
	*sync.Cond
}

// returns a new buffer
func newBuffer() *buffer {
	m := &sync.Mutex{}
	return &buffer{bytes.Buffer{}, m, sync.NewCond(m)}
}
