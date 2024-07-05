package gorilla

import (
	"encoding/json"
	"fmt"
	"loveair/models"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func (s *Socket) Connect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// check if connected client are a certain number and if so drop all future connections.
	if connCount := s.clients.Count(); connCount >= 100000 {
		s.sLogger.Log.Errorln("Instance Full ------------------------------------------------")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.sLogger.Log.Errorln(err)
	}

	receiveCh := make(chan models.Incomming)
	sendCh := make(chan *models.Outgoing)
	errCh := make(chan error)

	go s.reader(conn, receiveCh, errCh)
	go s.writer(conn, sendCh, errCh)

	Client := &models.Client{
		ID:     id,
		Conn:   conn,
		SendCh: sendCh,
	}

	s.join <- Client
	s.sLogger.Log.Infof("User Connected, ID: %s", id)

	defer func() {
		s.leave <- Client
		s.sLogger.Log.Infof("User Disconnected, ID: %s", id)
	}()

	s.delegate(receiveCh, sendCh, errCh)
}

func (s *Socket) delegate(receiveCh chan models.Incomming, sendCh chan *models.Outgoing, errCh chan error) {
	for {
		select {
		case i := <-receiveCh:
			var pl models.Payload
			if err := json.Unmarshal(i.InPayload, &pl); err != nil {
				s.sLogger.Log.Errorln(err)
			}
			switch pl.Tag {
			case "ping":
				s.sLogger.Log.Infof("ping")
				sendCh <- &models.Outgoing{
					OutPayload: models.Payload{
						Tag: "pong",
					},
				}
			case "init-match-call":
				s.initMatchCall(&pl.Data, sendCh)
			case "reinit-match-call":
				s.reInitMatchCall(&pl.Data, sendCh)
			case "meet-request-status-update":
				s.meetRequestStatusUpdate(&pl.Data)
			case "match-status-update":
				s.matchStatusUpdate(&pl.Data)
			case "new-message":
				s.newMessage(&pl.Data)
			case "update-message-status":
				s.updateMessageStatus(&pl.Data)
			case "pass-status-update":
				s.passStatusUpdate(&pl.Data)
			case "report":
				s.report(&pl.Data)

			default:
				s.sLogger.Log.Errorln("Invalid Tag")
			}
		case err := <-errCh:
			s.sLogger.Log.Errorln("read error: ", err)
			close(errCh)
			close(sendCh)
			close(receiveCh)
			return
		}
	}
}

//~ Meet request.

func (s *Socket) ClientConnectedToThisInstance(recieverID string, og *models.Outgoing, ackChan chan bool) {
	defer close(ackChan)
	client, ok := s.clients.Get(recieverID)
	ackChan <- ok
	if ok {
		client.SendCh <- og
	}
}

func uniqueValues(slice1, slice2 []string) []string {
	valueCount := make(map[string]int)

	// Count occurrences in the first slice
	for _, v := range slice1 {
		valueCount[v]++
	}

	// Count occurrences in the second slice
	for _, v := range slice2 {
		valueCount[v]++
	}

	// Collect the unique values
	var unique []string
	for value, count := range valueCount {
		if count == 1 {
			unique = append(unique, value)
		}
	}

	return unique
}

