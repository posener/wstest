package pipe

import (
	"bytes"
	"io"
	"sync"
	"time"
)

// buffer is a Reader/Writer/Closer from an in-memory bytes.Buffer that:
// * Blocks reads until some write occur (not returning io.EOF when buffer is empty)
// * Has read and write deadlines capabilities.
// It uses the state struct to hold the read and write current state.
type buffer struct {
	buf    bytes.Buffer
	mutex  *sync.Mutex
	cond   *sync.Cond
	rState *state
	wState *state
}

// returns a new buffer
func newBuffer() *buffer {

	m := &sync.Mutex{}
	c := sync.NewCond(m)

	b := &buffer{
		buf:    bytes.Buffer{},
		mutex:  m,
		cond:   c,
		rState: newState(c.Broadcast),
		wState: newState(nil),
	}

	return b
}

// Close buffer will not allow any farther reading from it
// If any reader is waiting, it will be returned with io.EOF error
func (b *buffer) Close() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// if there are current deadline goroutine, clean them up.
	b.rState.CancelDeadline()
	b.wState.CancelDeadline()

	// set the error that the buffer is closed and notify waiters
	b.rState.SetError(io.EOF)
	b.wState.SetError(io.EOF)

	return nil
}

// Read reads from buffer.
// if the buffer is empty, it will wait until something will be written into it or
// it will be closed.
func (b *buffer) Read(d []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for {
		// non-blocking rState from buffer, if it is empty EOF will be returned
		n, err := b.buf.Read(d)
		if err != io.EOF {
			return n, err
		}

		// check for errors
		err = b.rState.Error()
		if err != nil {
			return 0, err
		}

		// wait for a signal from the other side writer
		b.cond.Wait()
	}
}

// Write to buffer, and signal a reader that content was written
func (b *buffer) Write(d []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// if the write error is set, return it instead of writing.
	err := b.wState.Error()
	if err != nil {
		return 0, err
	}

	n, err := b.buf.Write(d)

	// signal so if there is any reader waiting, it will rState the data
	b.cond.Signal()
	return n, err
}

// SetReadDeadline sets a deadline to the reader.
// if the deadline is reached, and there is a reader waiting for content,
// it will exit with the appropriate Context deadline error.
func (b *buffer) SetReadDeadline(deadline time.Time) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.rState.Deadline(deadline)
}

// SetWriteDeadline sets a deadline for the writer.
// since the writer is never blocked, it is only useful for
// past deadlines.
func (b *buffer) SetWriteDeadline(deadline time.Time) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.wState.Deadline(deadline)
}
