package broker

import "sync"

type OrderEvent struct {
	OrderID string
	Status  string
}

type Hub struct {
	mu          sync.RWMutex
	subscribers map[string][]chan OrderEvent
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[string][]chan OrderEvent),
	}
}

func (h *Hub) Publish(event OrderEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if chans, found := h.subscribers[event.OrderID]; found {
		for _, ch := range chans {
			select {
			case ch <- event:
			default:
			}
		}
	}
}

func (h *Hub) Subscribe(orderID string) chan OrderEvent {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan OrderEvent, 10)
	h.subscribers[orderID] = append(h.subscribers[orderID], ch)

	return ch
}
