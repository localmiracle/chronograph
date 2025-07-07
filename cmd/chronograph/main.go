package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/localmiracle/chronograph"
	"github.com/spf13/cobra"
)

var (
	threshold time.Duration
	output    string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "chronograph",
		Short: "ChronoGraph CLI — отладка распределённых трасс",
		Run:   runCmd,
	}

	rootCmd.Flags().DurationVarP(&threshold, "threshold", "t", 1*time.Millisecond,
		"Порог фильтрации спанов (например 500µs, 2ms)")
	rootCmd.Flags().StringVarP(&output, "output", "o", "graph.dot",
		"Файл для экспорта сжатого графа в DOT-формате")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runCmd(cmd *cobra.Command, args []string) {
	// 1. Инициализация коллектор + контекст
	col := chronograph.NewCollector()
	ctx := chronograph.ContextWithCollector(context.Background(), col)

	// 2. Запускаем главный span и эмулируем работу
	ctx, _ = chronograph.StartSpan(ctx, "main")
	step1(ctx)
	step2(ctx)
	chronograph.EndSpan(ctx, "main")

	// 3. Собираем записи
	events := col.Events()
	records, err := chronograph.BuildRecords(events)
	if err != nil {
		fmt.Println("Ошибка:", err)
		os.Exit(1)
	}

	// 4. Summary
	fmt.Println("=== Summary ===")
	summary := chronograph.Summarize(records, threshold)
	chronograph.PrintSummary(summary)

	// 5. Pruned graph + экспорт DOT
	pruned := chronograph.PrunedGraph(records, threshold)
	fmt.Printf("\nPruned graph: nodes=%d, edges=%d\n",
		pruned.Nodes().Len(), pruned.Edges().Len(),
	)

	f, err := os.Create(output)
	if err != nil {
		fmt.Println("Не удалось создать файл:", err)
		os.Exit(1)
	}
	defer f.Close()

	f.WriteString("digraph ChronoGraph {\n")
	for it := pruned.Nodes(); it.Next(); {
		n := it.Node()
		fmt.Fprintf(f, "  \"%d\";\n", n.ID())
	}
	for it := pruned.Edges(); it.Next(); {
		e := it.Edge()
		fmt.Fprintf(f, "  \"%d\" -> \"%d\";\n", e.From().ID(), e.To().ID())
	}
	f.WriteString("}\n")
	fmt.Println("DOT сохранён в", output)

	// 6. Root cause
	fmt.Println("\n=== Root Cause ===")
	chain := chronograph.InferRootCause(records, "main")
	chronograph.PrintRootCause(chain)
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
