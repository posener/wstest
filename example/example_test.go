package example

import (
	"testing"

	"fmt"
	"github.com/posener/wstest"
)

func TestServer_ServeHTTP(t *testing.T) {
	var (
		server = newEchoServer()
		client = wstest.NewClient()
	)

	err := client.Connect(server)
	if err != nil {
		t.Fatalf("Failed connecting to echoServer: %server", err)
	}

	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf("hello, world! %d", i)

		client.Send(wstest.NewTextMessage([]byte(msg)))
		received, err := client.Receive()
		if err != nil {
			t.Fatal(err)
		}

		if want, got := msg, string(received.Data); want != got {
			t.Errorf("Failed echoing: %s != %s", want, got)
		}
	}

	client.Close()
	<-server.Wait()
}
