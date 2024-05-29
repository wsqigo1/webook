package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"github.com/wsqigo/basic-go/webook/internal/repository/cache"
	"github.com/wsqigo/basic-go/webook/internal/repository/dao"
	"github.com/wsqigo/basic-go/webook/pkg/logger"
	"gorm.io/gorm"
	"time"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetByID(ctx context.Context, id int64) (domain.Article, error)
	GetPubByID(ctx context.Context, id int64) (domain.Article, error)
}

type CachedArticleRepository struct {
	// 操作单一的库
	dao   dao.ArticleDAO
	cache cache.ArticleCache
	// 因为如果你直接访问 UserDAO，你就绕开了 repository，
	// repository 一般都有一些缓存机制
	userRepo UserRepository
	l        logger.LoggerV1

	// SyncV1 用
	readerDAO dao.ArticleReaderDAO
	authorDAO dao.ArticleAuthorDAO

	// SyncV2 用
	db *gorm.DB
}

func NewCachedArticleRepository(dao dao.ArticleDAO,
	userRepo UserRepository,
	cache cache.ArticleCache,
	l logger.LoggerV1) ArticleRepository {
	return &CachedArticleRepository{
		dao:      dao,
		cache:    cache,
		userRepo: userRepo,
		l:        l,
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

func (c *CachedArticleRepository) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 首先第一步，判定要不要查询缓存
	// 事实上，limit <= 100 都可以查询缓存
	if offset == 0 && limit == 100 {
		//if offset == 0 && limit <= 100
		res, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			return res, err
		} else {
			// 要考虑记录日志
			// 缓存未命中，你是可以忽略的
		}
	}
	arts, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	res := slice.Map[dao.Article, domain.Article](arts, func(idx int, src dao.Article) domain.Article {
		return c.toDomain(src)
	})

	go func() {
		if offset == 0 && limit == 100 {
			// 缓存回写失败，不一定是大问题，但有可能是大问题
			err = c.cache.SetFirstPage(ctx, uid, res)
			if err != nil {
				// 记录日志
				// 我需要监控这里
			}
		}
	}()

	// 一般都是让调用者来控制是否异步。
	go func() {
		c.preCache(ctx, res)
	}()

	return res, nil
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error {
	err := c.dao.SyncStatus(ctx, uid, id, status)
	if err == nil {
		er := c.cache.DelFirstPage(ctx, uid)
		if er != nil {
			// 也要记录日志
		}
	}

	return err
}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
		}
	}
	// 在这里尝试，设置缓存
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 你可以灵活设置过期时间
		user, er := c.userRepo.FindById(ctx, art.Author.Id)
		if er != nil {
			// 要记录日志
			return
		}
		art.Author = domain.Author{
			Id:   user.Id,
			Name: user.Nickname,
		}
		er = c.cache.SetPub(ctx, art)
		if er != nil {
			// 记录日志
		}
	}()
	return id, err
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
	res, err := c.dao.Insert(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
		}
	}
	return res, nil
}

func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	err := c.dao.UpdateById(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
		}
	}
	return err
}

func (c *CachedArticleRepository) GetByID(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.Get(ctx, id)
	if err != nil {
		return res, err
	}
	art, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = c.toDomain(art)
	go func() {
		er := c.cache.Set(ctx, res)
		if er != nil {
			// 记录日志
		}
	}()
	return res, nil
}

func (c *CachedArticleRepository) GetPubByID(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.GetPub(ctx, id)
	if err == nil {
		return res, err
	} else {
		// 要考虑记录日志
		// 缓存未命中，你是可以忽略的
	}
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	// 我现在要去查对应的 User 信息，拿到创作者信息
	res = c.toDomain(dao.Article(art))
	author, err := c.userRepo.FindById(ctx, art.AuthorId)
	if err != nil {
		return domain.Article{}, err
		// 要额外记录日志，因为你吞掉了错误信息
		//return res, nil
	}
	res.Author.Name = author.Nickname
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := c.cache.SetPub(ctx, res)
		if er != nil {
			// 记录日志
		}
	}()
	return res, nil
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

func (c *CachedArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			// 这里有一个错误
			Id: art.AuthorId,
		},
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
		Status: domain.ArticleStatus(art.Status),
	}
}

func (c *CachedArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	// 1MB
	const size = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) < size {
		err := c.cache.Set(ctx, arts[0])
		if err != nil {
			// 记录日志
			c.l.Error("提前准备缓存失败", logger.Error(err))
		}
	}
}
