package email

type Interface interface {
	SendEmailVerificationPin(email, pin string) (int, error)

	// SendWelcomeEmail(string, string) (int, error)
	// SendResetPinEmail(string, string) (int, error)
	// SendVerificationPinEmail(string, string) (int, error)
	// SendReferralRequestVerificationPinEmail(string, string) (int, error)
}
