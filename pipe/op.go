package pipe

import "context"

// op is a struct used for read and write operations
type op struct {
	// cancel is a function for cancelling a running deadline goroutine
	cancel context.CancelFunc
	// error saves the current error for the specified operation
	// a deadlineExceeded error or EOF error.
	err error
	// broadcast specify if to broadcast the condition for the
	// specific operation.
	broadcast bool
}

// cancelDeadline cancels the deadline if relevant
func (o op) cancelDeadline() {
	if o.cancel != nil {
		o.cancel()
	}
	o.cancel = nil
}
