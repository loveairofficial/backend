package gorilla

import (
	"loveair/models"

	"github.com/gorilla/websocket"
)

func (s *Socket) reader(conn *websocket.Conn, incChan chan models.Incomming, eCh chan error) {
	for {
		messageType, inPayload, err := conn.ReadMessage()
		if err != nil {
			eCh <- err
			return
		}
		inc := models.Incomming{
			Mt:        messageType,
			InPayload: inPayload,
		}
		incChan <- inc
	}
}
