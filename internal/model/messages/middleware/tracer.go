package middleware

import (
	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/clients/tg"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"golang.org/x/net/context"
)

type Tracer struct {
	next tg.Message
}

func NewTracer(next tg.Message) *Tracer {
	return &Tracer{next: next}
}

func (t *Tracer) IncomingMessage(ctx context.Context, msg *dto.Message) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "incoming command")
	defer span.Finish()

	err := t.next.IncomingMessage(ctx, msg)
	if err != nil {
		return err
	}
	return nil
}
