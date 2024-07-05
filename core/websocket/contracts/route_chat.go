package contracts

type RouteChat struct {
	InstanceID string `json:"instanceid"`
	Chat       Chat   `json:"chat"`
}

type Chat struct {
	ID         string `json:"id,omitempty" bson:"id,omitempty"`
	Recipients []Account
	Messages   []Message
}

type Account struct {
	UID               string `json:"uid,omitempty" bson:"uid,omitempty"`
	FirstName         string `json:"first_name,omitempty" bson:"first_name"`
	ProfilePictureURL string `json:"profile_picture_URL" bson:"profile_picture_UR"`
}

func (c RouteChat) ContractName() string {
	return "route.chat"
}
