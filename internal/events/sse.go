package events

import (
	"github.com/Ruclo/MyMeals/internal/dtos"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/gin-gonic/gin"
	"io"
	"log"
)

type OrderChan chan *dtos.OrderResponse

type SSEServer struct {
	// Events are pushed to this channel by the main events-gathering routine
	broadcast OrderChan

	// New client connections
	register chan OrderChan

	// Closed client connections
	unregister chan OrderChan

	// Total client connections
	clients map[OrderChan]bool
}

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
			log.Println("New client connected")
			log.Println(s.clients)
			s.clients[client] = true

		// Remove closed client
		case client := <-s.unregister:
			delete(s.clients, client)
			close(client)

		// Broadcast message to client
		case order := <-s.broadcast:
			for client := range s.clients {
				select {
				case client <- order:

				default:
					log.Println("Failed to send message to client, disconnnecting client")
					s.unregister <- client
				}

			}
		}
	}
}

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

type sseOrderBroadcaster struct {
	broadcastChan OrderChan
}

// BroadcastOrder sends an order to the SSE message channel
func (b *sseOrderBroadcaster) BroadcastOrder(order *models.Order) error {
	b.broadcastChan <- dtos.ToOrderResponse(order)
	return nil
}
