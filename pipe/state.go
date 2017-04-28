package pipe

import (
	"context"
	"io"
	"sync"
	"time"
)

var deadlineZero time.Time

// state is a struct used for read and write operations
// it holds the operation state
type state struct {

	// err saves the current error for the specified operation
	// a deadlineExceeded error or EOF error.
	// errMu used to lock err state changes
	err   error
	errMu *sync.Mutex

	// broadcast specify if to broadcast the condition for the
	// specific operation.
	broadcast func()

	// cancel is a function for cancelling a running deadline goroutine
	cancel context.CancelFunc

	// doneCh is channel used to communicate between background goroutine
	// to their cancellation thread
	doneCh chan struct{}

	// syncMu is mutex for the cancel and done members
	syncMu *sync.Mutex
}

func newState(broadcast func()) *state {
	return &state{
		broadcast: broadcast,
		syncMu:    &sync.Mutex{},
		errMu:     &sync.Mutex{},
	}
}

// SetError sets the operation state error
func (s *state) SetError(err error) {
	s.errMu.Lock()
	defer s.errMu.Unlock()
	if s.err == io.EOF {
		return
	}
	s.err = err
	if s.err != nil && s.broadcast != nil {
		// If error is not nil, wake up whoever waits
		s.broadcast()
	}
}

func (s *state) Error() error {
	s.errMu.Lock()
	defer s.errMu.Unlock()
	return s.err
}

// Deadline sets deadline for the operation
func (s *state) Deadline(deadline time.Time) {

	// if there is a current deadline goroutine, cancel it before anything else.
	s.Cancel()

	// if deadline is deadlineZero, don't set a new deadline
	if deadline == deadlineZero {
		s.SetError(nil)
		return
	}

	// if the deadline is in the past, update the error, nothing else is need to be done.
	if deadline.Before(time.Now()) {
		s.SetError(context.DeadlineExceeded)
		return
	}

	ctx := s.context(deadline)

	// start a deadline function that will fire according to the Context
	go s.background(ctx)
}

// Cancel cancels the background deadline goroutine if relevant
// and wait for the background goroutine if it is running
func (s *state) Cancel() {
	s.syncMu.Lock()
	if s.cancel == nil {
		s.syncMu.Unlock()
		return
	}
	s.cancel()
	s.syncMu.Unlock()
	<-s.doneCh
}

// background waits for the Context to be done
// if the Context was cancelled, nothing will happen.
// if the deadline was reached, the current reader will return
// with the appropriate error.
func (s *state) background(ctx context.Context) {

	// wait for context to finish
	<-ctx.Done()

	// set the current state of the operation to done
	s.done()

	// when deadline happened, set the appropriate error of the operation
	if ctx.Err() != context.Canceled {
		s.SetError(ctx.Err())
	}
}

// context creates a context and updates the cancel func of the operation
func (s *state) context(deadline time.Time) context.Context {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()

	// create a new Context with the desired deadline
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	s.cancel = cancel
	s.doneCh = make(chan struct{})
	return ctx
}

func (s *state) done() {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()

	if s.cancel == nil {
		return
	}
	s.cancel = nil
	close(s.doneCh)
}
