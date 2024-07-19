package push

type Interface interface {
	SendPushNotification(title, body string, pIDs []string) error
}
