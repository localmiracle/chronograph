package chronograph

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey string

const (
	collectorKey ctxKey = "chronograph-collector"
	spanKey      ctxKey = "chronograph-span"
)

// ContextWithCollector встраивает Collector в контекст
func ContextWithCollector(ctx context.Context, col *EventCollector) context.Context {
	return context.WithValue(ctx, collectorKey, col)
}

func getCollector(ctx context.Context) *EventCollector {
	v := ctx.Value(collectorKey)
	if col, ok := v.(*EventCollector); ok {
		return col
	}
	panic("chronograph: Collector не найден в контексте")
}

// StartSpan создаёт событие enter и возвращает context с текущим span ID
func StartSpan(ctx context.Context, name string) (context.Context, uuid.UUID) {
	col := getCollector(ctx)
	parent := col.CurrentSpan()
	enter := NewEnter(name, parent)
	col.Push(enter)
	return context.WithValue(ctx, spanKey, enter.ID), enter.ID
}

// EndSpan создаёт событие exit, беря span ID из context
func EndSpan(ctx context.Context, name string) {
	col := getCollector(ctx)
	v := ctx.Value(spanKey)
	id, ok := v.(uuid.UUID)
	if !ok {
		panic("chronograph: EndSpan без StartSpan")
	}
	exit := NewExit(name, id)
	col.Push(exit)
}
