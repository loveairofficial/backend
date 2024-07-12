package redis

import (
	"context"
	"encoding/json"
	"loveair/base/cache"
	"time"

	"github.com/nitishm/go-rejson/v4"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	remoteClient *redis.Client
	remoteRJH    *rejson.Handler

	localClient0 *redis.Client
	localClient1 *redis.Client
}

func InitRedisConnection(cacheConfig map[string]string) cache.Interface {
	// Cache meet request, chat conversation &  user instance data (3)
	remoteClient := redis.NewClient(&redis.Options{
		Network:    "tcp",
		Addr:       cacheConfig["remote_url"],
		Username:   cacheConfig["remote_username"], // no username set
		Password:   cacheConfig["remote_password"], // no password set
		DB:         0,                              // Redis has 16 logical db's in an instance.
		MaxRetries: 9000})

	remoteRJH := rejson.NewReJSONHandler()
	remoteRJH.SetGoRedisClientWithContext(context.Background(), remoteClient)

	// Cache potential match user data.
	localClient0 := redis.NewClient(&redis.Options{
		Network:    "tcp",
		Addr:       cacheConfig["local_url"],
		Username:   cacheConfig["local_username"], // no username set
		Password:   cacheConfig["local_password"], // no password set
		DB:         1,                             // Redis has 16 logical db's in an instance.
		MaxRetries: 9000})

	// Cache email pin data.
	localClient1 := redis.NewClient(&redis.Options{
		Network:    "tcp",
		Addr:       cacheConfig["local_url"],
		Username:   cacheConfig["local_username"], // no username set
		Password:   cacheConfig["local_password"], // no password set
		DB:         2,                             // Redis has 16 logical db's in an instance.
		MaxRetries: 9000})

	return &Redis{
		remoteClient,
		remoteRJH,
		localClient0,
		localClient1,
	}
}

// Creting context
func unMarshal(byt []byte, i interface{}) error {
	return json.Unmarshal(byt, i)
}

// Creting context
func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}
