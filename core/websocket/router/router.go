// Make router a module.
package router

import (
	"encoding/json"
	"loveair/core/websocket/contracts"
	"loveair/log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/xid"
)

type Router struct {
	Connection *websocket.Conn
	// channel is buffered so as to mitigate the chance of blocking the writer to router incase of a router restart or failure.
	EmitCh            chan contracts.Contract
	WriterCh          chan Payload
	WriterTerminateCh chan bool
	ReaderCh          chan *Payload
	ReaderErrCh       chan error
	addr              string
	instanceID        string
	sLogger           log.SLoger
}

func (r *Router) KeepAlive() {

	t := 2 * time.Second
	for {
		r.sLogger.Log.Infof("Retrying in: %v seconds", t.Seconds())
		time.Sleep(t)
		Connection, _, err := websocket.DefaultDialer.Dial(r.addr+r.instanceID, nil)
		if err != nil {
			r.sLogger.Log.Errorln(err)
			t = 2 * t
			continue
		} else {
			r.Connection = Connection
			r.StartIO()
			return
		}
	}
}

func NewRouter(addr string, serviceLogger log.SLoger) (*Router, error) {
	id := xid.New().String()
	Connection, _, err := websocket.DefaultDialer.Dial(addr+id, nil)
	if err != nil {
		return &Router{
			Connection:        Connection,
			EmitCh:            make(chan contracts.Contract, 1024),
			WriterCh:          make(chan Payload),
			WriterTerminateCh: make(chan bool),
			ReaderCh:          make(chan *Payload),
			ReaderErrCh:       make(chan error),
			addr:              addr,
			instanceID:        id,
			sLogger:           serviceLogger,
		}, err
	}

	return &Router{
		Connection:        Connection,
		EmitCh:            make(chan contracts.Contract, 1024),
		WriterCh:          make(chan Payload),
		WriterTerminateCh: make(chan bool),
		ReaderCh:          make(chan *Payload),
		ReaderErrCh:       make(chan error),
		addr:              addr,
		instanceID:        id,
		sLogger:           serviceLogger,
	}, nil
}

func (r *Router) Close() {
	r.Connection.Close()
}

func (r *Router) Daemon() {
	go func() {
		for payload := range r.EmitCh {
			byt, err := json.Marshal(payload)
			if err != nil {
				r.sLogger.Log.Errorln(err)
				continue
			}

			r.WriterCh <- Payload{
				Headers: Table{"contract-name": payload.ContractName()},
				Body:    byt,
			}

		}
	}()
}

func (r *Router) StartIO() {
	go r.reader(r.ReaderCh, r.ReaderErrCh)
	go r.writer(r.WriterCh, r.WriterTerminateCh)
}

// The message types are defined in RFC 6455, section 11.8.
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1
)

type Table map[string]string

type Payload struct {
	Headers Table `json:"headers"`
	// The application specific payload of the message
	Body []byte `json:"body"`
}

func (r *Router) GetInstanceID() string {
	return r.instanceID
}
