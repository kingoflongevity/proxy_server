package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"proxy_server/internal/service"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源的WebSocket连接
	},
}

type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type LogEntry struct {
	Timestamp   string `json:"timestamp"`
	Level       string `json:"level"`
	Message     string `json:"message"`
	Source      string `json:"source"`
	ConnectionID string `json:"connectionId,omitempty"`
}

type ConnectionInfo struct {
	ID           string `json:"id"`
	Protocol     string `json:"protocol"`
	SourceIP     string `json:"sourceIp"`
	DestHost     string `json:"destHost"`
	DestPort     int    `json:"destPort"`
	UploadBytes  int64  `json:"uploadBytes"`
	DownloadBytes int64 `json:"downloadBytes"`
	StartTime    string `json:"startTime"`
}

type WebSocketClient struct {
	conn      *websocket.Conn
	send      chan []byte
	closeChan chan struct{}
}

type WebSocketHub struct {
	clients    map[*WebSocketClient]bool
	broadcast  chan []byte
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	mu         sync.RWMutex
}

var wsHub = &WebSocketHub{
	clients:    make(map[*WebSocketClient]bool),
	broadcast:  make(chan []byte, 256),
	register:   make(chan *WebSocketClient),
	unregister: make(chan *WebSocketClient),
}

func init() {
	go wsHub.run()
}

func (h *WebSocketHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("WebSocket client disconnected. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *WebSocketHub) Broadcast(messageType string, data interface{}) {
	msg := WebSocketMessage{
		Type: messageType,
		Data: data,
	}
	if data, err := json.Marshal(msg); err == nil {
		h.broadcast <- data
	}
}

func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.closeChan:
			return
		}
	}
}

func (c *WebSocketClient) readPump() {
	defer func() {
		wsHub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

type WebSocketHandler struct {
	logService service.LogService
}

func NewWebSocketHandler(logService service.LogService) *WebSocketHandler {
	return &WebSocketHandler{
		logService: logService,
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &WebSocketClient{
		conn:      conn,
		send:      make(chan []byte, 256),
		closeChan: make(chan struct{}),
	}
	wsHub.register <- client

	go client.writePump()
	go client.readPump()
	go h.sendInitialData(client)
}

func (h *WebSocketHandler) sendInitialData(client *WebSocketClient) {
	time.Sleep(100 * time.Millisecond)
	wsHub.Broadcast("connected", map[string]string{
		"status": "connected",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func BroadcastLog(level, message, source string) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Source:    source,
	}
	wsHub.Broadcast("log", entry)
}

func BroadcastConnection(info *ConnectionInfo) {
	wsHub.Broadcast("connection", info)
}

func BroadcastDisconnection(id string) {
	wsHub.Broadcast("disconnection", map[string]string{
		"id": id,
	})
}

func BroadcastTrafficUpdate(connectionId string, upload, download int64) {
	wsHub.Broadcast("traffic", map[string]interface{}{
		"connectionId":   connectionId,
		"uploadBytes":    upload,
		"downloadBytes":  download,
		"timestamp":      time.Now().Format(time.RFC3339),
	})
}

func BroadcastNodeUpdate(nodeId string, status string, latency int) {
	wsHub.Broadcast("node_update", map[string]interface{}{
		"nodeId":   nodeId,
		"status":   status,
		"latency":  latency,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// BroadcastNode 广播节点更新（简化版）
func BroadcastNode(nodeId, status string, latency int) {
	BroadcastNodeUpdate(nodeId, status, latency)
}

func BroadcastSubscriptionUpdate(subscriptionId string, nodeCount int, status string) {
	wsHub.Broadcast("subscription_update", map[string]interface{}{
		"subscriptionId": subscriptionId,
		"nodeCount":      nodeCount,
		"status":         status,
		"timestamp":      time.Now().Format(time.RFC3339),
	})
}

// BroadcastSubscription 广播订阅更新（简化版）
func BroadcastSubscription(subscriptionId string, nodeCount int, status string) {
	BroadcastSubscriptionUpdate(subscriptionId, nodeCount, status)
}

func GetConnectedClients() int {
	wsHub.mu.RLock()
	defer wsHub.mu.RUnlock()
	return len(wsHub.clients)
}
