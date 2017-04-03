package wstest

// newPairedConnections returns two connections, paired by buffers.
// any message written into the first connection, will be read in the second
// and vice-versa.
func newPairedConnections() (server, client *conn) {
	var (
		s2c   = newBuffer()
		c2s   = newBuffer()
		cAddr = &address{"tcp", "1.2.3.4:12345"}
		sAddr = &address{"tcp", "5.6.7.8:12346"}
	)

	server = &conn{name: "server", in: c2s, out: s2c, local: sAddr, remote: cAddr}
	client = &conn{name: "client", in: s2c, out: c2s, local: cAddr, remote: sAddr}
	return
}