func (s *Socket) initMatchCall(pl *models.Data, sendCh chan *models.Outgoing) {
	callID := s.generateID()

	og := &models.Outgoing{
		OutPayload: models.Payload{
			Tag: "init-match-call",
			Data: models.Data{
				CallID: callID,
			},
		},
	}
	sendCh <- og

	usr, err := s.dbase.GetPotentialMatch(pl.SenderID)
	if err != nil {
		s.sLogger.Log.Errorln(err)
		return
	}

	usr.LastName = ""
	usr.Presence = pl.Presence
	usr.MutualInterest = pl.MutualInterest
	usr.ExclusiveInterest = uniqueValues(usr.Interests, usr.MutualInterest)

	// Init a new meet request
	mr := new(models.MeetRequest)

	mr.ID = s.generateID()
	mr.Status = "undefined"
	mr.Timestamp = time.Now().UTC()
	mr.CallID = callID
	mr.User = usr
	mr.Compliment = pl.Compliment
	mr.Rose = pl.Rose
	mr.SenderID = pl.SenderID
	mr.RecipientID = pl.RecipientID

	og = &models.Outgoing{
		OutPayload: models.Payload{
			Tag: "meet-request",
			Data: models.Data{
				MeetRequest: *mr,
			},
		},
	}

	// Asynchronously check if client is connected to this instance, then send payload to client.
	ackChan := make(chan bool)
	go s.ClientConnectedToThisInstance(pl.RecipientID, og, ackChan)
	if ok := <-ackChan; !ok {
		//handle checking and sending to other instance if user exist there.
		s.sLogger.Log.Infoln("user is not connected on this instance")
	}

	// Cache the meet request in redis.
	mr.SenderStatus = "undefined"
	mr.RecipientStatus = "undefined"

	if err = s.cbaseIf.CacheMeetRequest(mr); err != nil {
		s.sLogger.Log.Errorln(err)
		return
	}

	// store requested-to-meet relationship
	err = s.mbase.AddRequestedToMeetRelationship(mr)
	if err != nil {
		s.sLogger.Log.Errorln(err)
		return
	}
}

func (s *Socket) reInitMatchCall(pl *models.Data, sendCh chan *models.Outgoing) {
	callID := s.generateID()

	og := &models.Outgoing{
		OutPayload: models.Payload{
			Tag: "init-match-call",
			Data: models.Data{
				CallID: callID,
			},
		},
	}
	sendCh <- og

	usr, err := s.dbase.GetPotentialMatch(pl.MeetRequest.RecipientID)
	if err != nil {
		s.sLogger.Log.Errorln(err)
		return
	}

	pl.MeetRequest.User = usr
	pl.MeetRequest.CallID = callID

	og = &models.Outgoing{
		OutPayload: models.Payload{
			Tag: "return-meet-request",
			Data: models.Data{
				MeetRequest: pl.MeetRequest,
			},
		},
	}

	// Asynchronously check if client is connected to this instance, then send payload to client.
	ackChan := make(chan bool)
	go s.ClientConnectedToThisInstance(pl.MeetRequest.SenderID, og, ackChan)
	if ok := <-ackChan; !ok {
		//handle checking and sending to other instance if user exist there.
		s.sLogger.Log.Infoln("user is not connected on this instance")
	}

	pl.MeetRequest.RecipientStatus = "undefined"
	pl.MeetRequest.SenderStatus = "undefined"

	if err = s.cbaseIf.CacheMeetRequest(&pl.MeetRequest); err != nil {
		s.sLogger.Log.Errorln(err)
		return
	}
}

func (s *Socket) meetRequestStatusUpdate(pl *models.Data) {
	og := &models.Outgoing{
		OutPayload: models.Payload{
			Tag:  "meet-request-status-update",
			Data: *pl,
		},
	}

	// Asynchronously check if client is connected to this instance, then send payload to client.
	ackChan := make(chan bool)

	go s.ClientConnectedToThisInstance(pl.RecipientID, og, ackChan)
	if ok := <-ackChan; !ok {
		//handle checking and sending to other instance if user exist there.
		s.sLogger.Log.Infoln("user is not connected on this instance")
	}

	if pl.Status == "Meet request declined" {
		// add pass relationshp
		err := s.mbase.AddPassRelationship(time.Now().UTC(), pl.RecipientID, pl.SenderID)
		if err != nil {
			s.sLogger.Log.Errorln(err)
			return
		}

	}
}

