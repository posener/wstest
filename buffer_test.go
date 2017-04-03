package wstest

import (
	"io"
	"io/ioutil"
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

	testRead(t, b, "hello, world!")

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
		testRead(t, b, "hello, world!")
		close(readDone)
	}()

	// sleep before writing to make sure that the read thread has
	// started and is blocked until any writing is done
	time.Sleep(time.Second)

	testWrite(t, b, "hello")
	testWrite(t, b, ", ")
	testWrite(t, b, "world!")
	b.Close()

	select {
	case <-readDone:
	case <-time.After(time.Second):
		t.Error("Did not finished reading")
	}

	testClosed(t, b)
}

func testRead(t *testing.T, b *buffer, want string) {
	read, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(read); got != want {
		t.Errorf("read = %s, want: %s", got, want)
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
	data := make([]byte, 1024)
	n, err := b.Read(data)

	// after all was read, test that we get EOF
	if want, got := 0, n; want != got {
		t.Errorf("n = %d, want: %d", got, want)
	}
	if want, got := io.EOF, err; want != got {
		t.Errorf("err = %s, want: %s", got, want)
	}
}
