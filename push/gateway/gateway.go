package gateway

import (
	"loveair/push"
	"loveair/push/expo"
)

type PTYPE string

const (
	EXPO PTYPE = "expo"
)

func PConnect(options PTYPE) push.Interface {
	switch options {
	case EXPO:
		return expo.InitExpoPushInstance()
	}
	return nil
}
