package example

import (
	"testing"

	"fmt"
	"github.com/posener/wstest"
)

// TestExample demonstrate the usage of wstest package
func TestExample(t *testing.T) {
	var (
		// simple echo server that returns everything it receives on a websocket
		server = newEchoServer()

		// create a new websocket test client
		client = wstest.NewClient()
	)

	// first connect to server.
	// this send an HTTP request to the http.Handler, and wait for the connection upgrade response.
	// it uses the gorilla's websocket.Dial function, over a fake net.Conn struct.
	// it runs the server's ServeHTTP function in a goroutine, so server can communicate with a
	// client running on the current program flow
	err := client.Connect(server)
	if err != nil {
		t.Fatalf("Failed connecting to echoServer: %server", err)
	}

	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf("hello, world! %d", i)

		// send a message in the websocket
		client.Send(wstest.NewTextMessage([]byte(msg)))

		// receive a message from the websocket
		received, err := client.Receive()
		if err != nil {
			t.Fatal(err)
		}

		// check if the echo server returned the same message
		if want, got := msg, string(received.Data); want != got {
			t.Errorf("Failed echoing: %s != %s", want, got)
		}
	}

	// close the client side of the weboscket connection.
	client.Close()

	// after the client have closed the connection, the server's connection handling
	// thread should also break. (this is specific for the echo server implementation)
	<-server.Wait()
}
