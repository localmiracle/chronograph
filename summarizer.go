package chronograph

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
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
