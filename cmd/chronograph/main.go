package main

import (
	"context"
	"fmt"
	"os"
	"time"

	chronopkg "github.com/localmiracle/chronograph"
	"github.com/spf13/cobra"
)

var (
	threshold time.Duration
	output    string
)

func main() {
	root := &cobra.Command{
		Use:   "chronograph",
		Short: "CLI для ChronoGraph",
		Run:   run,
	}
	root.Flags().DurationVarP(&threshold, "threshold", "t", 1*time.Millisecond, "порог")
	root.Flags().StringVarP(&output, "output", "o", "pruned.dot", "DOT-файл")

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	col := chronopkg.NewCollector()
	ctx := chronopkg.ContextWithCollector(context.Background(), col)

	ctx, _ = chronopkg.StartSpan(ctx, "main")
	step1(ctx)
	step2(ctx)
	chronopkg.EndSpan(ctx, "main")

	events := col.Events()
	records, err := chronopkg.BuildRecords(events)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("=== Summary ===")
	summary := chronopkg.Summarize(records, threshold)
	chronopkg.PrintSummary(summary)

	pruned := chronopkg.PrunedGraph(records, threshold)
	fmt.Printf("\nPruned graph: nodes=%d, edges=%d\n", pruned.Nodes().Len(), pruned.Edges().Len())
	f, _ := os.Create(output)
	defer f.Close()
	f.WriteString("digraph ChronoGraph {\n")
	for it := pruned.Nodes(); it.Next(); {
		n := it.Node()
		fmt.Fprintf(f, " \"%d\";\n", n.ID())
	}
	for it := pruned.Edges(); it.Next(); {
		e := it.Edge()
		fmt.Fprintf(f, " \"%d\" -> \"%d\";\n", e.From().ID(), e.To().ID())
	}
	f.WriteString("}\n")
	fmt.Println("DOT saved to", output)

	fmt.Println("\n=== Root Cause ===")
	chain := chronopkg.InferRootCause(records, "main")
	chronopkg.PrintRootCause(chain)
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
