package cache

import (
	"context"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"time"
)

// 因为本身数据只有一份，所以不需要借
type RankingLocalCache struct {
	topN       *atomicx.Value[[]domain.Article]
	ddl        *atomicx.Value[time.Time]
	expiration time.Duration
}

func (r *RankingLocalCache) Set(ctx context.Context, arts []domain.Article) error {
	r.topN.Store(arts)
	r.ddl.Store(time.Now().Add(r.expiration))
	return nil
}

func (r *RankingLocalCache) Get(ctx context.Context) ([]domain.Article, error) {
	ddl := r.ddl.Load()
	arts := r.topN.Load()
	if len(arts) == 0 || ddl.Before(time.Now()) {
		return nil, errors.New("本地缓存失效了")
	}
	return arts, nil
}

func (r *RankingLocalCache) ForceGet(ctx context.Context) ([]domain.Article, error) {
	arts := r.topN.Load()
	if len(arts) == 0 {
		return nil, errors.New("本地缓存失效了")
	}
	return arts, nil
}
