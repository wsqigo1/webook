package repository

import (
	"context"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"github.com/wsqigo/basic-go/webook/internal/repository/dao"
)

type ArticleReaderRepository interface {
	// Save 有则更新，无则插入，也就是 insert or update 语句
	Save(ctx context.Context, art domain.Article) error
}

type CachedArticleReaderRepository struct {
	dao dao.ArticleReaderDAO
}
