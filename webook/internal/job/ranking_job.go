package job

import (
	"context"
	"github.com/gotomicro/redis-lock"
	"github.com/wsqigo/basic-go/webook/internal/service"
	"github.com/wsqigo/basic-go/webook/pkg/logger"
	"sync"
	"time"
)

type RankingJob struct {
	svc service.RankingService
	l   logger.LoggerV1
	// 一次运行超时的时间
	timeout time.Duration
	// 分布式锁
	client *rlock.Client
	key    string

	localLock *sync.Mutex
	lock      *rlock.Lock

	// 作业提示
	// 随机生成一个，就代表负载，你可以每隔一分钟生成一个
	load int32
}

func NewRankingJob(
	svc service.RankingService,
	l logger.LoggerV1,
	client *rlock.Client,
	timeout time.Duration) *RankingJob {
	return &RankingJob{
		svc:       svc,
		key:       "job:ranking",
		l:         l,
		client:    client,
		localLock: &sync.Mutex{},
		timeout:   timeout,
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

// go fun() { r.Run()}

func (r *RankingJob) Run() error {
	r.localLock.Lock()
	lock := r.lock
	if lock == nil {
		// 抢分布式锁
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()
		// 加锁本身，我们使用一个ctx
		// 本身我们这里设计的就是要在 r.timeout 内计算完成
		// 刚好也做成分布式锁的超时时间
		lock, err := r.client.Lock(ctx, r.key, r.timeout, // 锁的过期时间
			&rlock.FixIntervalRetry{
				// 每隔 100 ms 重试一次，每次重试的超时时间是 1s
				Interval: 100 * time.Millisecond,
				Max:      3,
				// 重试的超时
			}, time.Second)
		// 我们这里不需要处理 error，因为大部分情况下，可以相信别的节点会继续拿锁
		if err != nil {
			// 这边不需要返回 error，因为这时候可能是别的节点一直占着锁
			r.localLock.Unlock()
			r.l.Warn("获取分布式锁失败", logger.Error(err))
			return nil
		}
		r.lock = lock
		r.localLock.Unlock()
		go func() {
			// 并不是非得一半就续约
			er := lock.AutoRefresh(r.timeout/2, r.timeout)
			if er != nil {
				// 续约失败了
				// 你也没办法中断当下正在调度的热榜计算（如果有）
				r.localLock.Lock()
				r.lock = nil
				//lock.Unlock()
				r.localLock.Unlock()
			}
		}()
	}
	// 这边就是你拿到了锁
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	return r.svc.TopN(ctx)
}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}

//func (r *RankingJob) Run() error {
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
//	defer cancel()
//	lock, err := r.client.Lock(ctx, r.key, r.timeout,
//		&rlock.FixIntervalRetry{
//			Interval: time.Millisecond * 100,
//			Max:      3,
//			// 重试的超时
//		}, time.Second)
//	if err != nil {
//		return err
//	}
//	defer func() {
//		// 解锁
//		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//		defer cancel()
//		er := lock.Unlock(ctx)
//		if er != nil {
//			r.l.Error("ranking job释放分布式锁失败", logger.Error(er))
//		}
//	}()
//	ctx, cancel = context.WithTimeout(context.Background(), r.timeout)
//	defer cancel()
//
//	return r.svc.TopN(ctx)
//}
