package pipe

import (
	"context"
	"io"
	"testing"
	"time"
)

func TestBufferWriteRead(t *testing.T) {
	t.Parallel()

	b := newBuffer()
	testWrite(t, b, "hello")
	testWrite(t, b, ", ")
	testWrite(t, b, "world!")
	b.Close()

	testRead(t, b, "hello, world!", nil)

	testClosed(t, b)
}

func TestReadClosed(t *testing.T) {
	t.Parallel()
	b := newBuffer()
	b.Close()
	testClosed(t, b)
}

func TestReadBeforeWrite(t *testing.T) {
	t.Parallel()
	readDone := make(chan struct{})

	b := newBuffer()
	go func() {
		testRead(t, b, "hello, world!", nil)
		close(readDone)
	}()

	// sleep before writing to make sure that the read thread has
	// started and is blocked until any writing is done
	time.Sleep(time.Second)

	testWrite(t, b, "hello, world!")
	b.Close()

	select {
	case <-readDone:
	case <-time.After(time.Second):
		t.Error("Did not finished reading")
	}

	testClosed(t, b)
}

func TestBuffer_SetReadDeadline(t *testing.T) {
	t.Parallel()
	b := newBuffer()

	b.SetReadDeadline(time.Now())
	testRead(t, b, "", context.DeadlineExceeded)

	b.SetReadDeadline(time.Now().Add(time.Minute))
	testWrite(t, b, "hello")
	testRead(t, b, "hello", nil)

	// sets deadline to 0, should disable the timeout
	b.SetReadDeadline(time.Time{})
	testWrite(t, b, "hello")
	testRead(t, b, "hello", nil)

	b.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	<-time.After(100 * time.Millisecond)
	testRead(t, b, "", context.DeadlineExceeded)

	b.Close()
	testClosed(t, b)
}

func testRead(t *testing.T, b *buffer, want string, wantErr error) {
	data := make([]byte, 1024)
	n, err := b.Read(data)

	if err != wantErr {
		t.Errorf("err = %s, want: %s", err, wantErr)
	}
	if wantErr == nil {
		if got := string(data[:n]); got != want {
			t.Errorf("read = %s, want: %s", got, want)
		}
	} else {
		if want, got := 0, n; want != got {
			t.Errorf("n = %d, want: %d", got, want)
		}
	}
}

func testWrite(t *testing.T, b *buffer, data string) {
	n, err := b.Write([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := n, len([]byte(data)); got != want {
		t.Errorf("n = %d, want: %d", got, want)
	}
}

func testClosed(t *testing.T, b *buffer) {
	testRead(t, b, "", io.EOF)
}
