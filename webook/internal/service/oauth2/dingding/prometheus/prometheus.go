package prometheus

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"github.com/wsqigo/basic-go/webook/internal/service/oauth2/dingding"
	"time"
)

type Decorator struct {
	dingding.Service
	sum prometheus.Summary
}

func NewDecorator(svc dingding.Service, sum prometheus.Summary) *Decorator {
	return &Decorator{
		Service: svc,
		sum:     sum,
	}
}

func (d *Decorator) VerifyCode(ctx context.Context, code string) (domain.DDingInfo, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		d.sum.Observe(float64(duration))
	}()
	return d.Service.VerifyCode(ctx, code)
}
