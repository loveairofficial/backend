package cache

import "loveair/models"

type Interface interface {
	CacheMeetRequest(*models.MeetRequest) error
	RetrieveMeetRequest(string) (*models.MeetRequest, error)
	UpdateMeetRequest(string, string, string) error
	DeleteCachedMeetRequest(id string) error

	ChatExist(string) (bool, error)
	CacheMessage(models.Message) error
	UpdateMessageStatus(string, []string, string) error
	CacheChat(models.ChatCache) error
	GetCachedChat(id string) (*models.ChatCache, error)
	DeleteCachedChat(id string) error

	ClientExist(string) (bool, error)
	AddClient(clientID string, cc models.ClientCache) error
	RemoveClient(clientID string) error
	UpdateClientCachedChatSlice(clientID, chatID string) error
	GetClientCachedChatSlice(clientID string) (*[]string, error)
}
