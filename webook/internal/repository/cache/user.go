package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"time"
)

var ErrKeyNotExist = redis.Nil

//go:generate mockgen -destination=./mocks/user.mock.go -package=cachemocks -source=./user.go UserCache
type UserCache interface {
	Get(ctx context.Context, uid int64) (domain.User, error)
	Set(ctx context.Context, du domain.User) error
}

type RedisUserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func (c *RedisUserCache) Get(ctx context.Context, uid int64) (domain.User, error) {
	key := c.key(uid)
	// 假定这个地方用 JSON
	data, err := c.cmd.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal([]byte(data), &u)
	//if err != nil {
	//	return domain.User{}, err
	//}
	//return u, nil
	return u, err
}

func (c *RedisUserCache) Set(ctx context.Context, du domain.User) error {
	key := c.key(du.Id)
	data, err := json.Marshal(du)
	if err != nil {
		return err
	}

	return c.cmd.Set(ctx, key, data, c.expiration).Err()
}

func (c *RedisUserCache) key(uid int64) string {
	// user-info-
	// user.info.
	// user/info/
	// user_info_
	return fmt.Sprintf("user:info:%d", uid)
}

func NewUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,
		expiration: 15 * time.Minute,
	}
}

// 一定不要自己去初始化你需要的东西，让外面传进来
//func NewUserCacheV1(addr string) *UserCache {
//	client := redis.NewClient(&redis.Options{Addr: addr})
//	return &UserCache{
//		client:        client,
//		expiration: 15 * time.Minute,
//	}
//}
