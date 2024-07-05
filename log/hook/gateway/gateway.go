package gateway

import (
	"loveair/log/hook"
	"loveair/log/hook/kinesis"
)

type HOOKTYPE string

const (
	AMAZONKINESIS HOOKTYPE = "amazonKinesis"
)

func ConnectHook(ht HOOKTYPE, hookConfig map[string]string) (hook.Hook, error) {
	switch ht {
	case AMAZONKINESIS:
		return kinesis.NewKinesisConnection(hookConfig)
	}
	return nil, nil
}
