package repository

import (
	"context"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"github.com/wsqigo/basic-go/webook/internal/repository/dao"
	"gorm.io/gorm"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error
}

type CachedArticleRepository struct {
	// 操作单一的库
	dao dao.ArticleDAO

	// SyncV1 用
	readerDAO dao.ArticleReaderDAO
	authorDAO dao.ArticleAuthorDAO

	// SyncV2 用
	db *gorm.DB
}

func NewCachedArticleRepository(dao dao.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{
		dao: dao,
	}
}

func NewNewCachedArticleRepositoryV1(authorDAO dao.ArticleAuthorDAO,
	readerDAO dao.ArticleReaderDAO) *CachedArticleRepository {
	return &CachedArticleRepository{
		authorDAO: authorDAO,
		readerDAO: readerDAO,
	}
}

func NewNewCachedArticleRepositoryV2(
	authorDAO dao.ArticleAuthorDAO,
	readerDAO dao.ArticleReaderDAO) ArticleRepository {
	return &CachedArticleRepository{
		authorDAO: authorDAO,
		readerDAO: readerDAO,
	}
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error {
	return c.dao.SyncStatus(ctx, uid, id, status)
}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Sync(ctx, c.toEntity(art))
}

func (c *CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	artn := c.toEntity(art)

	var (
		id  = art.Id
		err error
	)

	if id > 0 {
		err = c.authorDAO.Update(ctx, artn)
	} else {
		id, err = c.authorDAO.Create(ctx, artn)
	}
	if err != nil {
		return 0, err
	}

	artn.Id = id
	err = c.readerDAO.Upsert(ctx, artn)
	return id, err
}

func (c *CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	// 防止后面业务panic
	// 直接 defer Rollback
	// 如果我们后续 Commit 了，这里会得到一个错误，但是没关系
	defer tx.Rollback()

	authorDAO := dao.NewGORMArticleAuthorDAO(tx)
	readerDAO := dao.NewGORMArticleReaderDAO(tx)

	artn := c.toEntity(art)

	var (
		id  = art.Id
		err error
	)

	if id > 0 {
		err = authorDAO.Update(ctx, artn)
	} else {
		id, err = authorDAO.Create(ctx, artn)
	}
	if err != nil {
		return 0, err
	}

	artn.Id = id
	err = readerDAO.UpsertV2(ctx, dao.PublishedArticle(artn))
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, nil
}

func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Insert(ctx, c.toEntity(art))
}

func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	return c.dao.UpdateById(ctx, c.toEntity(art))
}

func (c *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		//Status: uint8(art.Status),
		Status: art.Status.ToUint8(),
	}
}
