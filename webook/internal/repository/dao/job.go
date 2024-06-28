package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type JobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, t time.Time) error
}

type GORMJobDAO struct {
	db *gorm.DB
}

func NewGORMJobDAO(db *gorm.DB) JobDAO {
	return &GORMJobDAO{db: db}
}

func (dao *GORMJobDAO) Preempt(ctx context.Context) (Job, error) {
	db := dao.db.WithContext(ctx)
	for {
		var j Job
		// 到了调度的时间
		now := time.Now().UnixMilli()
		// 作业：这里是缺少找到续约失败的 JOB 出来执行
		err := db.Where("status = ? AND next_time < ?", jobStatusWaiting, now).
			First(&j).Error
		if err != nil {
			return j, err
		}
		// 然后要开始抢占
		// 这里利用 version 来执行 CAS 操作
		// 其它一些公司可能会有一些 utime 之类的字段
		res := db.Model(&Job{}).
			Where("id = ? AND version = ?", j.Id, j.Version).
			Updates(map[string]any{
				"status":  jobStatusRunning,
				"version": j.Version + 1,
				"utime":   now,
			})
		if res.Error != nil {
			return Job{}, err
		}
		if res.RowsAffected == 0 {
			// 没抢到
			continue
		}
		return j, err
	}
}

func (dao *GORMJobDAO) Release(ctx context.Context, jid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", jid).Updates(map[string]any{
		"status": jobStatusWaiting,
		"utime":  now,
	}).Error
}

func (dao *GORMJobDAO) UpdateUtime(ctx context.Context, jid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", jid).
		Updates(map[string]any{
			"utime": now,
		}).Error
}

func (dao *GORMJobDAO) UpdateNextTime(ctx context.Context, id int64, t time.Time) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", id).Updates(map[string]any{
		"utime":     now,
		"next_time": t.UnixMilli(),
	}).Error
}

type Job struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Name       string `gorm:"type:varchar(128);unique"`
	Executor   string // 执行器
	Expression string
	Cfg        string
	// 状态来表达，是不是可以抢占，有没有被人抢占
	Status int

	Version  int
	NextTime int64 `gorm:"index"`
	Utime    int64
	Ctime    int64
}

const (
	// jobStatusWaiting 没人抢
	jobStatusWaiting = iota
	// jobStatusRunning 已经被人抢了
	jobStatusRunning
	// jobStatusPaused 不再需要调度了
	jobStatusPaused
)
