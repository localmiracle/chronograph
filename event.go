package chronograph

import (
	"time"

	"github.com/google/uuid"
)

// SpanEvent представляет одно событие входа/выхода из span’а
type SpanEvent struct {
	ID        uuid.UUID         // уникальный идентификатор события
	ParentID  *uuid.UUID        // nil для корневых span’ов
	Name      string            // метка события, например "step1.enter"
	Timestamp time.Time         // время создания события
	Meta      map[string]string // дополнительные метаданные
}

// NewEnter создаёт событие типа "enter"
func NewEnter(name string, parent *uuid.UUID) SpanEvent {
	return SpanEvent{
		ID:        uuid.New(),
		ParentID:  parent,
		Name:      name + ".enter",
		Timestamp: time.Now().UTC(),
		Meta:      make(map[string]string),
	}
}

// NewExit создаёт событие типа "exit"
func NewExit(name string, parent uuid.UUID) SpanEvent {
	return SpanEvent{
		ID:        uuid.New(),
		ParentID:  &parent,
		Name:      name + ".exit",
		Timestamp: time.Now().UTC(),
		Meta:      make(map[string]string),
	}
}

// вспомогательные функции для распознавания типа события

func isEnter(e SpanEvent) bool {
	return len(e.Name) >= 6 && e.Name[len(e.Name)-6:] == ".enter"
}

func isExit(e SpanEvent) bool {
	return len(e.Name) >= 5 && e.Name[len(e.Name)-5:] == ".exit"
}
