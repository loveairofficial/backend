package redis

import "loveair/models"

func (r *Redis) ClientExist(clientID string) (bool, error) {
	ctx, cancel := getContext()
	defer cancel()

	exists, err := r.clientOne.Exists(ctx, clientID).Result()
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
	_, err := r.rhjOne.JSONSet(clientID, ".", cc)
	return err
}

func (r *Redis) RemoveClient(clientID string) error {
	_, err := r.rhjOne.JSONDel(clientID, ".")

	return err
}

func (r *Redis) UpdateClientCachedChatSlice(clientID, chatID string) error {
	_, err := r.rhjOne.JSONArrAppend(clientID, ".cachedChat", chatID)
	return err
}

func (r *Redis) GetClientCachedChatSlice(clientID string) (*[]string, error) {
	cc := new([]string)

	res, err := r.rhjOne.JSONGet(clientID, ".cachedChat")

	if err != nil {
		return nil, err
	}

	err = unMarshal(res.([]byte), &cc)
	if err != nil {
		return nil, err
	}

	return cc, nil
}