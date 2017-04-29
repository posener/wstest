package pipe

import (
	"context"
	"io"
	"sync"
	"time"
)

// deadlineZero is used to clear the deadline of the operation
var deadlineZero time.Time

// state is used for storing the read and write operations state
type state struct {

	// err saves the current error for the specified operation
	// a deadlineExceeded error or EOF error.
	// errMu used to lock err state changes
	err   error
	errMu *sync.Mutex

	// errBroadcast is a cond.Broadcast function for notifying when
	// operation has gone into error state.
	// it can be set to nil if no broadcast is needed.
	errBroadcast func()

	// cancel is a function for cancelling a running background deadline goroutine
	cancel context.CancelFunc

	// doneCh is channel used to communicate between background deadline goroutine
	// to their cancellation thread
	doneCh chan struct{}

	// syncMu is mutex for the cancel and done members
	syncMu *sync.Mutex
}

// newState returns a new operation state with errBroadcast function
func newState(errBroadcast func()) *state {
	return &state{
		errBroadcast: errBroadcast,
		syncMu:       &sync.Mutex{},
		errMu:        &sync.Mutex{},
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
	if s.err != nil && s.errBroadcast != nil {
		// If error is not nil, wake up whoever waits
		s.errBroadcast()
	}
}

// Error returns operation error state
func (s *state) Error() error {
	s.errMu.Lock()
	defer s.errMu.Unlock()
	return s.err
}

// CancelDeadline cancels the background deadline goroutine if relevant
// and wait for the background goroutine if it is running
func (s *state) CancelDeadline() {
	s.syncMu.Lock()
	if s.cancel == nil {
		s.syncMu.Unlock()
		return
	}

	// cancel the current running background deadline goroutine
	s.cancel()
	s.syncMu.Unlock()

	// wait for deadline operation to finish
	<-s.doneCh
}

// Deadline sets deadline for the operation
func (s *state) Deadline(deadline time.Time) {

	// if there is a current deadline goroutine, cancel it before anything else.
	s.CancelDeadline()

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

// done clears the cancel and doneCh when background goroutine finishes its run
func (s *state) done() {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()

	if s.cancel == nil {
		return
	}
	s.cancel = nil
	close(s.doneCh)
}