func (s *Socket) matchStatusUpdate(pl *models.Data) {
	if pl.Status == "unmatch" {

		//Remove user from chat? No - remove user from recipients and add to non recipients
		err := s.dbase.RemoveUserFromChat(pl.ID, pl.SenderID)
		if err != nil {
			s.sLogger.Log.Errorln(err)
			return
		}

		//Update chat with unmatch note
		if pl.Note != "" {
			unmatchMessage := models.Message{ID: s.generateID(), ChatID: pl.ID, SenderID: pl.SenderID, RecieverID: pl.RecipientID, Content: pl.Note, Type: "note", Status: "sent", Timestamp: time.Now().UTC()}
			fmt.Println(unmatchMessage)

			err = s.dbase.AddMessage(&unmatchMessage)
			if err != nil {
				s.sLogger.Log.Errorln(err)
				return
			}
		}

		//update mbase
		err = s.mbase.AddUnmatchRelationship(time.Now().UTC(), pl.SenderID, pl.RecipientID)
		if err != nil {
			s.sLogger.Log.Errorln(err)
			return
		}
		return
	}

	if pl.SenderID == pl.UserID {
		//update senderstatus
		err := s.cbaseIf.UpdateMeetRequest(pl.CallID, "sender", pl.Status)
		if err != nil {
			s.sLogger.Log.Errorln(err)
			return
		}

	} else {
		err := s.cbaseIf.UpdateMeetRequest(pl.CallID, "recipient", pl.Status)
		if err != nil {
			s.sLogger.Log.Errorln(err)
			return
		}
	}

	mr, err := s.cbaseIf.RetrieveMeetRequest(pl.CallID)
	if err != nil {
		s.sLogger.Log.Errorln(err)
		return
	}

	if mr.RecipientStatus != "undefined" && mr.SenderStatus != "undefined" {
		// check if user match or not and emit match or pass
		if mr.RecipientStatus == "match" && mr.SenderStatus == "match" {
			s.sLogger.Log.Info("its a match")
			// emit its a match
			og := &models.Outgoing{
				OutPayload: models.Payload{
					Tag: "match-status-update",
					Data: models.Data{
						Status: "match",
					},
				},
			}

			// Asynchronously check if client is connected to this instance, then send payload to client.
			ackChan1 := make(chan bool)
			go s.ClientConnectedToThisInstance(pl.RecipientID, og, ackChan1)
			if ok := <-ackChan1; !ok {

				//~handle checking and sending to other instance if user exist there.
				s.sLogger.Log.Infoln("user is not connected on this instance")
			}

			ackChan2 := make(chan bool)
			go s.ClientConnectedToThisInstance(pl.SenderID, og, ackChan2)
			if ok := <-ackChan2; !ok {
				//~handle checking and sending to other instance if user exist there.
				s.sLogger.Log.Infoln("user is not connected on this instance")
			}

			// add match relationship
			err = s.mbase.AddMatchRelationship(time.Now().UTC(), mr.SenderID, mr.RecipientID)
			if err != nil {
				s.sLogger.Log.Errorln(err)
				return
			}

			//! create chat session between users and emit to them
			chat := new(models.Chat)
			chat.ID = s.generateID()
			chat.Recipients = append(chat.Recipients, models.User{ID: pl.SenderID}, models.User{ID: pl.RecipientID})
			chat.Messages = []models.Message{}
			chat.MatchedAt = time.Now().UTC()
			chat.NonRecipient = models.User{}

			//persist chat in db
			err = s.dbase.AddChat(chat)
			if err != nil {
				s.sLogger.Log.Errorln(err)
			}
		} else {
			s.sLogger.Log.Info("its a pass")
			// emit its a pass
			og := &models.Outgoing{
				OutPayload: models.Payload{
					Tag: "match-status-update",
					Data: models.Data{
						Status: "pass",
					},
				},
			}

			// Asynchronously check if client is connected to this instance, then send payload to client.
			ackChan1 := make(chan bool)
			go s.ClientConnectedToThisInstance(pl.RecipientID, og, ackChan1)
			if ok := <-ackChan1; !ok {
				//~handle checking and sending to other instance if user exist there.
				s.sLogger.Log.Infoln("user is not connected on this instance")
			}

			ackChan2 := make(chan bool)
			go s.ClientConnectedToThisInstance(pl.SenderID, og, ackChan2)
			if ok := <-ackChan2; !ok {
				//~handle checking and sending to other instance if user exist there.
				s.sLogger.Log.Infoln("user is not connected on this instance")
			}

			// add pass relationshp
			err = s.mbase.AddPassRelationship(time.Now().UTC(), mr.SenderID, mr.RecipientID)
			if err != nil {
				s.sLogger.Log.Errorln(err)
				return
			}
		}
	}
}

