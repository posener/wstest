package wstest

import "github.com/gorilla/websocket"

type Message struct {
	Type int
	Data []byte
}

func NewTextMessage(data []byte) *Message {
	return &Message{websocket.TextMessage, data}

}
