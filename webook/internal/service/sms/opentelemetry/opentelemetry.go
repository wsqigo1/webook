package opentelemetry

import (
	"context"
	"github.com/wsqigo/basic-go/webook/internal/service/sms"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Decorator struct {
	svc    sms.Service
	tracer trace.Tracer
}

func NewDecorator(svc sms.Service, tracer trace.Tracer) *Decorator {
	return &Decorator{svc: svc, tracer: tracer}
}

func (d *Decorator) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	ctx, span := d.tracer.Start(ctx, "sms")
	defer span.End()
	// 你也可以考虑拼接进去 span name 里面
	span.SetAttributes(attribute.String("tpl", tplId))
	span.AddEvent("发短信")
	err := d.svc.Send(ctx, tplId, args, numbers...)
	if err != nil {
		span.RecordError(err)
	}
	return err
}
