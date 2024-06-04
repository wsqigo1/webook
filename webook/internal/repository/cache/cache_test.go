package cache

import (
	"github.com/redis/go-redis/v9"
	"testing"
)

func TestCache_Set(t *testing.T) {
	// 测试set
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 服务器地址
		Password: "",               // Redis 访问密码，如果没有则留空
		DB:       0,                // 选择的数据库
	})
	cacheService := NewInteractiveRedisCache(client)
	biz := "article"

}
