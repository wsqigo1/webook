package repository

import (
	"context"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"time"
)

type CronJobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, time time.Time) error
}

type PreemptJobRepository struct {
	dao dao.JobDAO
}

func (p *PreemptJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PreemptJobRepository) Release(ctx context.Context, jid int64) error {
	//TODO implement me
	panic("implement me")
}

func (p *PreemptJobRepository) UpdateUtime(ctx context.Context, id int64) error {
	//TODO implement me
	panic("implement me")
}

func (p *PreemptJobRepository) UpdateNextTime(ctx context.Context, id int64, time time.Time) error {
	//TODO implement me
	panic("implement me")
}

func NewPreemptJobRepository(dao dao.JobDAO) CronJobRepository {
	return &PreemptJobRepository{dao: dao}
}
