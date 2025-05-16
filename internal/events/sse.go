package events

import (
	"github.com/Ruclo/MyMeals/internal/dtos"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/gin-gonic/gin"
	"io"
)

// OrderChan is a channel for sending Orders.
type OrderChan chan *dtos.OrderResponse

// SSEServer is a server-sent events implementation that manages client connections and broadcasts messages.
// It maintains channels for broadcasting events, registering new clients, and unregistering disconnected clients.
type SSEServer struct {
	// The orders sent to this channel get broadcasted to all clients connected to the SSE.
	broadcast OrderChan

	// New client connections
	register chan OrderChan

	// Closed client connections
	unregister chan OrderChan

	// Total client connections
	clients map[OrderChan]bool
}

// NewSSEServer initializes and returns a new SSEServer instance. It starts a goroutine which manages these channels.
func NewSSEServer() *SSEServer {
	server := &SSEServer{
		broadcast:  make(OrderChan),
		register:   make(chan OrderChan),
		unregister: make(chan OrderChan),
		clients:    make(map[OrderChan]bool),
	}

	go server.listen()
	return server
}

func (s *SSEServer) NewBroadcaster() OrderBroadcaster {
	return &sseOrderBroadcaster{broadcastChan: s.broadcast}
}

func (s *SSEServer) listen() {
	for {
		select {
		// Add new available client
		case client := <-s.register:
			s.clients[client] = true

		// Remove closed client
		case client := <-s.unregister:
			delete(s.clients, client)
			close(client)

		// Broadcast message to clients
		case order := <-s.broadcast:
			for client := range s.clients {
				select {
				case client <- order:

				default:
					s.unregister <- client
				}

			}
		}
	}
}

// Handler returns a gin.HandlersChain consisting of middlewares and a handler for managing SSE connections.
func (s *SSEServer) Handler() gin.HandlersChain {
	return gin.HandlersChain{headersMiddleware, s.clientConnectMiddleware, handler}
}

func (s *SSEServer) clientConnectMiddleware(c *gin.Context) {
	// Initialize client channel
	clientChan := make(OrderChan)

	// Send new connection to event server
	s.register <- clientChan

	defer func() {
		for range clientChan {
		}
		s.unregister <- clientChan
	}()
	// Send closed connection to event server

	c.Set("clientChan", clientChan)

	c.Next()

}

func headersMiddleware(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Next()
}

func handler(c *gin.Context) {
	clientChan := c.MustGet("clientChan").(OrderChan)
	done := c.Request.Context().Done()
	c.Stream(func(w io.Writer) bool {
		select {
		case <-done:
			return false
		case order, ok := <-clientChan:
			if !ok {
				return false
			}
			c.SSEvent("message", order)
			return true

		}
	})
}

// sseOrderBroadcaster implements the OrderBroadcaster interface
type sseOrderBroadcaster struct {
	broadcastChan OrderChan
}

// BroadcastOrder sends an order to the SSE message channel
func (b *sseOrderBroadcaster) BroadcastOrder(order *models.Order) error {
	b.broadcastChan <- dtos.ToOrderResponse(order)
	return nil
}
