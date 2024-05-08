package limiter

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisSlidingWindowLimiter struct {
	cmd      redis.Cmdable
	interval time.Duration
	// 阈值
	rate int
}

//go:embed slide_window.lua
var luaScript string

func NewRedisSlidingWindowLimiter(cmd redis.Cmdable, interval time.Duration, rate int) *RedisSlidingWindowLimiter {
	return &RedisSlidingWindowLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}

func (l *RedisSlidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return l.cmd.Eval(ctx, luaScript, []string{key},
		l.interval.Milliseconds(),
		l.rate, time.Now().UnixMilli()).Bool()
}
