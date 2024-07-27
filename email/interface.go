package email

type Interface interface {
	SendEmailVerificationPin(email, pin string) (int, error)
	SendWelcomeEmail(email, firstName string) (int, error)
	SendPasswordResetPin(email, pin string) (int, error)

	// SendWelcomeEmail(string, string) (int, error)
	// SendResetPinEmail(string, string) (int, error)
	// SendVerificationPinEmail(string, string) (int, error)
	// SendReferralRequestVerificationPinEmail(string, string) (int, error)

	//~ Admin
	SendAccountSuppressionEmail(email, firstName string) (int, error)
}
