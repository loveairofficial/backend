package redis

import (
	"loveair/models"
	"time"
)

func (r *Redis) ClientExist(clientID string) (bool, error) {
	ctx, cancel := getContext()
	defer cancel()

	exists, err := r.remoteClient.Exists(ctx, clientID).Result()
	if err != nil {
		return false, err
	}

	if exists == 1 {
		return true, err
	} else {
		return false, err
	}
}

func (r *Redis) AddClient(clientID string, cc models.ClientCache) error {
	_, err := r.remoteRJH.JSONSet(clientID, ".", cc)

	if err == nil {
		ctx, cancel := getContext()
		defer cancel()
		_ = r.remoteClient.Expire(ctx, clientID, 24*time.Hour).Err()
	}
	return err
}

func (r *Redis) RemoveClient(clientID string) error {
	_, err := r.remoteRJH.JSONDel(clientID, ".")

	return err
}

func (r *Redis) UpdateClientCachedChatSlice(clientID, chatID string) error {
	_, err := r.remoteRJH.JSONArrAppend(clientID, ".cachedChat", chatID)
	return err
}

func (r *Redis) GetClientCachedChatSlice(clientID string) (*[]string, error) {
	cc := new([]string)

	res, err := r.remoteRJH.JSONGet(clientID, ".cachedChat")

	if err != nil {
		return nil, err
	}

	err = unMarshal(res.([]byte), &cc)
	if err != nil {
		return nil, err
	}

	return cc, nil
}
