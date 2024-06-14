package limiter

import "context"

//go:generate mockgen -destination=./mocks/limiter.mock.go -package=limitermocks -source=./types.go Limiter
type Limiter interface {
	// Limit 是否触发限流
	// 返回 true，就是触发限流
	Limit(ctx context.Context, key string) (bool, error)
}
