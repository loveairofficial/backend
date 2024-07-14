package gateway

import (
	"loveair/email"
	"loveair/email/sendgrid"
)

type ETYPE string

const (
	MAILERSEND ETYPE = "mailersend"
	SENDGRID   ETYPE = "sendgrid"
)

func EConnect(options ETYPE, Config map[string]string) email.Interface {
	switch options {
	case SENDGRID:
		return sendgrid.InitSendGridDBInstance(Config)
	}
	return nil
}
