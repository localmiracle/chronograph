package chronograph

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"gonum.org/v1/gonum/graph/simple"
)

// SpanRecord — span с рассчитанной длительностью
type SpanRecord struct {
	ID        uuid.UUID
	ParentID  *uuid.UUID
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// BuildRecords конвертирует SpanEvent → SpanRecord
func BuildRecords(events []SpanEvent) ([]SpanRecord, error) {
	recMap := make(map[uuid.UUID]*SpanRecord)
	for _, e := range events {
		if IsEnter(e) {
			recMap[e.ID] = &SpanRecord{
				ID:        e.ID,
				ParentID:  e.ParentID,
				Name:      e.Name[:len(e.Name)-6],
				StartTime: e.Timestamp,
			}
		} else if IsExit(e) {
			if e.ParentID == nil {
				continue
			}
			pid := *e.ParentID
			if r, ok := recMap[pid]; ok {
				r.EndTime = e.Timestamp
				r.Duration = r.EndTime.Sub(r.StartTime)
			}
		}
	}
	out := make([]SpanRecord, 0, len(recMap))
	for _, r := range recMap {
		out = append(out, *r)
	}
	return out, nil
}

// Summarize фильтрует и сортирует records по порогу
func Summarize(records []SpanRecord, threshold time.Duration) []SpanRecord {
	var out []SpanRecord
	for _, r := range records {
		if r.Duration >= threshold {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Duration > out[j].Duration
	})
	return out
}

// PrintSummary выводит табличку записей
func PrintSummary(records []SpanRecord) {
	fmt.Println("Сводка спанов:")
	fmt.Printf("%-36s %-20s %-10s\n", "ID", "Name", "Duration")
	for _, r := range records {
		fmt.Printf("%-36s %-20s %v\n", r.ID, r.Name, r.Duration)
	}
}

// PrunedGraph строит «сжатый» граф span’ов >= порога
func PrunedGraph(records []SpanRecord, threshold time.Duration) *simple.DirectedGraph {
	recMap := make(map[uuid.UUID]SpanRecord)
	for _, r := range records {
		recMap[r.ID] = r
	}
	kept := make(map[uuid.UUID]SpanRecord)
	for _, r := range records {
		if r.Duration >= threshold {
			kept[r.ID] = r
		}
	}
	g := simple.NewDirectedGraph()
	for id := range kept {
		g.AddNode(simple.Node(id.ID()))
	}
	for id, r := range kept {
		parent := r.ParentID
		for parent != nil {
			pr := recMap[*parent]
			if pr.Duration >= threshold {
				g.SetEdge(g.NewEdge(
					simple.Node(pr.ID.ID()),
					simple.Node(id.ID()),
				))
				break
			}
			parent = pr.ParentID
		}
	}
	return g
}

// InferRootCause возвращает цепочку от rootName до самого долгого span
func InferRootCause(records []SpanRecord, rootName string) []SpanRecord {
	recMap := make(map[uuid.UUID]SpanRecord)
	for _, r := range records {
		recMap[r.ID] = r
	}
	var cands []SpanRecord
	for _, r := range records {
		if r.Name != rootName {
			cands = append(cands, r)
		}
	}
	if len(cands) == 0 {
		return nil
	}
	sort.Slice(cands, func(i, j int) bool {
		return cands[i].Duration > cands[j].Duration
	})
	culprit := cands[0]
	var path []SpanRecord
	cur := culprit
	for {
		path = append([]SpanRecord{cur}, path...)
		if cur.Name == rootName || cur.ParentID == nil {
			break
		}
		cur = recMap[*cur.ParentID]
	}
	return path
}

// PrintRootCause выводит найденную цепочку
func PrintRootCause(chain []SpanRecord) {
	if len(chain) == 0 {
		fmt.Println("Root cause not found")
		return
	}
	fmt.Println("Root cause path:")
	for _, r := range chain {
		fmt.Printf(" → %s (%v)\n", r.Name, r.Duration)
	}
}
