package kinesis

import (
	"loveair/log/hook"

	"github.com/evalphobia/logrus_kinesis"
	"github.com/sirupsen/logrus"
)

type Kinesis struct {
	driver *logrus_kinesis.KinesisHook
}

func NewKinesisConnection(hookConfig map[string]string) (hook.Hook, error) {
	hook, err := logrus_kinesis.New("outspire", logrus_kinesis.Config{
		AccessKey: hookConfig["AccessKey"], // AWS accessKeyId
		SecretKey: hookConfig["SecretKey"], // AWS secretAccessKey
		Region:    hookConfig["Region"],
	})

	if err != nil {
		return nil, err
	}

	// ignore field
	// hook.AddIgnore("context")

	// add custome filter
	// hook.AddFilter("error", logrus_kinesis.FilterError)

	hook.SetLevels([]logrus.Level{6})

	return &Kinesis{
		driver: hook,
	}, nil
}

func (k *Kinesis) GetHookOrigin() *logrus_kinesis.KinesisHook {
	return k.driver
}
