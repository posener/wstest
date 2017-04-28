package pipe

import (
	"bytes"
	"io"
	"sync"
	"time"
)

// buffer is lockable conditional buffer
type buffer struct {
	buf   bytes.Buffer
	mutex *sync.Mutex
	cond  *sync.Cond
	r     *state
	w     *state
}

// returns a new buffer
func newBuffer() *buffer {

	m := &sync.Mutex{}
	c := sync.NewCond(m)

	b := &buffer{
		buf:   bytes.Buffer{},
		mutex: m,
		cond:  c,
		r:     newState(c.Broadcast),
		w:     newState(nil),
	}

	return b
}

// Close buffer will not allow any farther reading from it
// If any reader is waiting, it will be returned with io.EOF error
func (b *buffer) Close() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// if there are current deadline goroutine, clean them up.
	b.r.Cancel()
	b.w.Cancel()

	// set the error that the buffer is closed and notify waiters
	b.r.SetError(io.EOF)
	b.w.SetError(io.EOF)

	return nil
}

// Read reads from buffer.
// if the buffer is empty, it will wait until something will be written into it or
// it will be closed.
func (b *buffer) Read(d []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for {
		// non-blocking r from buffer, if it is empty EOF will be returned
		n, err := b.buf.Read(d)
		if err != io.EOF {
			return n, err
		}

		// check for errors
		err = b.r.Error()
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
	err := b.w.Error()
	if err != nil {
		return 0, err
	}

	n, err := b.buf.Write(d)

	// signal so if there is any reader waiting, it will r the data
	b.cond.Signal()
	return n, err
}

// SetReadDeadline sets a deadline to the reader.
// if the deadline is reached, and there is a reader waiting for content,
// it will exit with the appropriate Context deadline error.
func (b *buffer) SetReadDeadline(deadline time.Time) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.r.Deadline(deadline)
}

// SetWriteDeadline sets a deadline for the writer.
// since the writer is never blocked, it is only useful for
// past deadlines.
func (b *buffer) SetWriteDeadline(deadline time.Time) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.w.Deadline(deadline)
}
