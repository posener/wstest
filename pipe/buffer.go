package pipe

import (
	"bytes"
	"context"
	"io"
	"sync"
	"time"
)

// buffer is lockable conditional buffer
type buffer struct {
	buf   bytes.Buffer
	mutex *sync.Mutex
	cond  *sync.Cond
	r     op
	w     op
}

// returns a new buffer
func newBuffer() *buffer {

	m := &sync.Mutex{}

	b := &buffer{
		buf:   bytes.Buffer{},
		mutex: m,
		cond:  sync.NewCond(m),
	}

	// reads require broadcasting
	b.r.broadcast = true

	return b
}

var deadlineZero time.Time

// Close buffer will not allow any farther reading from it
// If any reader is waiting, it will be returned with io.EOF error
func (b *buffer) Close() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// if there are current deadline goroutine, clean them up.
	b.r.cancelDeadline()
	b.w.cancelDeadline()

	// set the error that the buffer is closed and notify waiters
	b.r.err = io.EOF
	b.w.err = io.EOF
	b.cond.Broadcast()

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
		if b.r.err != nil {
			return 0, b.r.err
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
	if b.w.err != nil {
		return 0, b.w.err
	}

	n, err := b.buf.Write(d)

	// signal so if there is any reader waiting, it will r the data
	b.cond.Signal()
	return n, err
}

// SetReadDeadline sets a deadline to the reader.
// if the deadline is reached, and there is a reader waiting for content,
// it will exit with the appropriate context deadline error.
func (b *buffer) SetReadDeadline(deadline time.Time) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.setDeadline(deadline, &b.r)
}

// SetWriteDeadline sets a deadline for the writer.
// since the writer is never blocked, it is only useful for
// past deadlines.
func (b *buffer) SetWriteDeadline(deadline time.Time) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.setDeadline(deadline, &b.w)
}

func (b *buffer) setDeadline(deadline time.Time, op *op) {
	// if there is a current deadline goroutine, cancel it before creating a new one
	op.cancelDeadline()

	// if closed, there is no deadline to set
	if op.err == io.EOF {
		return
	}

	// if deadline is deadlineZero, don't set a new deadline
	if deadline == deadlineZero {
		op.err = nil
		return
	}

	// if the deadline is in the past, update the error and notify
	// nothing else is need to be done.
	if deadline.Before(time.Now()) {
		op.err = context.DeadlineExceeded
		if op.broadcast {
			b.cond.Broadcast()
		}
		return
	}

	// create a new context with the desired deadline
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	op.cancel = cancel

	// start a deadline function that will fire according to the context
	go b.deadline(ctx, op)
}

// deadline waits for the context to be done
// if the context was cancelled, nothing will happen.
// if the deadline was reached, the current reader will return
// with the appropriate error.
func (b *buffer) deadline(ctx context.Context, op *op) {
	<-ctx.Done()

	b.mutex.Lock()
	defer b.mutex.Unlock()

	if op.err == io.EOF {
		return
	}

	// Set the the buffer error if it was a deadline error and the
	// error was not set already. Then wake up all current readers.
	if ctx.Err() == context.DeadlineExceeded {
		op.err = ctx.Err()
		if op.broadcast {
			b.cond.Broadcast()
		}
		return
	}

	op.err = nil
}
