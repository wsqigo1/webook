package service

import (
	"context"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"github.com/wsqigo/basic-go/webook/internal/repository"
	"github.com/wsqigo/basic-go/webook/pkg/logger"
	"time"
)

type CronJobService interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
	//Release(ctx context.Context, job domain.Job) error
	// 暴露 job 的增删改查方法
}

type cronJobService struct {
	repo            repository.CronJobRepository
	l               logger.LoggerV1
	refreshInterval time.Duration
}

func NewCronJobService(repo repository.CronJobRepository, l logger.LoggerV1) CronJobService {
	return &cronJobService{
		repo:            repo,
		l:               l,
		refreshInterval: time.Minute,
	}
}

func (c *cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := c.repo.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}

	ticker := time.NewTicker(c.refreshInterval)
	go func() {
		for range ticker.C {
			c.refresh(j.Id)
		}
	}()
	// 你抢占之后，你一直抢占着吗？
	// 你要考虑一个释放的问题
	j.CancelFunc = func() {
		// 自己在这里释放掉
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := c.repo.Release(ctx, j.Id)
		if err != nil {
			c.l.Error("释放 job 失败",
				logger.Error(err),
				logger.Int64("jid", j.Id))
		}
	}
	return j, nil
}

func (c *cronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	nextTime := j.NextTime()
	return c.repo.UpdateNextTime(ctx, j.Id, nextTime)
}

func (c *cronJobService) refresh(id int64) {
	// 本质上就是更新一下时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.repo.UpdateUtime(ctx, id)
	if err != nil {
		c.l.Error("续约失败", logger.Error(err),
			logger.Int64("jid", id))
	}
}
