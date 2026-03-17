package websocket

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu        sync.RWMutex
	listeners map[string]map[chan string]bool
}

var DeployHub = &Hub{
	listeners: make(map[string]map[chan string]bool),
}

func (h *Hub) Subscribe(id string) chan string {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan string, 10)
	if _, ok := h.listeners[id]; !ok {
		h.listeners[id] = make(map[chan string]bool)
	}
	h.listeners[id][ch] = true
	return ch
}

func (h *Hub) Unsubscribe(id string, ch chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.listeners[id]; ok {
		delete(clients, ch)
		close(ch)
		if len(clients) == 0 {
			delete(h.listeners, id)
		}
	}
}

func (h *Hub) Broadcast(id, msg string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.listeners[id]; ok {
		for ch := range clients {
			select {
			case ch <- msg:
			default:
			}
		}
	}
}

func StreamDeployProgress(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing deployment ID", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	progressCh := DeployHub.Subscribe(id)
	defer DeployHub.Unsubscribe(id, progressCh)

	for msg := range progressCh {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			break
		}
	}
}
