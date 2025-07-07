package chronograph

import (
	"context"

	"github.com/google/uuid"
	logic "github.com/localmiracle/chronograph/internal/chronographlogic"
)

type ctxKey string

const (
	collectorKey ctxKey = "chronograph-collector"
	spanKey      ctxKey = "chronograph-span"
)

func NewCollector() *logic.EventCollector {
	return logic.NewCollector()
}

func ContextWithCollector(ctx context.Context, col *logic.EventCollector) context.Context {
	return context.WithValue(ctx, collectorKey, col)
}

func getCollector(ctx context.Context) *logic.EventCollector {
	v := ctx.Value(collectorKey)
	if col, ok := v.(*logic.EventCollector); ok {
		return col
	}
	panic("chronograph: EventCollector not found")
}

func StartSpan(ctx context.Context, name string) (context.Context, uuid.UUID) {
	col := getCollector(ctx)
	parent := col.CurrentSpan()
	enter := logic.NewEnter(name, parent)
	col.Push(enter)
	return context.WithValue(ctx, spanKey, enter.ID), enter.ID
}

func EndSpan(ctx context.Context, name string) {
	col := getCollector(ctx)
	v := ctx.Value(spanKey)
	id, ok := v.(uuid.UUID)
	if !ok {
		panic("chronograph: EndSpan without StartSpan")
	}
	exit := logic.NewExit(name, id)
	col.Push(exit)
}
