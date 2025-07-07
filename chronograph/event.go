package chronograph

import (
	"time"

	"github.com/google/uuid"
)

// SpanEvent — одно событие входа/выхода в span
type SpanEvent struct {
	ID        uuid.UUID
	ParentID  *uuid.UUID
	Name      string
	Timestamp time.Time
	Meta      map[string]string
}

// NewEnter создаёт событие ".enter"
func NewEnter(name string, parent *uuid.UUID) SpanEvent {
	return SpanEvent{
		ID:        uuid.New(),
		ParentID:  parent,
		Name:      name + ".enter",
		Timestamp: time.Now().UTC(),
		Meta:      make(map[string]string),
	}
}

// NewExit создаёт событие ".exit"
func NewExit(name string, parent uuid.UUID) SpanEvent {
	return SpanEvent{
		ID:        uuid.New(),
		ParentID:  &parent,
		Name:      name + ".exit",
		Timestamp: time.Now().UTC(),
		Meta:      make(map[string]string),
	}
}

// IsEnter проверяет, что событие — ".enter"
func IsEnter(e SpanEvent) bool {
	return len(e.Name) >= 6 && e.Name[len(e.Name)-6:] == ".enter"
}

// IsExit проверяет, что событие — ".exit"
func IsExit(e SpanEvent) bool {
	return len(e.Name) >= 5 && e.Name[len(e.Name)-5:] == ".exit"
}
