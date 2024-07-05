package gorilla

import (
	"encoding/json"
	"loveair/models"

	"github.com/gorilla/websocket"
)

func (s *Socket) writer(conn *websocket.Conn, OutPayloadCh chan *models.Outgoing, eCh chan error) {
	for payload := range OutPayloadCh {
		byt, err := json.Marshal(payload.OutPayload)
		if err != nil {
			eCh <- err
			return
		}
		err = conn.WriteMessage(TextMessage, byt)
		if err != nil {
			eCh <- err
			return
		}
	}
}
