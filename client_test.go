package wstest

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
)

const count = 100

// TestClient demonstrate the usage of wstest package
func TestClient(t *testing.T) {
	t.Parallel()
	var (
		// simple server
		s = &server{Upgraded: make(chan struct{})}

		// create a new websocket test client
		c = NewClient()
	)

	// first connect to s.
	// this send an HTTP request to the http.Handler, and wait for the connection upgrade response.
	// it uses the gorilla's websocket.Dial function, over a fake net.Conn struct.
	// it runs the s's ServeHTTP function in a goroutine, so s can communicate with a
	// c running on the current program flow
	err := c.Connect(s, "ws://example.org/ws")
	if err != nil {
		t.Fatalf("Failed connecting to s: %s", err)
	}

	<-s.Upgraded

	for i := 0; i < count; i++ {
		msg := fmt.Sprintf("hello, world! %d", i)

		// send a message in the websocket
		err := c.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			t.Fatal(err)
		}

		mT, m, err := s.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}

		if want, got := msg, string(m); want != got {
			t.Errorf("server got %s, want  %s", got, want)
		}
		if want, got := websocket.TextMessage, mT; want != got {
			t.Errorf("message type = %s , want %s", got, want)
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
			t.Errorf("client got %s, want  %s", got, want)
		}
		if want, got := websocket.TextMessage, mT; want != got {
			t.Errorf("message type = %s , want %s", got, want)
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
		s = &server{Upgraded: make(chan struct{})}
		c = NewClient()
	)
	c.SetLogger(t.Log)

	err := c.Connect(s, "ws://example.org/ws")
	if err != nil {
		t.Fatalf("Failed connecting to s: %s", err)
	}

	<-s.Upgraded

	// server sends messages in a goroutine
	go func() {
		for i := 0; i < count; i++ {
			s.WriteJSON(i)
		}
	}()

	received := make([]bool, count)

	for i := 0; i < count; i++ {
		var j int
		c.ReadJSON(&j)

		received[j] = true
	}

	missing := []int{}

	for i := range received {
		if !received[i] {
			missing = append(missing, i)
		}
	}
	if len(missing) > 0 {
		t.Errorf("Did not received: %v", missing)
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

type server struct {
	*websocket.Conn
	upgrader websocket.Upgrader
	Upgraded chan struct{}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	if r.URL.Path != "/ws" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	s.Conn, err = s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	close(s.Upgraded)
}
