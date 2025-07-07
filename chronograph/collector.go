package chronograph

import (
	"sync"

	"github.com/google/uuid"
	"gonum.org/v1/gonum/graph/simple"
)

// EventCollector собирает SpanEvent и строит DAG
type EventCollector struct {
	mu     sync.Mutex
	stack  []uuid.UUID
	events []SpanEvent
}

// NewCollector создаёт новый
func NewCollector() *EventCollector {
	return &EventCollector{
		stack:  make([]uuid.UUID, 0),
		events: make([]SpanEvent, 0),
	}
}

// CurrentSpan возвращает active span ID или nil
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
	if IsEnter(e) {
		c.stack = append(c.stack, e.ID)
	} else if IsExit(e) {
		if len(c.stack) == 0 {
			panic("chronograph: Exit без Enter")
		}
		c.stack = c.stack[:len(c.stack)-1]
	}
	c.events = append(c.events, e)
}

// Events возвращает копию всех собранных событий
func (c *EventCollector) Events() []SpanEvent {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]SpanEvent, len(c.events))
	copy(cp, c.events)
	return cp
}

// BuildGraph строит DirectedGraph span-эвентов
func (c *EventCollector) BuildGraph() *simple.DirectedGraph {
	c.mu.Lock()
	defer c.mu.Unlock()
	g := simple.NewDirectedGraph()
	for _, e := range c.events {
		node := simple.Node(e.ID.ID())
		if g.Node(node.ID()) == nil {
			g.AddNode(node)
		}
	}
	for _, e := range c.events {
		if e.ParentID != nil {
			parent := simple.Node(e.ParentID.ID())
			child := simple.Node(e.ID.ID())
			g.SetEdge(g.NewEdge(parent, child))
		}
	}
	return g
}
