package hub

import (
	"Realtime_Log_Aggregator/internal/models"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type Hub struct {
	Clients map[*websocket.Conn]bool

	Register       chan *websocket.Conn
	Unregister     chan *websocket.Conn
	PendingLogs    sync.WaitGroup
	Broadcast      chan models.LogInputRequest
	IsShuttingDown bool
	Mutex          sync.RWMutex
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

		case conn, ok := <-h.Register:
			if !ok {
				return
			}
			h.Mutex.Lock()
			h.Clients[conn] = true
			h.Mutex.Unlock()

		case conn, ok := <-h.Unregister:
			if !ok {
				return
			}
			h.Mutex.Lock()
			if _, ok := h.Clients[conn]; ok {
				delete(h.Clients, conn)
				conn.Close()
			}
			h.Mutex.Unlock()

		case log, ok := <-h.Broadcast:
			if !ok {
				return
			}
			h.broadcast(log)
			h.PendingLogs.Done()
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

func (h *Hub) Shutdown() {
	h.Mutex.Lock()
	h.IsShuttingDown = true
	h.Mutex.Unlock()
	h.PendingLogs.Wait()
	h.Mutex.RLock()
	clients := make([]*websocket.Conn, 0, len(h.Clients))
	for c := range h.Clients {
		clients = append(clients, c)
	}
	h.Mutex.RUnlock()

	var wg sync.WaitGroup
	for _, conn := range clients {
		wg.Add(1)
		go func(conn *websocket.Conn) {
			defer wg.Done()

			closeMsg := map[string]string{
				"type":    "shutdown",
				"message": "Server is stopping",
			}
			conn.WriteJSON(closeMsg)
			conn.Close()
		}(conn)
	}
	wg.Wait()

	close(h.Register)
	close(h.Unregister)
	close(h.Broadcast)

	log.Println("✅ Hub shutdown complete")
}
