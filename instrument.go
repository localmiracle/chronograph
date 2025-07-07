package chronograph

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey string

const (
	// collectorKey — ключ для хранения *EventCollector в контексте
	collectorKey ctxKey = "chronograph-collector"
	// spanKey — ключ для хранения текущего span ID в контексте
	spanKey ctxKey = "chronograph-span"
)

// ContextWithCollector возвращает новый контекст с вложенным EventCollector
func ContextWithCollector(ctx context.Context, col *EventCollector) context.Context {
	return context.WithValue(ctx, collectorKey, col)
}

// getCollector извлекает EventCollector из контекста
func getCollector(ctx context.Context) *EventCollector {
	val := ctx.Value(collectorKey)
	if col, ok := val.(*EventCollector); ok {
		return col
	}
	panic("chronograph: в контексте не найден EventCollector")
}

// StartSpan начинает span, пишет событие enter и сохраняет ID в контекст
func StartSpan(ctx context.Context, name string) (context.Context, uuid.UUID) {
	col := getCollector(ctx)
	parent := col.CurrentSpan()
	enter := NewEnter(name, parent)
	col.Push(enter)
	return context.WithValue(ctx, spanKey, enter.ID), enter.ID
}

// EndSpan закрывает span, пишет событие exit по ID из контекста
func EndSpan(ctx context.Context, name string) {
	col := getCollector(ctx)
	val := ctx.Value(spanKey)
	id, ok := val.(uuid.UUID)
	if !ok {
		panic("chronograph: EndSpan без предварительного StartSpan")
	}
	exit := NewExit(name, id)
	col.Push(exit)
}
