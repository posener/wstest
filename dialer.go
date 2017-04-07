package wstest

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/websocket"
)

type dialer struct {
	httptest.ResponseRecorder
	server *conn
	client *conn
}

// NewDialer creates a wstest dialer to an http.Handler which accepts websocket upgrades.
// This send an HTTP request to the http.Handler, and wait for the connection upgrade response.
// it runs the dialer's ServeHTTP function in a goroutine, so dialer can communicate with a
// client running on the current program flow
//
// h is the handler that should handle websocket connections.
// debugLog is a function for a log.Println-like function for printing everything that
// is passed over the connection.
// It returns a *websocket.Dial struct, which can then be used to dial to the handler.
func NewDialer(h http.Handler, debugLog log) *websocket.Dialer {
	var (
		s2c   = newBuffer()
		c2s   = newBuffer()
		cAddr = &address{"tcp", "1.2.3.4:12345"}
		sAddr = &address{"tcp", "5.6.7.8:12346"}
	)

	conn := &dialer{
		server: &conn{name: "dialer", in: c2s, out: s2c, local: sAddr, remote: cAddr, Log: debugLog},
		client: &conn{name: "client", in: s2c, out: c2s, local: cAddr, remote: sAddr, Log: debugLog},
	}

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
		panic(err)
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
