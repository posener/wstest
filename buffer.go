package wstest

import (
	"bytes"
	"context"
	"io"
	"sync"
	"time"
)

// buffer is lockable conditional buffer
type buffer struct {
	buf            bytes.Buffer
	mutex          *sync.Mutex
	cond           *sync.Cond
	err            error
	cancelDeadline context.CancelFunc
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

	// if there is a current reader deadline goroutine, clean it up.
	if b.cancelDeadline != nil {
		b.cancelDeadline()
		b.cancelDeadline = nil
	}

	// close the buffer and notify
	b.err = io.EOF
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
		// non-blocking read from buffer, if it is empty EOF will be returned
		n, err = b.buf.Read(d)
		if err != io.EOF {
			break
		}

		// check for errors
		if b.err != nil {
			return 0, b.err
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

// SetReadDeadline sets a deadline to the reader.
// if the deadline is reached, and there is a reader waiting for content,
// it will exit with the appropriate context deadline error.
func (b *buffer) SetReadDeadline(deadline time.Time) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// if there is a current readDeadline goroutine, cancel it before creating a new one
	if b.cancelDeadline != nil {
		b.cancelDeadline()
		b.cancelDeadline = nil
	}

	// create a new context with the desired readDeadline
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	b.cancelDeadline = cancel

	// start a readDeadline function that will fire according to the context
	go b.readDeadline(ctx)
}

// readDeadline waits for the context to be done
// if the context was cancelled, nothing will happen.
// if the readDeadline was reached, the current reader will return
// with the appropriate error.
func (b *buffer) readDeadline(ctx context.Context) {
	<-ctx.Done()

	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Set the the buffer error if it was a deadline error and the
	// error was not set already. Then wake up all current readers.
	if ctx.Err() == context.DeadlineExceeded && b.err == nil {
		b.err = ctx.Err()
		b.cond.Broadcast()
	}
}
