// cmd/chronograph/main.go

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	chronopkg "github.com/localmiracle/chronograph/chronograph"
	"github.com/spf13/cobra"
)

// JSONOutput — структура для JSON-вывода
type JSONOutput struct {
	Summary     []chronopkg.SpanRecord `json:"summary"`
	PrunedNodes []uint64               `json:"pruned_nodes"`
	PrunedEdges [][2]uint64            `json:"pruned_edges"`
	RootCause   []chronopkg.SpanRecord `json:"root_cause"`
}

var (
	threshold time.Duration
	output    string
	format    string
)

func main() {
	root := &cobra.Command{
		Use:   "chronograph",
		Short: "CLI для ChronoGraph: трассировка и анализ span’ов",
		Run:   run,
	}
	root.Flags().DurationVarP(&threshold, "threshold", "t", 1*time.Millisecond,
		"Порог фильтрации спанов (например 500µs, 2ms)")
	root.Flags().StringVarP(&output, "output", "o", "pruned.dot",
		"Файл для экспорта сжатого графа в DOT")
	root.Flags().StringVarP(&format, "format", "f", "text",
		"Формат вывода: text или json")

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	// 1) Инициализация коллектор + контекст
	col := chronopkg.NewCollector()
	ctx := chronopkg.ContextWithCollector(context.Background(), col)

	// 2) Главный span
	ctx, _ = chronopkg.StartSpan(ctx, "main")
	step1(ctx)
	step2(ctx)
	chronopkg.EndSpan(ctx, "main")

	// 3) Собираем span-записи
	events := col.Events()
	records, err := chronopkg.BuildRecords(events)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error building records:", err)
		os.Exit(1)
	}

	// 4) Применяем фильтр + строим pruned-graph + root-cause
	summary := chronopkg.Summarize(records, threshold)
	pruned := chronopkg.PrunedGraph(records, threshold)
	rootPath := chronopkg.InferRootCause(records, "main")

	switch format {
	case "json":
		// Подготавливаем JSONOutput
		var out JSONOutput
		out.Summary = summary

		// Собираем узлы
		for it := pruned.Nodes(); it.Next(); {
			out.PrunedNodes = append(out.PrunedNodes, uint64(it.Node().ID()))
		}
		// Собираем рёбра
		for it := pruned.Edges(); it.Next(); {
			e := it.Edge()
			out.PrunedEdges = append(out.PrunedEdges, [2]uint64{
				uint64(e.From().ID()),
				uint64(e.To().ID()),
			})
		}
		out.RootCause = rootPath

		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "JSON marshal error:", err)
			os.Exit(1)
		}
		fmt.Println(string(b))

	default: // text
		// Текстовый вывод
		fmt.Println("=== Summary ===")
		chronopkg.PrintSummary(summary)

		fmt.Printf("\nPruned graph: nodes=%d, edges=%d\n",
			pruned.Nodes().Len(), pruned.Edges().Len(),
		)

		// Экспорт DOT
		f, err := os.Create(output)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot create file:", err)
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
		fmt.Println("DOT saved to", output)

		fmt.Println("\n=== Root Cause ===")
		chronopkg.PrintRootCause(rootPath)
	}
}

func step1(ctx context.Context) {
	ctx, _ = chronopkg.StartSpan(ctx, "step1")
	time.Sleep(5 * time.Millisecond)
	chronopkg.EndSpan(ctx, "step1")
}

func step2(ctx context.Context) {
	ctx, _ = chronopkg.StartSpan(ctx, "step2")
	time.Sleep(500 * time.Microsecond)
	chronopkg.EndSpan(ctx, "step2")
}
