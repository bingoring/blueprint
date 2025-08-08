package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// SSEClient represents a connected SSE client
type SSEClient struct {
	ID           string
	MilestoneID  uint
	Channel      chan []byte
	Request      *http.Request
	Writer       gin.ResponseWriter
}

// SSEMessage represents a Server-Sent Event message
type SSEMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// MarketUpdateEvent represents a market update event
type MarketUpdateEvent struct {
	MilestoneID uint        `json:"milestone_id"`
	MarketData  interface{} `json:"market_data"`
	Timestamp   int64       `json:"timestamp"`
}

// SSEService manages Server-Sent Events for real-time updates
type SSEService struct {
	clients    map[string]*SSEClient
	clientsMux sync.RWMutex

	// Channel for broadcasting messages to all clients
	broadcast chan SSEMessage

	// Channel for adding new clients
	register chan *SSEClient

	// Channel for removing clients
	unregister chan *SSEClient
}

// NewSSEService creates a new SSE service
func NewSSEService() *SSEService {
	service := &SSEService{
		clients:    make(map[string]*SSEClient),
		broadcast:  make(chan SSEMessage, 100),
		register:   make(chan *SSEClient),
		unregister: make(chan *SSEClient),
	}

	// Start the service in a goroutine
	go service.run()

	return service
}

// run handles the main event loop for the SSE service
func (s *SSEService) run() {
	for {
		select {
		case client := <-s.register:
			s.clientsMux.Lock()
			s.clients[client.ID] = client
			s.clientsMux.Unlock()

			log.Printf("SSE client connected: %s for milestone %d", client.ID, client.MilestoneID)

			// Send welcome message
			welcomeMsg := SSEMessage{
				Type:      "connection",
				Data:      map[string]interface{}{"status": "connected", "milestone_id": client.MilestoneID},
				Timestamp: time.Now().Unix(),
			}
			s.sendToClient(client, welcomeMsg)

		case client := <-s.unregister:
			s.clientsMux.Lock()
			if _, ok := s.clients[client.ID]; ok {
				delete(s.clients, client.ID)
				close(client.Channel)
			}
			s.clientsMux.Unlock()

			log.Printf("SSE client disconnected: %s", client.ID)

		case message := <-s.broadcast:
			s.clientsMux.RLock()
			for _, client := range s.clients {
				s.sendToClient(client, message)
			}
			s.clientsMux.RUnlock()
		}
	}
}

// sendToClient sends a message to a specific client
func (s *SSEService) sendToClient(client *SSEClient, message SSEMessage) {
	select {
	case client.Channel <- s.formatSSEMessage(message):
	default:
		// Client channel is full, remove the client
		s.unregister <- client
	}
}

// formatSSEMessage formats a message for SSE transmission
func (s *SSEService) formatSSEMessage(message SSEMessage) []byte {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling SSE message: %v", err)
		return []byte("data: {\"error\": \"Failed to format message\"}\n\n")
	}

	return []byte(fmt.Sprintf("data: %s\n\n", string(data)))
}

// HandleSSEConnection handles new SSE connections
func (s *SSEService) HandleSSEConnection(c *gin.Context) {
	// Get milestone ID from URL parameter (changed from milestoneId to id for consistency)
	milestoneIDStr := c.Param("id")
	milestoneID, err := strconv.ParseUint(milestoneIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid milestone ID"})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// Create new client
	clientID := fmt.Sprintf("%s_%d_%d", c.ClientIP(), milestoneID, time.Now().UnixNano())
	client := &SSEClient{
		ID:          clientID,
		MilestoneID: uint(milestoneID),
		Channel:     make(chan []byte, 10),
		Request:     c.Request,
		Writer:      c.Writer,
	}

	// Register the client
	s.register <- client

	// Handle client disconnection
	defer func() {
		s.unregister <- client
	}()

	// Send messages to client
	c.Stream(func(w io.Writer) bool {
		select {
		case message := <-client.Channel:
			w.Write(message)
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}

// BroadcastMarketUpdate broadcasts market data updates
func (s *SSEService) BroadcastMarketUpdate(event MarketUpdateEvent) {
	message := SSEMessage{
		Type:      "market_update",
		Data:      event,
		Timestamp: time.Now().Unix(),
	}

	select {
	case s.broadcast <- message:
	default:
		log.Println("Warning: SSE broadcast channel is full")
	}
}

// BroadcastTradeUpdate broadcasts trade updates to clients watching specific milestone
func (s *SSEService) BroadcastTradeUpdate(milestoneID uint, optionID string, tradeData map[string]interface{}) {
	message := SSEMessage{
		Type:      "trade",
		Data:      tradeData,
		Timestamp: time.Now().Unix(),
	}

	select {
	case s.broadcast <- message:
	default:
		log.Println("Warning: SSE broadcast channel is full")
	}
}

// BroadcastOrderBookUpdate broadcasts order book updates to clients watching specific milestone
func (s *SSEService) BroadcastOrderBookUpdate(milestoneID uint, optionID string, orderBookData map[string]interface{}) {
	message := SSEMessage{
		Type:      "orderbook_update",
		Data:      orderBookData,
		Timestamp: time.Now().Unix(),
	}

	select {
	case s.broadcast <- message:
	default:
		log.Println("Warning: SSE broadcast channel is full")
	}
}

// BroadcastPriceChange broadcasts price changes to clients watching specific milestone
func (s *SSEService) BroadcastPriceChange(milestoneID uint, option string, oldPrice, newPrice float64) {
	priceChangeEvent := map[string]interface{}{
		"milestone_id": milestoneID,
		"option":       option,
		"old_price":    oldPrice,
		"new_price":    newPrice,
		"change":       newPrice - oldPrice,
		"change_pct":   func() float64 {
			if oldPrice > 0 {
				return ((newPrice - oldPrice) / oldPrice) * 100
			}
			return 0
		}(),
	}

	message := SSEMessage{
		Type:      "price_change",
		Data:      priceChangeEvent,
		Timestamp: time.Now().Unix(),
	}

	select {
	case s.broadcast <- message:
	default:
		log.Println("Warning: SSE broadcast channel is full")
	}
}

// GetConnectedClientsCount returns the number of connected clients
func (s *SSEService) GetConnectedClientsCount() int {
	s.clientsMux.RLock()
	defer s.clientsMux.RUnlock()
	return len(s.clients)
}

// GetClientsForMilestone returns the number of clients watching a specific milestone
func (s *SSEService) GetClientsForMilestone(milestoneID uint) int {
	s.clientsMux.RLock()
	defer s.clientsMux.RUnlock()

	count := 0
	for _, client := range s.clients {
		if client.MilestoneID == milestoneID {
			count++
		}
	}

	return count
}