func (s *Socket) passStatusUpdate(pl *models.Data) {
	// add pass relationshp
	err := s.mbase.AddPassRelationship(time.Now().UTC(), pl.SenderID, pl.RecipientID)
	if err != nil {
		s.sLogger.Log.Errorln(err)
		return
	}
}

/**
Things to save to redis
- meet request.
- Users (key: id, cached chat id []string, instance connected to id string)
- Chat session
**/

// ~ Chat.
func (s *Socket) newMessage(pl *models.Data) {
	pl.Message.Status = "Sent"
	pl.Message.Type = "chat"

	og := &models.Outgoing{
		OutPayload: models.Payload{
			Tag: "new-message",
			Data: models.Data{
				Message: pl.Message,
			},
		},
	}

	// Asynchronously check if client is connected to this instance, then send payload to client.
	ackChan := make(chan bool)
	go s.ClientConnectedToThisInstance(pl.Message.RecieverID, og, ackChan)
	if ok := <-ackChan; !ok {

		//~handle checking and sending to other instance if user exist there.
		s.sLogger.Log.Infoln("user is not connected on this instance")
		//! Synchronously check if client is connected to other instance, then send payload to client.

		//persist
		err := s.dbase.AddMessage(&pl.Message)
		if err != nil {
			s.sLogger.Log.Errorln(err)
			return
		}
	} else {
		// check if chat is cached of this is the first message.
		ok, err := s.cbaseIf.ChatExist(pl.Message.ChatID)
		if err != nil {
			s.sLogger.Log.Errorln(err)
			return
		}

		// if it exist cache the message to the chat
		if ok {
			err = s.cbaseIf.CacheMessage(pl.Message)
			if err != nil {
				s.sLogger.Log.Errorln(err)
				return
			}
		} else {
			//else add the message to the chat and cache it.
			chat := new(models.ChatCache)
			chat.ChatID = pl.Message.ChatID
			chat.Messages = []models.Message{pl.Message}

			err = s.cbaseIf.CacheChat(*chat)
			if err != nil {
				s.sLogger.Log.Errorln(err)
				return
			}

			// add to both recipients incase one goes offline and comes back they can retries all chat from mongo.
			// Add chat ID to client (receiver) cachedChat slice.
			if err := s.cbaseIf.UpdateClientCachedChatSlice(pl.Message.RecieverID, pl.Message.ChatID); err != nil {
				s.sLogger.Log.Errorln(err)
			}

			// Add chat ID to client (sender) cachedChat slice.
			if err := s.cbaseIf.UpdateClientCachedChatSlice(pl.Message.SenderID, pl.Message.ChatID); err != nil {
				s.sLogger.Log.Errorln(err)
			}
		}
	}
}

func (s *Socket) updateMessageStatus(pl *models.Data) {
	fmt.Println("ids: ", pl.IDs)

	// check if chat is cached of this is the first message.
	ok, err := s.cbaseIf.ChatExist(pl.ID)
	if err != nil {
		s.sLogger.Log.Errorln(err)
		return
	}

	if ok {
		err := s.cbaseIf.UpdateMessageStatus(pl.ID, pl.IDs, "Read")
		if err != nil {
			s.sLogger.Log.Errorln(err)
			return
		}

	} else {
		// persist the status to database
		err := s.dbase.UpdateMessageStatus(pl.ID, pl.IDs)
		if err != nil {
			s.sLogger.Log.Errorln(err)
			return
		}
	}
}

// ~ Report.
func (s *Socket) report(pl *models.Data) {
	pl.Report.ID = s.generateID()
	pl.Report.Status = "Pending"
	pl.Report.Timestamp = time.Now().UTC()

	err := s.dbase.AddReport(pl.Report)
	if err != nil {
		s.sLogger.Log.Errorln(err)
		return
	}
}
