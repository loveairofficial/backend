package hook

import (
	"github.com/evalphobia/logrus_kinesis"
)

type Hook interface {
	GetHookOrigin() *logrus_kinesis.KinesisHook
}
