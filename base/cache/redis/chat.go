package redis

import (
	"fmt"
	"loveair/models"
)

func (r *Redis) ChatExist(chatID string) (bool, error) {
	ctx, cancel := getContext()
	defer cancel()

	exists, err := r.remoteClient.Exists(ctx, chatID).Result()
	if err != nil {
		return false, err
	}

	if exists == 1 {
		return true, err
	} else {
		return false, err
	}
}

func (r *Redis) CacheMessage(msg models.Message) error {
	_, err := r.remoteRJH.JSONArrAppend(msg.ChatID, ".messages", msg)
	return err
}

// UpdateMessageStatus updates the status of messages with given IDs to "delivered" directly in Redis
func (r *Redis) UpdateMessageStatus(chatID string, messageIDs []string, status string) error {
	for _, messageID := range messageIDs {
		path := fmt.Sprintf(".messages[?(@.id=='%s')].status", messageID)
		_, err := r.remoteRJH.JSONSet(chatID, path, status)

		if err != nil {
			return fmt.Errorf("failed to update message status for ID %s: %v", messageID, err)
		}
	}

	return nil
}

func (r *Redis) CacheChat(cc models.ChatCache) error {
	// _, err := r.rhjTwo.JSONArrAppend(msg.ChatID, ".messages", msg)
	_, err := r.remoteRJH.JSONSet(cc.ChatID, ".", cc)
	return err
}

func (r *Redis) GetCachedChat(id string) (*models.ChatCache, error) {
	cc := new(models.ChatCache)

	res, err := r.remoteRJH.JSONGet(id, ".")
	if err != nil {
		return nil, err
	}

	err = unMarshal(res.([]byte), &cc)
	if err != nil {
		return nil, err
	}

	return cc, nil
}

func (r *Redis) DeleteCachedChat(id string) error {
	_, err := r.remoteRJH.JSONDel(id, ".")
	if err != nil {
		return err
	}

	return nil
}
