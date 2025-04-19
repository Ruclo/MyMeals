package handlers

import (
	"github.com/Ruclo/MyMeals/internal/auth"
	"github.com/Ruclo/MyMeals/internal/dtos"
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type OrdersHandler struct {
	orderService services.OrderService
}

func NewOrdersHandler(orderService services.OrderService) *OrdersHandler {
	return &OrdersHandler{orderService: orderService}
}

func (oh *OrdersHandler) GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		pageSizeStr := c.DefaultQuery("pagesize", "10")
		pageSize, err := strconv.ParseUint(pageSizeStr, 10, 32)
		if err != nil {
			c.Error(errors.NewValidationErr("Invalid page size", err))
			return
		}

		olderThanStr := c.Query("olderThan")
		olderThan := time.Time{}
		if olderThanStr != "" {
			olderThan, err = time.Parse(time.RFC3339, olderThanStr)
			if err != nil {
				c.Error(errors.NewValidationErr("Invalid older than argument", err))
				return
			}
		}

		orders, err := oh.orderService.GetOrders(olderThan, uint(pageSize))

		// TODO: to dtos
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, orders)

	}
}

func (oh *OrdersHandler) GetPendingOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		orders, err := oh.orderService.GetAllPendingOrders()
		//TODO: to dtos
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, orders)
	}
}

func (oh *OrdersHandler) PostOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request dtos.CreateOrderRequest

		err := c.ShouldBindJSON(&request)
		if err != nil {
			c.Error(errors.NewValidationErr("Invalid request", err))
			return
		}

		order := request.ToModel()

		err = oh.orderService.Create(order)
		if err != nil {
			c.Error(err)
			return
		}

		err = auth.SetCustomerTokenCookie(order.ID, c)
		if err != nil {
			c.Error(errors.NewInternalServerErr("Failed to set the cookie", err))
			return
		}
		//TODO: to DTO
		c.JSON(http.StatusCreated, order)
	}
}

func (oh *OrdersHandler) PostOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIdStr := c.Param("orderID")

		orderId, err := strconv.ParseUint(orderIdStr, 10, 64)
		if err != nil {
			c.Error(errors.NewValidationErr("Invalid order id", err))
			return
		}

		var orderItem dtos.OrderMealRequest

		err = c.ShouldBindJSON(&orderItem)
		if err != nil {
			c.Error(errors.NewValidationErr("Invalid request", err))
			return
		}

		order, err := oh.orderService.AddMealToOrder(uint(orderId), orderItem.MealID, orderItem.Quantity)
		if err != nil {
			c.Error(err)
			return
		}

		// TODO: status created ?, DTO
		c.JSON(http.StatusOK, order)
	}
}

func (oh *OrdersHandler) PostOrderReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIdStr := c.Param("orderID")

		orderId, err := strconv.ParseUint(orderIdStr, 10, 64)
		if err != nil {
			c.Error(errors.NewValidationErr("Invalid order id", err))
			return
		}
		var review models.Review
		err = c.ShouldBind(&review)
		if err != nil {
			c.Error(err)
			return
		}

		review.OrderID = uint(orderId)

		photos := c.Request.MultipartForm.File["photos"]

		if err = oh.orderService.CreateReview(c, &review, photos); err != nil {
			c.Error(err)
			return
		}
		// TODO DTO
		c.Status(http.StatusCreated)
	}
}

func (oh *OrdersHandler) UpdateStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIDStr := c.Param("orderID")

		orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
		if err != nil {
			c.Error(errors.NewValidationErr("Invalid order id", err))
			return
		}

		mealIDStr := c.Param("mealID")

		mealID, err := strconv.ParseUint(mealIDStr, 10, 64)
		if err != nil {
			c.Error(errors.NewValidationErr("Invalid meal id", err))
			return
		}

		order, err := oh.orderService.MarkCompleted(uint(orderID), uint(mealID))
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, order)
	}
}
