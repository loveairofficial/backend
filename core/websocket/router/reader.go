package router

import "encoding/json"

func (r *Router) reader(readerCh chan *Payload, eCh chan error) {

	for {
		messageType, inPayload, err := r.Connection.ReadMessage()
		if err != nil {
			eCh <- err
			r.sLogger.Log.Errorln(err)
			return
		}

		pl := new(Payload)

		if err := json.Unmarshal(inPayload, &pl); err != nil {
			eCh <- err
			r.sLogger.Log.Errorln(err)
		}

		if messageType == TextMessage {
			readerCh <- pl
			continue
		}
	}
}
