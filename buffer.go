package wstest

import (
	"bytes"
	"io"
	"sync"
)

// buffer is lockable conditional buffer
type buffer struct {
	buf    bytes.Buffer
	mutex  *sync.Mutex
	cond   *sync.Cond
	closed bool
}

// returns a new buffer
func newBuffer() *buffer {

	m := &sync.Mutex{}

	return &buffer{
		buf:   bytes.Buffer{},
		mutex: m,
		cond:  sync.NewCond(m),
	}
}

// Close buffer will not allow any farther reading from it
// If any reader is waiting, it will be returned with io.EOF error
func (b *buffer) Close() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// close the buffer and notify
	b.closed = true
	b.cond.Broadcast()
	return nil
}

// Read reads from buffer.
// if the buffer is empty, it will wait until something will be written into it or
// it will be closed.
func (b *buffer) Read(d []byte) (n int, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for {
		n, err = b.buf.Read(d)
		if err != io.EOF {
			break
		}

		// buffer is empty, nothing to read.
		// if it is closed, return EOF
		if b.closed {
			return 0, io.EOF
		}

		// wait for a signal from the other side writer
		b.cond.Wait()
	}

	return
}

// Write to buffer, and signal a reader that content was written
func (b *buffer) Write(d []byte) (n int, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	n, err = b.buf.Write(d)

	// signal so if there is any reader waiting, it will read the data
	b.cond.Signal()
	return
}
