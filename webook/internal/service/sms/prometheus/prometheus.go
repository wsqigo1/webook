package prometheus

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/wsqigo/basic-go/webook/internal/service/sms"
	"time"
)

type Decorator struct {
	svc    sms.Service
	vector *prometheus.SummaryVec
}

func NewDecorator(svc sms.Service, opt prometheus.SummaryOpts) *Decorator {
	return &Decorator{
		svc:    svc,
		vector: prometheus.NewSummaryVec(opt, []string{"tpl_id"}),
	}
}

func (d *Decorator) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		d.vector.WithLabelValues("tpl_id").Observe(float64(duration))
	}()
	return d.svc.Send(ctx, tplId, args, numbers...)
}
