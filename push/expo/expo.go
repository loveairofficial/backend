package expo

import (
	"loveair/push"

	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
)

type ExpoPush struct {
	client *expo.PushClient
}

func InitExpoPushInstance() push.Interface {
	return &ExpoPush{
		client: expo.NewPushClient(nil),
	}
}

func (ep *ExpoPush) SendPushNotification(title, body string, pIDs []string) error {
	// Publish message
	var err error
	var pushTkns []expo.ExponentPushToken

	for _, pID := range pIDs {
		pushTkns = append(pushTkns, expo.ExponentPushToken(pID))
	}

	_, err = ep.client.Publish(
		&expo.PushMessage{
			To:       pushTkns,
			Title:    "Loveair",
			Body:     "New message",
			Data:     map[string]string{"withSome": "data"},
			Sound:    "default",
			Priority: expo.DefaultPriority,
		},
	)

	return err
}

// func (sg *SendGrid) SendEmailVerificationPin(email, pin string) (int, error) {
// 	newMail := mail.NewV3Mail()
// 	newEmail := mail.NewEmail("Loveair", "no_reply@loveair.co")
// 	newMail.SetFrom(newEmail)

// 	//Set email template ID
// 	newMail.SetTemplateID("d-cfa717aca3bb45899f243ee2e83b014e")

// 	//Create new Personalization
// 	p := mail.NewPersonalization()
// 	tos := []*mail.Email{
// 		mail.NewEmail("", email),
// 	}
// 	p.AddTos(tos...)

// 	//Set dynamic email template data.
// 	p.SetDynamicTemplateData("pin", pin)

// 	newMail.AddPersonalizations(p)

// 	// check for responce status code 202
// 	res, err := sg.client.Send(newMail)

// 	return res.StatusCode, err
// }
