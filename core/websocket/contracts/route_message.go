package contracts

import "time"

type RouteMessage struct {
	InstanceID string  `json:"instanceid"`
	Message    Message `json:"directMessage"`
}

type Message struct {
	ID         string    `json:"id,omitempty" bson:"id,omitempty"`
	ChatID     string    `json:"sessionid,omitempty" bson:"sessionid,omitempty"`
	RecieverID string    `json:"recipientid,omitempty" bson:"recipientid,omitempty"`
	SenderID   string    `json:"senderid,omitempty" bson:"senderid,omitempty"`
	Content    string    `json:"content,omitempty" bson:"content,omitempty"`
	Timestamp  time.Time `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
}

func (sm RouteMessage) ContractName() string {
	return "route.message"
}
