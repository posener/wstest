package pipe

import "net"

func New(debugLog func(...interface{})) (net.Conn, net.Conn) {
	var (
		s2c   = newBuffer()
		c2s   = newBuffer()
		cAddr = &address{"tcp", "1.2.3.4:12345"}
		sAddr = &address{"tcp", "5.6.7.8:12346"}
	)

	client := &conn{name: "server", in: c2s, out: s2c, local: sAddr, remote: cAddr, logger: debugLog}
	server := &conn{name: "client", in: s2c, out: c2s, local: cAddr, remote: sAddr, logger: debugLog}

	return client, server
}
