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
			// TODO: Invalid
			return
		}

		err = auth.SetCustomerTokenCookie(order.ID, c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set customer cookie"})
			return
		}

		c.JSON(http.StatusCreated, order)
	}
}

func (oh *OrdersHandler) PostOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIdStr := c.Param("orderID")
		if orderIdStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing orderID parameter"})
			return
		}
		orderId, err := strconv.ParseUint(orderIdStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid orderID parameter"})
			return
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
			// TODO: Not existing meal or order or whatever
			return
		}

		// StatusCreated?
		c.JSON(http.StatusOK, order)
	}
}

func (oh *OrdersHandler) PostOrderReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIdStr := c.Param("orderID")
		if orderIdStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing orderID parameter"})
			return
		}

		orderId, err := strconv.ParseUint(orderIdStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid orderID parameter"})
			return
		}
		var review models.Review
		err = c.ShouldBind(&review)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		photos := c.Request.MultipartForm.File["photos"]

		if len(photos) > models.MaxReviewPhotos {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Too many photos requested"})
		}

		var photoUrls []string

		for _, photo := range photos {
			result, err := oh.cloudinary.Upload.Upload(c, photo,
				uploader.UploadParams{Transformation: "c_limit,h_1920,w_1920"})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload photo"})
				return
			}

			photoUrls = append(photoUrls, result.SecureURL)
		}

		review.OrderID = uint(orderId)
		review.PhotoURLs = photoUrls
		if err := oh.orderRepository.AddReview(&review); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add review"})
			//TODO: Invalid order ?
			return
		}

		c.Status(http.StatusCreated)
	}
}

func (oh *OrdersHandler) UpdateStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIDStr := c.Param("orderID")
		if orderIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing orderID parameter"})
			return
		}
		orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid orderID parameter"})
			return
		}

		mealIDStr := c.Param("mealID")
		if mealIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing mealID parameter"})
			return
		}
		mealID, err := strconv.ParseUint(mealIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mealID parameter"})
			return
		}

		order, err := oh.orderRepository.MarkCompleted(uint(orderID), uint(mealID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
			// TODO: Non existant order meal ?
			return
		}

		c.JSON(http.StatusOK, order)
	}
}
