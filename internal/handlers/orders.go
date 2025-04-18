package handlers

import (
	"github.com/Ruclo/MyMeals/internal/auth"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type OrdersHandler struct {
	orderRepository repositories.OrderRepository
	cloudinary      *cloudinary.Cloudinary
}

func NewOrdersHandler(orderRepository repositories.OrderRepository, cloudinary *cloudinary.Cloudinary) *OrdersHandler {
	return &OrdersHandler{orderRepository: orderRepository, cloudinary: cloudinary}
}

func (oh *OrdersHandler) GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		pageSizeStr := c.DefaultQuery("pagesize", "10")
		pageSize, err := strconv.ParseUint(pageSizeStr, 10, 32)
		if err != nil {
			c.Error(err)
			return
		}

		olderThanStr := c.Query("olderThan")
		olderThan := time.Time{}
		if olderThanStr != "" {
			olderThan, err = time.Parse(time.RFC3339, olderThanStr)
			if err != nil {
				c.Error(err)
				return
			}
		}

		orders, err := oh.orderRepository.GetOrders(olderThan, uint(pageSize))

		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, orders)

	}
}

func (oh *OrdersHandler) GetPendingOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		orders, err := oh.orderRepository.GetAllPendingOrders()

		if err != nil {
			c.Error(err)
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
			c.Error(err)
			return
		}

		err = oh.orderRepository.Create(&order)
		if err != nil {
			c.Error(err)
			return
		}

		err = auth.SetCustomerTokenCookie(order.ID, c)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusCreated, order)
	}
}

func (oh *OrdersHandler) PostOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIdStr := c.Param("orderID")

		orderId, err := strconv.ParseUint(orderIdStr, 10, 64)
		if err != nil {
			c.Error(err)
			return
		}

		var orderItem models.OrderMeal

		err = c.ShouldBindJSON(&orderItem)
		if err != nil {
			c.Error(err)
			return
		}

		order, err := oh.orderRepository.AddMealToOrder(uint(orderId), orderItem.MealID, orderItem.Quantity)
		if err != nil {
			c.Error(err)
			return
		}

		// StatusCreated?
		c.JSON(http.StatusOK, order)
	}
}

func (oh *OrdersHandler) PostOrderReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIdStr := c.Param("orderID")

		orderId, err := strconv.ParseUint(orderIdStr, 10, 64)
		if err != nil {
			c.Error(err)
			return
		}
		var review models.Review
		err = c.ShouldBind(&review)
		if err != nil {
			c.Error(err)
			return
		}

		photos := c.Request.MultipartForm.File["photos"]

		if len(photos) > models.MaxReviewPhotos {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Too many photos requested"}) // TODO ?
		}

		var photoUrls []string

		for _, photo := range photos {
			result, err := oh.cloudinary.Upload.Upload(c, photo,
				uploader.UploadParams{Transformation: "c_limit,h_1920,w_1920"})
			if err != nil {
				c.Error(err)
				return
			}

			photoUrls = append(photoUrls, result.SecureURL)
		}

		review.OrderID = uint(orderId)
		review.PhotoURLs = photoUrls
		if err := oh.orderRepository.AddReview(&review); err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusCreated)
	}
}

func (oh *OrdersHandler) UpdateStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIDStr := c.Param("orderID")

		orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
		if err != nil {
			c.Error(err)
			return
		}

		mealIDStr := c.Param("mealID")

		mealID, err := strconv.ParseUint(mealIDStr, 10, 64)
		if err != nil {
			c.Error(err)
			return
		}

		order, err := oh.orderRepository.MarkCompleted(uint(orderID), uint(mealID))
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, order)
	}
}
