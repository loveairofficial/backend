package contracts

type UpdateClient struct {
	ChatID     string `json:"sessionid,omitempty" bson:"sessionid,omitempty"`
	RecieverID string `json:"recipientid,omitempty" bson:"recipientid,omitempty"`
}

func (uc UpdateClient) ContractName() string {
	return "update.client"
}
