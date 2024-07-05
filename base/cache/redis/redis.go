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
	client *redis.Client
	rhj    *rejson.Handler

	clientOne *redis.Client
	rhjOne    *rejson.Handler

	clientTwo *redis.Client
	rhjTwo    *rejson.Handler
}

func InitRedisConnection(cacheConfig map[string]string) cache.Interface {
	// Cache meet request data
	client := redis.NewClient(&redis.Options{
		Network:    "tcp",
		Addr:       cacheConfig["url"],
		Username:   cacheConfig["username"], // no username set
		Password:   cacheConfig["password"], // no password set
		DB:         0,                       // Redis has 16 logical db's in an instance.
		MaxRetries: 9000})

	rhj := rejson.NewReJSONHandler()
	rhj.SetGoRedisClientWithContext(context.Background(), client)

	// Cache client data
	clientOne := redis.NewClient(&redis.Options{
		Network:    "tcp",
		Addr:       cacheConfig["url"],
		Username:   cacheConfig["username"], // no username set
		Password:   cacheConfig["password"], // no password set
		DB:         1,                       // Redis has 16 logical db's in an instance.
		MaxRetries: 9000})

	rhjOne := rejson.NewReJSONHandler()
	rhjOne.SetGoRedisClientWithContext(context.Background(), clientOne)

	// Cache chat data
	clientTwo := redis.NewClient(&redis.Options{
		Network:    "tcp",
		Addr:       cacheConfig["url"],
		Username:   cacheConfig["username"], // no username set
		Password:   cacheConfig["password"], // no password set
		DB:         2,                       // Redis has 16 logical db's in an instance.
		MaxRetries: 9000})

	rhjTwo := rejson.NewReJSONHandler()
	rhjTwo.SetGoRedisClientWithContext(context.Background(), clientTwo)

	return &Redis{
		client,
		rhj,
		clientOne,
		rhjOne,
		clientTwo,
		rhjTwo,
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
