package redis

import (
	"fmt"
	"loveair/models"
)

func (r *Redis) CacheMeetRequest(mr *models.MeetRequest) error {
	res, err := r.rhj.JSONSet(mr.CallID, ".", mr)
	fmt.Println("redis ", res, err)
	return err
}

func (r *Redis) RetrieveMeetRequest(cid string) (*models.MeetRequest, error) {
	mr := new(models.MeetRequest)
	res, err := r.rhj.JSONGet(cid, ".")
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
		_, err := r.rhj.JSONSet(callID, ".senderStatus", status)
		return err
	} else {
		_, err := r.rhj.JSONSet(callID, ".recipientStatus", status)
		return err
	}
}

func (r *Redis) DeleteCachedMeetRequest(id string) error {
	_, err := r.rhj.JSONDel(id, ".")
	if err != nil {
		return err
	}

	return nil
}
