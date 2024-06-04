package service

import (
	"context"
	"errors"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"github.com/wsqigo/basic-go/webook/internal/events/article"
	"github.com/wsqigo/basic-go/webook/internal/repository"
	"github.com/wsqigo/basic-go/webook/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid int64, id int64) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubByID(ctx context.Context, id int64, uid int64) (domain.Article, error)
}

type articleService struct {
	repo     repository.ArticleRepository
	producer article.Producer

	// V1 写法专用
	authorRepo repository.ArticleAuthorRepository
	readerRepo repository.ArticleReaderRepository
	logger     logger.LoggerV1
}

func NewArticleService(repo repository.ArticleRepository,
	producer article.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		producer: producer,
	}
}

func NewArticleServiceV1(authorRepo repository.ArticleAuthorRepository,
	readerRepo repository.ArticleReaderRepository, logger logger.LoggerV1) *articleService {
	return &articleService{
		authorRepo: authorRepo,
		readerRepo: readerRepo,
		logger:     logger,
	}
}

func (a *articleService) Withdraw(ctx context.Context, uid int64, id int64) error {
	return a.repo.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate)
}

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	} else {
		return a.repo.Create(ctx, art)
	}
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, art)
}

func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	// 想到这里要先操作制作库
	var (
		id  = art.Id
		err error
	)

	// 这一段逻辑其实就是 Save
	if art.Id > 0 {
		err = a.authorRepo.Update(ctx, art)
	} else {
		id, err = a.authorRepo.Create(ctx, art)
	}
	if err != nil {
		return 0, err
	}

	// 这里操作线上库
	// 保持制作库和线上库的 ID 是一样的
	art.Id = id
	for i := 0; i < 3; i++ {
		// 我可能线上库已经有数据了
		// 也可能没有
		err = a.readerRepo.Save(ctx, art)
		if err != nil {
			a.logger.Error("保存到制作库成功但是到线上库失败",
				logger.Int64("art_id", art.Id),
				logger.Error(err))
			// 在接入了 metrics 或者 tracing 之后，
			// 这边要进一步记录必要的 DEBUG 信息。
		} else {
			return id, nil
		}
	}
	a.logger.Error("保存到线上库失败，重试次数耗尽",
		logger.Int64("art_id", art.Id),
		logger.Error(err))

	return id, errors.New("保存到线上库失败，重试次数耗尽")
}

func (a *articleService) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.GetByAuthor(ctx, uid, offset, limit)
}

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetByID(ctx, id)
}

func (a *articleService) GetPubByID(ctx context.Context, id, uid int64) (domain.Article, error) {
	res, err := a.repo.GetPubByID(ctx, id)
	go func() {
		if err == nil {
			// 在这里发一个消息
			er := a.producer.ProduceReadEvent(article.ReadEvent{
				Aid: id,
				Uid: uid,
			})
			if er != nil {
				a.logger.Error("发送 ReadEvent 失败",
					logger.Int64("aid", id),
					logger.Int64("uid", uid),
					logger.Error(err))
			}
		}
	}()
	return res, err
}
