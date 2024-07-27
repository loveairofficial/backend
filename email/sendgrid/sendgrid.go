package sendgrid

import (
	"loveair/email"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGrid struct {
	client *sendgrid.Client
}

func InitSendGridDBInstance(Config map[string]string) email.Interface {
	return &SendGrid{
		client: sendgrid.NewSendClient(Config["API_KEY"]),
	}
}

//~ Client

func (sg *SendGrid) SendEmailVerificationPin(email, pin string) (int, error) {
	newMail := mail.NewV3Mail()
	newEmail := mail.NewEmail("Loveair", "no_reply@loveair.co")
	newMail.SetFrom(newEmail)

	//Set email template ID
	newMail.SetTemplateID("d-cfa717aca3bb45899f243ee2e83b014e")

	//Create new Personalization
	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail("", email),
	}
	p.AddTos(tos...)

	//Set dynamic email template data.
	p.SetDynamicTemplateData("pin", pin)

	newMail.AddPersonalizations(p)

	// check for responce status code 202
	res, err := sg.client.Send(newMail)

	return res.StatusCode, err
}

func (sg *SendGrid) SendWelcomeEmail(email, firstName string) (int, error) {
	newMail := mail.NewV3Mail()
	newEmail := mail.NewEmail("Loveair", "no_reply@loveair.co")
	newMail.SetFrom(newEmail)

	//Set email template ID
	newMail.SetTemplateID("d-5065f5ee8ee14b39bf2090a9df92788a")

	//Create new Personalization
	p := mail.NewPersonalization()

	tos := []*mail.Email{
		mail.NewEmail(firstName, email),
	}
	p.AddTos(tos...)

	//Set dynamic email template data.
	p.SetDynamicTemplateData("name", firstName)

	newMail.AddPersonalizations(p)

	// check for responce status code 202
	res, err := sg.client.Send(newMail)

	return res.StatusCode, err
}

func (sg *SendGrid) SendPasswordResetPin(email, pin string) (int, error) {
	newMail := mail.NewV3Mail()
	newEmail := mail.NewEmail("Loveair", "no_reply@loveair.co")
	newMail.SetFrom(newEmail)

	//Set email template ID
	newMail.SetTemplateID("d-2b5a5a73cb1344249cdafa70f6cb5396")

	//Create new Personalization
	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail("", email),
	}
	p.AddTos(tos...)

	//Set dynamic email template data.
	p.SetDynamicTemplateData("pin", pin)

	newMail.AddPersonalizations(p)

	// check for responce status code 202
	res, err := sg.client.Send(newMail)

	return res.StatusCode, err
}

// ~ Admin
func (sg *SendGrid) SendAccountSuppressionEmail(email, firstName string) (int, error) {
	newMail := mail.NewV3Mail()
	newEmail := mail.NewEmail("Loveair", "no_reply@loveair.co")
	newMail.SetFrom(newEmail)

	//Set email template ID
	newMail.SetTemplateID("d-c5d296f77c76415d8c2cfdf7deeb288a")

	//Create new Personalization
	p := mail.NewPersonalization()

	tos := []*mail.Email{
		mail.NewEmail(firstName, email),
	}
	p.AddTos(tos...)

	//Set dynamic email template data.
	p.SetDynamicTemplateData("name", firstName)

	newMail.AddPersonalizations(p)

	// check for responce status code 202
	res, err := sg.client.Send(newMail)

	return res.StatusCode, err
}
