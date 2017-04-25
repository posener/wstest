package wstest_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/posener/wstest"
)

// TestClient demonstrate the usage of wstest package
func TestClient(t *testing.T) {
	t.Parallel()
	var (
		s = &handler{Upgraded: make(chan struct{})}
		d = wstest.NewDialer(s, t.Log)
	)

	c, resp, err := d.Dial("ws://example.org/ws", nil)
	if err != nil {
		t.Fatalf("Failed connecting to s: %s", err)
	}

	<-s.Upgraded

	if got, want := resp.StatusCode, http.StatusSwitchingProtocols; got != want {
		t.Errorf("resp.StatusCode = %q, want %q", got, want)
	}

	for i := 0; i < 3; i++ {
		msg := fmt.Sprintf("hello, world! %d", i)

		err := c.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			t.Fatal(err)
		}

		mT, m, err := s.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}

		if want, got := msg, string(m); want != got {
			t.Errorf("dialer got %q, want  %q", got, want)
		}
		if want, got := websocket.TextMessage, mT; want != got {
			t.Errorf("message type = %q, want %q", got, want)
		}

		s.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			t.Fatal(err)
		}

		mT, m, err = c.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}

		if want, got := msg, string(m); want != got {
			t.Errorf("client got %q, want  %q", got, want)
		}
		if want, got := websocket.TextMessage, mT; want != got {
			t.Errorf("message type = %q , want %q", got, want)
		}
	}

	err = c.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = s.Close()
	if err != nil {
		t.Fatal(err)
	}
}

// TestConcurrent tests concurrent reads and writes from a connection
func TestConcurrent(t *testing.T) {
	t.Parallel()
	var (
		s     = &handler{Upgraded: make(chan struct{})}
		d     = wstest.NewDialer(s, nil)
		count = 20
	)

	c, _, err := d.Dial("ws://example.org/ws", nil)
	if err != nil {
		t.Fatalf("Failed connecting to s: %s", err)
	}

	<-s.Upgraded

	for _, pair := range []struct{ src, dst *websocket.Conn }{{s.Conn, c}, {c, s.Conn}} {
		go func() {
			for i := 0; i < count; i++ {
				pair.src.WriteJSON(i)
			}
		}()

		received := make([]bool, count)

		for i := 0; i < count; i++ {
			var j int
			pair.dst.ReadJSON(&j)

			received[j] = true
		}

		missing := []int{}

		for i := range received {
			if !received[i] {
				missing = append(missing, i)
			}
		}
		if len(missing) > 0 {
			t.Errorf("%q -> %q: Did not received: %q", pair.src.LocalAddr(), pair.dst.LocalAddr(), missing)
		}
	}

	err = c.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = s.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestBadAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		url  string
		code int
	}{

		{
			url:  "ws://example.org/not-ws",
			code: http.StatusNotFound,
		},
		{
			url: "http://example.org/ws",
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			s := &handler{Upgraded: make(chan struct{})}
			d := wstest.NewDialer(s, nil)
			c, resp, err := d.Dial(tt.url, nil)
			if c != nil {
				t.Errorf("d = %T, want nil", c)
			}
			if err == nil {
				t.Error("opError is nil")
			}
			if tt.code != 0 {
				if got, want := resp.StatusCode, tt.code; got != want {
					t.Errorf("resp.StatusCode = %q, want %q", got, want)
				}
			}

			err = s.Close()
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestConnectDeadline tests connection deadlines
func TestDeadlines(t *testing.T) {
	t.Parallel()
	h := &handler{Upgraded: make(chan struct{})}
	d := wstest.NewDialer(h, nil)

	c, _, err := d.Dial("ws://example.org/ws", nil)
	if err != nil {
		t.Fatalf("Failed connecting to h: %q", err)
	}

	<-h.Upgraded

	var i int

	for _, pair := range []struct{ src, dst *websocket.Conn }{{h.Conn, c}, {c, h.Conn}} {

		// set the deadline to now, and test for timeout
		pair.dst.SetReadDeadline(time.Now())
		err = pair.dst.ReadJSON(i)
		if got, want := err.Error(), context.DeadlineExceeded.Error(); !strings.Contains(got, want) {
			t.Errorf("err = %q, not conains %q", got, want)
		}
		err = pair.dst.ReadJSON(i)
		if got, want := err.Error(), context.DeadlineExceeded.Error(); !strings.Contains(got, want) {
			t.Errorf("err = %q, not conains %q", got, want)
		}

		pair.src.WriteJSON(1)
		err = pair.dst.ReadJSON(i)
		if got, want := err.Error(), context.DeadlineExceeded.Error(); !strings.Contains(got, want) {
			t.Errorf("err = %q, not conains %q", got, want)
		}

		// even after updating the deadline, should get an error
		pair.dst.SetReadDeadline(time.Now().Add(time.Second))
		err = pair.dst.ReadJSON(i)
		if got, want := err.Error(), context.DeadlineExceeded.Error(); !strings.Contains(got, want) {
			t.Errorf("err = %q, not conains %q", got, want)
		}
	}
}

// TestConnectDeadline tests connection deadline
func TestConnectDeadline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path    string
		timeout time.Duration
		err     error
	}{
		{
			"/ws/delay",
			time.Millisecond,
			context.DeadlineExceeded,
		},
		{
			"/ws",
			time.Second,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s/%s", tt.path, tt.timeout), func(t *testing.T) {
			s := &handler{Upgraded: make(chan struct{})}
			d := wstest.NewDialer(s, nil)
			d.HandshakeTimeout = tt.timeout
			_, _, err := d.Dial("ws://example.org"+tt.path, nil)
			if tt.err == nil {
				if err != nil {
					t.Errorf("err = %q, want nil", err)
				}
			} else {
				if got, want := err.(*net.OpError).Err, tt.err; got != want {
					t.Errorf("err = %q, want %q", got, want)
				}
			}

			if tt.err == nil {
				select {
				case <-s.Upgraded:
				case <-time.After(time.Second):
					t.Fatal("connection was not upgraded after 1s")
				}
			}
		})
	}
}

// dialer for test purposes, can't handle multiple websocket connections concurrently
type handler struct {
	*websocket.Conn
	upgrader websocket.Upgrader
	Upgraded chan struct{}
}

func (s *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.URL.Path {
	case "/ws":
		s.connect(w, r)

	case "/ws/delay":
		<-time.After(500 * time.Millisecond)
		s.connect(w, r)

	default:
		w.WriteHeader(http.StatusNotFound)
	}

}

func (s *handler) connect(w http.ResponseWriter, r *http.Request) {
	var err error
	s.Conn, err = s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	close(s.Upgraded)
}

func (s *handler) Close() error {
	if s.Conn == nil {
		return nil
	}
	return s.Conn.Close()
}
