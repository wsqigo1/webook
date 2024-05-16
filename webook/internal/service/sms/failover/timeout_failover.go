package failover

import (
	"context"
	"github.com/wsqigo/basic-go/webook/internal/service/sms"
	"sync/atomic"
)

type TimeoutFailoverSMSService struct {
	svcs []sms.Service
	// 当前正在使用节点
	idx int32
	// 连续几个超时了
	cnt int32
	// 切换的阈值，只读的
	threshold int32
}

func NewTimeoutFailoverSMSService(svcs []sms.Service, threshold int32) *TimeoutFailoverSMSService {
	return &TimeoutFailoverSMSService{
		svcs:      svcs,
		threshold: threshold,
	}
}

// 不是非常严格的“连续 N 个超时就切换“
func (t *TimeoutFailoverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)
	// 超过阈值，执行切换
	if cnt >= t.threshold {
		// 触发切换，计算新的下标
		newIdx := (idx + 1) % int32(len(t.svcs))
		// 可能有两个 goroutine 都进来
		// CAS 操作失败，说明有人切换了，所以你这里不需要检测返回值
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			// 充值这个 cnt 计数
			atomic.StoreInt32(&t.cnt, 0)
		}
		idx = newIdx
	}

	svc := t.svcs[idx]
	err := svc.Send(ctx, tplId, args, numbers...)
	switch err {
	case nil:
		// 连续超时，所以不超时的时候要充值到 0
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	case context.DeadlineExceeded:
		atomic.AddInt32(&t.cnt, 1)
	default:
		// 遇到了错误，但是又不是超时错误，这个时候，你要考虑怎么搞
		// 我可以增加，也可以不增加
		// 如果强调一定是超时，那么就不增加
		// 如果是 EOF 之类的错误，你还可以考虑直接切换
	}
	return err
}
