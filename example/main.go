package main

import (
	"context"
	"fmt"
	"time"

	"github.com/localmiracle/chronograph"
)

func main() {
	// 1. Инициализация collector и контекста
	col := chronograph.NewCollector()
	ctx := chronograph.ContextWithCollector(context.Background(), col)

	// 2. Главный span
	ctx, _ = chronograph.StartSpan(ctx, "main")

	// 3. Функции-спаны
	step1(ctx)
	step2(ctx)

	// 4. Завершение главного span
	chronograph.EndSpan(ctx, "main")

	// 5. Построение графа для проверки
	g := col.BuildGraph()
	fmt.Printf("Граф: узлов=%d, ребер=%d\n", g.Nodes().Len(), g.Edges().Len())

	// 6. Сборка span-записей и фильтрация
	events := col.Events()
	records, err := chronograph.BuildRecords(events)
	if err != nil {
		panic(err)
	}
	summary := chronograph.Summarize(records, 1*time.Millisecond)
	chronograph.PrintSummary(summary)
}

func step1(ctx context.Context) {
	ctx, _ = chronograph.StartSpan(ctx, "step1")
	time.Sleep(5 * time.Millisecond)
	chronograph.EndSpan(ctx, "step1")
}

func step2(ctx context.Context) {
	ctx, _ = chronograph.StartSpan(ctx, "step2")
	time.Sleep(500 * time.Microsecond)
	chronograph.EndSpan(ctx, "step2")
}
