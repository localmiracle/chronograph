package chronograph

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"gonum.org/v1/gonum/graph/simple"
)

// SpanRecord — завершённый span с длительностью
type SpanRecord struct {
	ID        uuid.UUID
	ParentID  *uuid.UUID
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// BuildRecords преобразует SpanEvent → SpanRecord
func BuildRecords(events []SpanEvent) ([]SpanRecord, error) {
	recMap := make(map[uuid.UUID]*SpanRecord)
	for _, e := range events {
		if isEnter(e) {
			recMap[e.ID] = &SpanRecord{
				ID:        e.ID,
				ParentID:  e.ParentID,
				Name:      e.Name[:len(e.Name)-6], // убираем суффикс ".enter"
				StartTime: e.Timestamp,
			}
		} else if isExit(e) {
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

	// в слайс
	records := make([]SpanRecord, 0, len(recMap))
	for _, r := range recMap {
		records = append(records, *r)
	}
	return records, nil
}

// Summarize фильтрует и сортирует по длительности
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

// PrintSummary выводит таблицу результатов
func PrintSummary(records []SpanRecord) {
	fmt.Println("Сводка спанов (длительность >= порог):")
	fmt.Printf("%-36s %-20s %-10s\n", "ID", "Name", "Duration")
	for _, r := range records {
		fmt.Printf("%-36s %-20s %v\n", r.ID, r.Name, r.Duration)
	}
}

// PrunedGraph строит «сжатый» граф: оставляет только span’ы с Duration>=threshold,
// а узлы меньшей длительности удаляет, но «напрямую» соединяет их ближайших kept-родителей с потомками.
func PrunedGraph(records []SpanRecord, threshold time.Duration) *simple.DirectedGraph {
	// соберём map всех записей по ID
	recMap := make(map[uuid.UUID]SpanRecord, len(records))
	for _, r := range records {
		recMap[r.ID] = r
	}
	// выберем span’ы >= threshold
	kept := make(map[uuid.UUID]SpanRecord)
	for _, r := range records {
		if r.Duration >= threshold {
			kept[r.ID] = r
		}
	}
	g := simple.NewDirectedGraph()
	// добавим все «kept» узлы
	for id := range kept {
		g.AddNode(simple.Node(id.ID()))
	}
	// для каждого kept-узла найдём ближайшего kept-предка и свяжем
	for id, r := range kept {
		parentID := r.ParentID
		for parentID != nil {
			pr := recMap[*parentID]
			if pr.Duration >= threshold {
				g.SetEdge(g.NewEdge(
					simple.Node(pr.ID.ID()),
					simple.Node(id.ID()),
				))
				break
			}
			parentID = pr.ParentID
		}
	}
	return g
}

// InferRootCause возвращает цепочку span’ов от корневого (rootName) до
// наиболее долгого span’а (кроме самого root), считая, что самый
// проблемный — с наибольшей Duration.
func InferRootCause(records []SpanRecord, rootName string) []SpanRecord {
	recMap := make(map[uuid.UUID]SpanRecord, len(records))
	for _, r := range records {
		recMap[r.ID] = r
	}
	// кандидаты — все, кроме rootName
	var cands []SpanRecord
	for _, r := range records {
		if r.Name != rootName {
			cands = append(cands, r)
		}
	}
	if len(cands) == 0 {
		return nil
	}
	// самый долгий
	sort.Slice(cands, func(i, j int) bool {
		return cands[i].Duration > cands[j].Duration
	})
	culprit := cands[0]
	// строим путь от root до culprit
	var chain []SpanRecord
	cur := culprit
	for {
		chain = append([]SpanRecord{cur}, chain...)
		if cur.Name == rootName || cur.ParentID == nil {
			break
		}
		cur = recMap[*cur.ParentID]
	}
	return chain
}

// PrintRootCause выводит найденную цепочку «корня зла»
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
