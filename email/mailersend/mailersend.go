package mailersend

import (
	"context"
	"loveair/email"
	"time"

	"github.com/mailersend/mailersend-go"
)

type MailerSend struct {
	client *mailersend.Mailersend
	// senderName  string
	// senderEmail string
}

func InitMailerSend(Config map[string]string) email.Interface {
	return &MailerSend{
		client: mailersend.NewMailersend(Config["API_KEY"]),
		// senderName:  Config["SENDER_NAME"],
		// senderEmail: Config["SENDER_EMAIL"],
	}
}

// Creting context
func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func (ms *MailerSend) SendEmailVerificationPin(email, pin string) (int, error) {
	ctx, cancel := getContext()
	defer cancel()

	recipients := []mailersend.Recipient{
		{
			Name:  "",
			Email: email,
		},
	}

	personalization := []mailersend.Personalization{
		{
			Email: "recipient@email.com",
			Data: map[string]interface{}{
				"pin": pin,
			},
		},
	}

	message := ms.client.Email.NewMessage()
	message.SetRecipients(recipients)
	message.SetTemplateID("k68zxl2m7y54j905")
	message.SetPersonalization(personalization)

	res, err := ms.client.Email.Send(ctx, message)

	return res.StatusCode, err
}
