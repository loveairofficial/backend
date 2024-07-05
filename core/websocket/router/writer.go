package router

import "encoding/json"

func (r *Router) writer(writerChan chan Payload, terminate chan bool) {

	for {
		select {
		case payload := <-writerChan:
			byt, err := json.Marshal(payload)
			if err != nil {
				r.sLogger.Log.Errorln(err)
				return
			}

			err = r.Connection.WriteMessage(TextMessage, byt)
			if err != nil {
				r.sLogger.Log.Errorln(err)
				continue
			}
			continue
		case <-terminate:
			// let Emit know to stop writing and probably accumulate the payload.
			// or fallback to anther way
			return
		}
	}
}
