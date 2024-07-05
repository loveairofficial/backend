package gateway

import (
	"loveair/base/cache"
	"loveair/base/cache/redis"
)

type CBTYPE string

const (
	REDIS    CBTYPE = "redis"
	MEMCACHE CBTYPE = "memcache"
)

func ConnectCache(options CBTYPE, cacheConfig map[string]string) cache.Interface {
	switch options {
	case REDIS:
		return redis.InitRedisConnection(cacheConfig)
	case MEMCACHE:
		// return memcache.NewMemCacheConnection(connection)
	}
	return nil
}
