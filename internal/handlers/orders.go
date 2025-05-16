package handlers

import (
	"github.com/Ruclo/MyMeals/internal/apperrors"
	"github.com/Ruclo/MyMeals/internal/auth"
	"github.com/Ruclo/MyMeals/internal/dtos"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

// OrdersHandler handles HTTP requests related to order operations.
type OrdersHandler struct {
	orderService services.OrderService
}

func NewOrdersHandler(orderService services.OrderService) *OrdersHandler {
	return &OrdersHandler{orderService: orderService}
}

// GetMyOrder handles HTTP GET requests to retrieve the details of users specific order
// from based on the order ID in their JWT.
func (oh *OrdersHandler) GetMyOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIDstr, exists := c.Get("orderID")
		if !exists {
			c.Error(apperrors.NewUnauthorizedErr("You are not authenticated", nil))
			return
		}
		orderID, err := strconv.ParseUint(orderIDstr.(string), 10, 64)
		if err != nil {
			c.Error(apperrors.NewValidationErr("Invalid order id", err))
			return
		}

		order, err := oh.orderService.GetByID(uint(orderID))
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, dtos.ToOrderResponse(order))
	}
}

// GetOrders handles HTTP GET requests to retrieve a list of orders
// supports cursor based pagination based on the createdAt timestamp.
func (oh *OrdersHandler) GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		pageSizeStr := c.DefaultQuery("pageSize", "10")
		pageSize, err := strconv.ParseUint(pageSizeStr, 10, 32)
		if err != nil {
			c.Error(apperrors.NewValidationErr("Invalid page size", err))
			return
		}

		olderThanStr := c.Query("olderThan")
		olderThan := time.Time{}
		if olderThanStr != "" {
			olderThan, err = time.Parse(time.RFC3339, olderThanStr)
			if err != nil {
				c.Error(apperrors.NewValidationErr("Invalid older than argument", err))
				return
			}
		}

		orders, err := oh.orderService.GetOrders(olderThan, uint(pageSize))
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, dtos.ToOrderReponseList(orders))

	}
}

// GetPendingOrders handles HTTP GET requests to retrieve a list of all pending orders.
func (oh *OrdersHandler) GetPendingOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		orders, err := oh.orderService.GetAllPendingOrders()

		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, dtos.ToOrderReponseList(orders))
	}
}

// PostOrder handles HTTP POST requests to create a new order.
// Includes a cookie in the response, which is used to further authorize the creator of the order.
func (oh *OrdersHandler) PostOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request dtos.CreateOrderRequest

		err := c.ShouldBindJSON(&request)
		if err != nil {
			c.Error(apperrors.NewValidationErr("Invalid request", err))
			return
		}

		order := request.ToModel()

		err = oh.orderService.Create(order)
		if err != nil {
			c.Error(err)
			return
		}

		token, expirationTime, err := auth.GenerateCustomerJWT(order.ID)
		if err != nil {
			c.Error(err)
			return
		}

		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie("token", token, int(time.Until(expirationTime).Seconds()),
			"/", "", true, true)

		c.JSON(http.StatusCreated, dtos.ToOrderResponse(order))
	}
}

// PostOrderItems handles HTTP POST requests to add meals to an existing order
// based on the provided order ID and order items.
func (oh *OrdersHandler) PostOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIdStr := c.Param("orderID")

		orderId, err := strconv.ParseUint(orderIdStr, 10, 64)
		if err != nil {
			c.Error(apperrors.NewValidationErr("Invalid order id", err))
			return
		}

		var orderItems []dtos.OrderMealRequest

		err = c.ShouldBindJSON(&orderItems)
		if err != nil {
			c.Error(apperrors.NewValidationErr("Invalid request", err))
			return
		}

		var orderMeals []models.OrderMeal

		for _, orderItem := range orderItems {
			orderMeals = append(orderMeals, models.OrderMeal{
				OrderID:  uint(orderId),
				MealID:   orderItem.MealID,
				Quantity: orderItem.Quantity,
			})
		}

		order, err := oh.orderService.AddMealsToOrder(&orderMeals)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, dtos.ToOrderResponse(order))
	}
}

// PostOrderReview handles HTTP POST requests for submitting a review for a specific order with optional photo uploads.
func (oh *OrdersHandler) PostOrderReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIdStr := c.Param("orderID")

		orderId, err := strconv.ParseUint(orderIdStr, 10, 64)
		if err != nil {
			c.Error(apperrors.NewValidationErr("Invalid order id", err))
			return
		}
		var reviewRequest dtos.ReviewRequest
		err = c.ShouldBind(&reviewRequest)
		if err != nil {
			c.Error(apperrors.NewValidationErr("Invalid request", err))
			return
		}

		review := reviewRequest.ToModel()
		review.OrderID = uint(orderId)

		photos := c.Request.MultipartForm.File["photos"]

		if err = oh.orderService.CreateReview(c, review, photos); err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusCreated)
	}
}

// UpdateStatus handles HTTP POST requests to mark a specific meal within an order
// as completed based on order and meal ID.
func (oh *OrdersHandler) UpdateStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIDStr := c.Param("orderID")

		orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
		if err != nil {
			c.Error(apperrors.NewValidationErr("Invalid order id", err))
			return
		}

		mealIDStr := c.Param("mealID")

		mealID, err := strconv.ParseUint(mealIDStr, 10, 64)
		if err != nil {
			c.Error(apperrors.NewValidationErr("Invalid meal id", err))
			return
		}

		order, err := oh.orderService.MarkCompleted(uint(orderID), uint(mealID))
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, dtos.ToOrderResponse(order))
	}
}
