package events_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Ruclo/MyMeals/internal/events"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSSEServerHandlerAndBroadcaster(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create an SSE server
	server := events.NewSSEServer()
	broadcaster := server.NewBroadcaster()

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Add the SSE handler to the router
	r.GET("/events", server.Handler()...)

	// Create a test server

	// Create a test order
	order := &models.Order{ID: 123, TableNo: 5}

	// Create a channel to track received events
	receivedChan := make(chan bool, 1)

	// Start a client to connect to the SSE endpoint
	go func() {
		// Connect to the SSE endpoint
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/events", ts.URL), nil)
		if err != nil {
			t.Logf("Error creating SSE request: %v", err)
			return
		}

		client := ts.Client()
		resp, err := client.Do(req)

		if err != nil {
			t.Logf("Error connecting to SSE: %v", err)
			return
		}
		defer resp.Body.Close()

		// Read the event stream
		scanner := bufio.NewScanner(resp.Body)
		var eventData string
		for scanner.Scan() {
			text := scanner.Text()
			if strings.HasPrefix(text, "data:") {
				eventData = strings.TrimPrefix(text, "data:")
				break
			}
		}

		// Parse the order JSON
		var receivedOrder models.Order
		err = json.Unmarshal([]byte(eventData), &receivedOrder)
		assert.NoError(t, err)
		assert.Equal(t, order.ID, receivedOrder.ID)

		receivedChan <- true
		fmt.Println("ending")
	}()

	// Give the client time to connect
	time.Sleep(100 * time.Millisecond)

	// Broadcast the order
	err := broadcaster.BroadcastOrder(order)
	assert.NoError(t, err)

	// Wait for the order to be received or timeout
	select {
	case <-receivedChan:
		println("Received order")
		break
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for broadcast order to be received")
	}

	ts.CloseClientConnections()
}
