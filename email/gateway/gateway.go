package gateway

import (
	"loveair/email"
	"loveair/email/mailersend"
)

type ETYPE string

const (
	MAILERSEND ETYPE = "mailersend"
)

func EConnect(options ETYPE, Config map[string]string) email.Interface {
	switch options {
	case MAILERSEND:
		return mailersend.InitMailerSend(Config)
	}
	return nil
}
