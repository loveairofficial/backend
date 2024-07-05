package models

type ChatCache struct {
	ChatID   string    `json:"chatid"`
	Messages []Message `json:"messages"`
}

type ClientCache struct {
	InstanceID string   `json:"instanceid"`
	CachedChat []string `json:"cachedChat"`
}
