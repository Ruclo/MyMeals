package handlers

import (
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type OrdersHandler struct {
	orderRepository repositories.OrderRepository
}

func NewOrdersHandler(orderRepository repositories.OrderRepository) *OrdersHandler {
	return &OrdersHandler{orderRepository: orderRepository}
}

func (oh *OrdersHandler) GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		pageSizeStr := c.DefaultQuery("pagesize", "10")
		pageSize, err := strconv.ParseUint(pageSizeStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagesize parameter"})
			return
		}

		olderThanStr := c.Query("olderThan")
		olderThan := time.Time{}

		if olderThanStr != "" {
			olderThan, err = time.Parse(time.RFC3339, olderThanStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid olderThan parameter"})
				return
			}
		}

		orders, err := oh.orderRepository.GetOrders(olderThan, uint(pageSize))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetched orders"})
			return
		}

		c.JSON(http.StatusOK, orders)

	}
}

func (oh *OrdersHandler) GetPendingOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		orders, err := oh.orderRepository.GetAllPendingOrders()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetched orders"})
			return
		}

		c.JSON(http.StatusOK, orders)
	}
}

func (oh *OrdersHandler) PostOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var order models.Order

		err := c.ShouldBindJSON(&order)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		err = oh.orderRepository.Create(&order)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
			return
		}

		c.JSON(http.StatusCreated, order)
	}
}

func (oh *OrdersHandler) PostOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIdStr := c.Query("orderID")
		if orderIdStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing orderID parameter"})
			return
		}
		orderId, err := strconv.ParseUint(orderIdStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid orderID parameter"})
		}

		var orderItem models.OrderMeal

		err = c.ShouldBindJSON(&orderItem)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		order, err := oh.orderRepository.AddMealToOrder(uint(orderId), orderItem.MealID, orderItem.Quantity)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add meal to order"})
		}

		// StatusCreated?
		c.JSON(http.StatusOK, order)
	}
}

func (oh *OrdersHandler) PostOrderReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIdStr := c.Query("orderID")
		if orderIdStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing orderID parameter"})
			return
		}

		orderId, err := strconv.ParseUint(orderIdStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid orderID parameter"})
		}
		var review models.Review
		err = c.ShouldBindJSON(&review)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		review.OrderID = uint(orderId)
		if err := oh.orderRepository.AddReview(&review); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add review"})
		}

		c.Status(http.StatusCreated)
	}
}
