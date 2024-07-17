package models

import (
	"time"

	"github.com/gorilla/websocket"
)

type Incomming struct {
	Mt        int
	InPayload []byte
}

type Client struct {
	ID     string
	Conn   *websocket.Conn
	SendCh chan *Outgoing
}

type Outgoing struct {
	Mt         int
	OutPayload Payload
}

type Payload struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Tag        string `json:"tag"`
	Data       Data   `json:"data"`
}

type Data struct {
	ID             string      `json:"id"`
	IDs            []string    `json:"ids"`
	CallID         string      `json:"callID"`
	UserID         string      `json:"userID"`
	RecipientID    string      `json:"recipientID"`
	SenderID       string      `json:"senderID"`
	Compliment     string      `json:"compliment"`
	Rose           bool        `json:"rose"`
	MutualInterest []string    `json:"mutualInterest"`
	Presence       string      `json:"presence"`
	MeetRequest    MeetRequest `json:"meetRequest"`
	Status         string      `json:"status"`
	Chat           Chat        `json:"chat"`
	Message        Message     `json:"message" bson:"message"`
	Note           string      `json:"note"`
	Report         Report      `json:"report"`
	Feedback       Feedback    `json:"feedback"`
}

type MeetRequest struct {
	ID              string    `json:"id"`
	Status          string    `json:"status"` //eg, Mtached, Passed, Undefined.
	Timestamp       time.Time `json:"timestamp"`
	CallID          string    `json:"callID"`
	Presence        string    `json:"presence"`
	LastSeen        time.Time `json:"lastSeen" bson:"last_seen"`
	User            User      `json:"user"`
	Compliment      string    `json:"compliment"`
	Rose            bool      `json:"rose"`
	SenderID        string    `json:"senderID"`
	RecipientID     string    `json:"recipientID"`
	SenderStatus    string    `json:"senderStatus"`
	RecipientStatus string    `json:"recipientStatus"`
}

type Chat struct {
	ID           string `json:"id" bson:"id"`
	Status       string `json:"status" bson:"status,"` // eg matched or unmatched
	Recipients   []User `json:"recipients," bson:"recipients,"`
	NonRecipient User   `json:"nonRecipients," bson:"non_recipients,"`

	Messages    []Message `json:"messages" bson:"messages"`
	MatchedAt   time.Time `json:"matchedAt" bson:"matched_at"`
	UnmatchedAt time.Time `json:"unmatchedAt" bson:"unmatched_at"`
}

type Message struct {
	ID         string    `json:"id" bson:"id,"`
	Type       string    `json:"type" bson:"type,"`
	ChatID     string    `json:"chatid" bson:"chat_id,"`
	Status     string    `json:"status" bson:"status,"`
	RecieverID string    `json:"receiverid" bson:"receiver_id,"`
	SenderID   string    `json:"senderid" bson:"sender_id,"`
	Content    string    `json:"content" bson:"content,"`
	Timestamp  time.Time `json:"timestamp" bson:"timestamp,"`
}
