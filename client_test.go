package wstest

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
)

// TestClient demonstrate the usage of wstest package
func TestClient(t *testing.T) {
	var (
		// simple server
		s = &server{}

		// create a new websocket test client
		c = NewClient(10)
	)

	defer c.Close()
	defer s.Close()

	// first connect to s.
	// this send an HTTP request to the http.Handler, and wait for the connection upgrade response.
	// it uses the gorilla's websocket.Dial function, over a fake net.Conn struct.
	// it runs the s's ServeHTTP function in a goroutine, so s can communicate with a
	// c running on the current program flow
	err := c.Connect(s, "ws://example.org/ws")
	if err != nil {
		t.Fatalf("Failed connecting to s: %s", err)
	}

	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf("hello, world! %d", i)
		var m *Message

		// send a message in the websocket
		c.Send(NewTextMessage([]byte(msg)))

		m, err = s.Receive()
		if err != nil {
			t.Fatal(err)
		}

		if want, got := msg, string(m.Data); want != got {
			t.Errorf("Failed sending to server: %s != %s", want, got)
		}

		s.Send(NewTextMessage([]byte(msg)))

		m, err = c.Receive()
		if err != nil {
			t.Fatal(err)
		}

		if want, got := msg, string(m.Data); want != got {
			t.Errorf("Failed sending to server: %s != %s", want, got)
		}
	}
}

type server struct {
	upgrader websocket.Upgrader
	conn     *websocket.Conn
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	if r.URL.Path != "/ws" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	s.conn, err = s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
}

func (s *server) Receive() (*Message, error) {
	mType, data, err := s.conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	return &Message{mType, data}, nil
}

func (s *server) Send(m *Message) error {
	err := s.conn.WriteMessage(m.Type, m.Data)
	if err != nil {
		return err
	}
	return nil
}

func (s *server) Close() error {
	return s.conn.Close()
}
