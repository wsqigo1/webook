package job

import (
	"context"
	"fmt"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"github.com/wsqigo/basic-go/webook/internal/service"
	"github.com/wsqigo/basic-go/webook/pkg/logger"
	"golang.org/x/sync/semaphore"
	"time"
)

// Executor 执行器，任务执行器
type Executor interface {
	Name() string
	// Exec ctx 这个是全局控制，Executor 的实现者注意要正确处理 ctx 超时或者取消
	Exec(ctx context.Context, j domain.Job) error
}

// LocalFuncExecutor 调用本地方法
type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{funcs: map[string]func(ctx context.Context, j domain.Job) error{}}
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) RegisterFunc(name string, fn func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = fn
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.Job) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return fmt.Errorf("未注册本地方法 %s", j.Name)
	}
	return fn(ctx, j)
}

type Scheduler struct {
	dbTimeout time.Duration

	svc       service.CronJobService
	executors map[string]Executor
	l         logger.LoggerV1

	limiter *semaphore.Weighted
}

func NewScheduler(svc service.CronJobService, l logger.LoggerV1) *Scheduler {
	return &Scheduler{
		dbTimeout: time.Second,
		svc:       svc,
		limiter:   semaphore.NewWeighted(100),
		l:         l,
		executors: map[string]Executor{},
	}
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.executors[exec.Name()] = exec
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	for {
		// 放弃调度了
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		// 抢占，获得可以运行的资格
		// 数据库查询的时候，超时时间是要短的
		dbCtx, cancel := context.WithTimeout(ctx, s.dbTimeout)
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			// 有 Error
			// 最简单的做法就是直接下一轮，也可以睡一段时间
			continue
		}

		// 肯定要调度指定 j
		exec, ok := s.executors[j.Executor]
		if !ok {
			// 你可以直接中断了，也可以下一轮
			s.l.Error("找不到执行器",
				logger.Int64("jid", j.Id),
				logger.String("executor", j.Executor))
			continue
		}

		// 要单独开一个 goroutine 来执行，这样我们就可以进入下一个调度
		go func() {
			defer func() {
				s.limiter.Release(1)
				// 这边要释放掉
				j.CancelFunc()
			}()
			err1 := exec.Exec(ctx, j)
			if err1 != nil {
				s.l.Error("执行任务失败",
					logger.Int64("jid", j.Id),
					logger.Error(err1))
				return
			}
			err1 = s.svc.ResetNextTime(ctx, j)
			if err1 != nil {
				s.l.Error("重置下次执行时间失败",
					logger.Int64("jid", j.Id),
					logger.Error(err1))
			}
		}()
	}
}
