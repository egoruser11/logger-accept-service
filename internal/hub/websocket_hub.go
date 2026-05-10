package hub

import (
	"Realtime_Log_Aggregator/internal/models"
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

type Hub struct {
	Clients map[*websocket.Conn]bool

	Register   chan *websocket.Conn
	Unregister chan *websocket.Conn

	Broadcast chan models.LogInputRequest

	Mutex sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*websocket.Conn]bool),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
		Broadcast:  make(chan models.LogInputRequest, 100),
	}
}

func (h *Hub) Run() {
	for {
		select {

		case conn := <-h.Register:
			h.Mutex.Lock()
			h.Clients[conn] = true
			h.Mutex.Unlock()

		case conn := <-h.Unregister:
			h.Mutex.Lock()
			if _, ok := h.Clients[conn]; ok {
				delete(h.Clients, conn)
				conn.Close()
			}
			h.Mutex.Unlock()

		case log := <-h.Broadcast:
			h.broadcast(log)
		}
	}
}

func (h *Hub) broadcast(log models.LogInputRequest) {

	h.Mutex.RLock()
	clients := make([]*websocket.Conn, 0, len(h.Clients))

	for c := range h.Clients {
		clients = append(clients, c)
	}
	h.Mutex.RUnlock()

	for _, conn := range clients {
		err := conn.WriteJSON(log)
		fmt.Println(err)
		if err != nil {
			// если клиент мёртв — отправляем на удаление
			h.Unregister <- conn
		}
	}
}
