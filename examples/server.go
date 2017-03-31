package example

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type echoServer struct {
	upgrader websocket.Upgrader
	done     chan struct{}
}

func newEchoServer() *echoServer {
	return &echoServer{
		done: make(chan struct{}),
	}
}

func (s *echoServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	defer close(s.done)

	conn, err := s.upgrader.Upgrade(w, r, nil)
	defer conn.Close()
	if err != nil {
		panic(err)
	}

	for r.Context().Err() == nil {

		mType, m, err := conn.ReadMessage()
		if err != nil {
			log.Println("failed read:", err)
			return
		}

		log.Println("server echo:", string(m))

		err = conn.WriteMessage(mType, m)
		if err != nil {
			log.Println("failed write:", err)
			return
		}
	}
}

func (s *echoServer) Wait() <-chan struct{} {
	return s.done
}
