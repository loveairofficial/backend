package redis

import (
	"fmt"
	"loveair/models"
	"time"
)

func (r *Redis) CacheMeetRequest(mr *models.MeetRequest) error {
	res, err := r.remoteRJH.JSONSet(mr.CallID, ".", mr)

	if err == nil {
		ctx, cancel := getContext()
		defer cancel()
		_ = r.remoteClient.Expire(ctx, mr.CallID, 3*time.Hour).Err()
	}

	fmt.Println("redis ", res, err)
	return err
}

func (r *Redis) RetrieveMeetRequest(cid string) (*models.MeetRequest, error) {
	mr := new(models.MeetRequest)
	res, err := r.remoteRJH.JSONGet(cid, ".")
	if err != nil {
		return nil, err
	}

	err = unMarshal(res.([]byte), mr)
	if err != nil {
		return nil, err
	}

	return mr, nil
}

func (r *Redis) UpdateMeetRequest(callID, who, status string) error {
	if who == "sender" {
		_, err := r.remoteRJH.JSONSet(callID, ".senderStatus", status)
		return err
	} else {
		_, err := r.remoteRJH.JSONSet(callID, ".recipientStatus", status)
		return err
	}
}

func (r *Redis) DeleteCachedMeetRequest(id string) error {
	_, err := r.remoteRJH.JSONDel(id, ".")
	if err != nil {
		return err
	}

	return nil
}
