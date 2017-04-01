package wstest

import "github.com/gorilla/websocket"

// Message passed through a websocket
// Type is defined in RFC 6455, section 11.8, can be used with
// the gorilla/websocket package constants.
type Message struct {
	Type int
	Data []byte
}

// NewTextMessage returns a Message with type == 1.
func NewTextMessage(data []byte) *Message {
	return &Message{websocket.TextMessage, data}

}
