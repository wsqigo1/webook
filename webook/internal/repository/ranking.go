package repository

import (
	"context"
	"github.com/wsqigo/basic-go/webook/internal/domain"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}
