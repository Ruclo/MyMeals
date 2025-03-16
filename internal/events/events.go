package events

import (
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/gin-gonic/gin"
	"io"
	"log"
)

type OrderEvents struct {
	// Events are pushed to this channel by the main events-gathering routine
	Message ClientChan

	// New client connections
	NewClients chan ClientChan

	// Closed client connections
	ClosedClients chan ClientChan

	// Total client connections
	TotalClients map[ClientChan]bool
}

func (events *OrderEvents) listen() {
	for {
		select {
		// Add new available client
		case client := <-events.NewClients:
			events.TotalClients[client] = true
			log.Printf("Client added. %d registered clients", len(events.TotalClients))

		// Remove closed client
		case client := <-events.ClosedClients:
			delete(events.TotalClients, client)
			close(client)
			log.Printf("Removed client. %d registered clients", len(events.TotalClients))

		// Broadcast message to client
		case order := <-events.Message:
			for clientMessageChan := range events.TotalClients {
				clientMessageChan <- order
			}
		}
	}
}

type ClientChan chan models.Order

func NewServer() (events *OrderEvents) {
	events = &OrderEvents{
		Message:       make(ClientChan),
		NewClients:    make(chan ClientChan),
		ClosedClients: make(chan ClientChan),
		TotalClients:  make(map[ClientChan]bool),
	}

	go events.listen()

	return
}

func (events *OrderEvents) Handler() gin.HandlersChain {
	return gin.HandlersChain{headersMiddleware, events.clientConnectMiddleware, handler}
}

func (events *OrderEvents) clientConnectMiddleware(c *gin.Context) {
	// Initialize client channel
	clientChan := make(ClientChan)

	// Send new connection to event server
	events.NewClients <- clientChan

	defer func() {
		// Drain client channel so that it does not block. Server may keep sending messages to this channel
		go func() {
			for range clientChan {
			}
		}()
		// Send closed connection to event server
		events.ClosedClients <- clientChan
	}()

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
	clientChan := c.MustGet("clientChan").(ClientChan)
	c.Stream(func(w io.Writer) bool {
		if order, ok := <-clientChan; ok {
			c.SSEvent("order", order)
			return true
		}
		return false
	})
}
