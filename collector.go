package chronograph

import (
	"sync"

	"github.com/google/uuid"
	"gonum.org/v1/gonum/graph/simple"
)

// EventCollector хранит стек активных span’ов и все события
type EventCollector struct {
	mu     sync.Mutex
	stack  []uuid.UUID // стек идентификаторов активных span’ов
	events []SpanEvent // все полученные события
}

// NewCollector создаёт новый EventCollector
func NewCollector() *EventCollector {
	return &EventCollector{
		stack:  make([]uuid.UUID, 0),
		events: make([]SpanEvent, 0),
	}
}

// CurrentSpan возвращает ID активного span или nil
func (c *EventCollector) CurrentSpan() *uuid.UUID {
	if len(c.stack) == 0 {
		return nil
	}
	return &c.stack[len(c.stack)-1]
}

// Push сохраняет событие и обновляет стек
func (c *EventCollector) Push(e SpanEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if isEnter(e) {
		c.stack = append(c.stack, e.ID)
	} else if isExit(e) {
		if len(c.stack) == 0 {
			panic("chronograph: Exit без соответствующего Enter")
		}
		c.stack = c.stack[:len(c.stack)-1]
	}
	c.events = append(c.events, e)
}

// Events возвращает копию всех собранных событий
func (c *EventCollector) Events() []SpanEvent {
	c.mu.Lock()
	defer c.mu.Unlock()
	copyEvents := make([]SpanEvent, len(c.events))
	copy(copyEvents, c.events)
	return copyEvents
}

// BuildGraph строит направленный граф (DAG) из всех span-событий
func (c *EventCollector) BuildGraph() *simple.DirectedGraph {
	c.mu.Lock()
	defer c.mu.Unlock()

	g := simple.NewDirectedGraph()

	// Добавляем все события как узлы
	for _, e := range c.events {
		node := simple.Node(e.ID.ID())
		if g.Node(node.ID()) == nil {
			g.AddNode(node)
		}
	}

	// Добавляем ребра: parent → child
	for _, e := range c.events {
		if e.ParentID != nil {
			parentNode := simple.Node(e.ParentID.ID())
			childNode := simple.Node(e.ID.ID())
			g.SetEdge(g.NewEdge(parentNode, childNode))
		}
	}
	return g
}
