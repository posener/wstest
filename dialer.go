// Package wstest provides a NewDialer function to test just the
// `http.Handler` that upgrades the connection to a websocket session.
// It runs the handler function in a goroutine without listening on
// any port. The returned `websocket.Dialer` then can be used to dial
// and communicate with the given handler.

package wstest

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/websocket"
	"github.com/posener/wstest/pipe"
)

type dialer struct {
	httptest.ResponseRecorder
	client net.Conn
	server net.Conn
}

// NewDialer creates a wstest dialer to an http.Handler which accepts websocket upgrades.
// This send an HTTP request to the http.Handler, and wait for the connection upgrade response.
// it runs the dialer's ServeHTTP function in a goroutine, so dialer can communicate with a
// client running on the current program flow
//
// h is an http.Handler that handles websocket connections.
// debugLog is a function for a log.Println-like function for printing everything that
// is passed over the connection. Can be set to nil if no logs are needed.
// It returns a *websocket.Dial struct, which can then be used to dial to the handler.
func NewDialer(h http.Handler, debugLog pipe.Println) *websocket.Dialer {
	c1, c2 := pipe.New(debugLog)
	conn := &dialer{client: c1, server: c2}

	// run the runServer in a goroutine, so when the Dial send the request to
	// the dialer on the connection, it will be parsed as an HTTPRequest and
	// sent to the Handler function.
	go conn.runServer(h)

	// use the websocket.NewDialer.Dial with the fake net.dialer to communicate with the dialer
	// the dialer gets the client which is the client side of the connection
	return &websocket.Dialer{NetDial: func(network, addr string) (net.Conn, error) { return conn.client, nil }}
}

// runServer reads the request sent on the connection to the dialer
// from the websocket.NewDialer.Dial function, and pass it to the dialer.
// once this is done, the communication is done on the wsConn
func (d *dialer) runServer(h http.Handler) {
	// read from the dialer connection the request sent by the dialer.Dial,
	// and use the handler to serve this request.
	req, err := http.ReadRequest(bufio.NewReader(d.server))
	if err != nil {
		return
	}
	h.ServeHTTP(d, req)
}

// Hijack the connection
func (d *dialer) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	// return to the dialer the dialer, which is the dialer side of the connection
	rw := bufio.NewReadWriter(bufio.NewReader(d.server), bufio.NewWriter(d.server))
	return d.server, rw, nil
}

// WriteHeader write HTTP header to the client and closes the connection
func (d *dialer) WriteHeader(code int) {
	r := http.Response{StatusCode: code}
	r.Write(d.server)
	d.server.Close()
}
