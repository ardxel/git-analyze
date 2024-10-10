package api

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"git-analyzer/pkg/config"
	"log"
	"time"

	"github.com/go-redis/cache/v9"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

type RedisDB struct {
	ctx         context.Context
	client      *redis.Client
	rateLimiter *redis_rate.Limiter
	cache       *cache.Cache
}

func CreateRedisDB() *RedisDB {
	ctx := context.Background()
	addr := fmt.Sprintf("%s:%s", config.Vars.RedisHost, config.Vars.RedisPort)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	_, err := client.Ping(ctx).Result()

	if err != nil {
		panic(err)
	}

	rateLimiter := redis_rate.NewLimiter(client)

	cache := cache.New(&cache.Options{
		Redis:      client,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	return &RedisDB{
		client:      client,
		rateLimiter: rateLimiter,
		cache:       cache,
		ctx:         ctx,
	}
}

func (r *RedisDB) RateLimitAllow(key string) *redis_rate.Result {
	res, err := r.rateLimiter.Allow(r.ctx, key, redis_rate.PerMinute(2))

	if err != nil {
		panic(err)
	}

	return res
}

func (r *RedisDB) GetCache(key string) (interface{}, bool) {
	var value interface{}
	err := r.cache.Get(r.ctx, key, &value)

	if err == cache.ErrCacheMiss {
		return nil, false
	} else if err != nil {
		log.Printf("Error fetching from Redis: %v", err)
		return nil, false
	}

	return value, true
}

func (r *RedisDB) SetCache(key string, value interface{}) {
	err := r.cache.Set(&cache.Item{
		Ctx:   r.ctx,
		Key:   key,
		Value: value,
		TTL:   time.Minute * 5,
	})

	if err != nil {
		panic(err)
	}
}

func RepoTaskResultKey(repoOwner, repoName string) (string, bool) {
	if repoOwner == "" || repoName == "" {
		return "", false
	}

	str := fmt.Sprintf("%s:%s", repoOwner, repoName)
	hash := md5.Sum([]byte(str))

	return hex.EncodeToString(hash[:]), true
}
