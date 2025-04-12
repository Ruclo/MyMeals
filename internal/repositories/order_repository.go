package repositories

import (
	"github.com/Ruclo/MyMeals/internal/models"
	"time"
)

type OrderRepository interface {
	GetAllPendingOrders() ([]*models.Order, error)
	GetOrders(olderThan time.Time, pageSize uint) ([]*models.Order, error)
	Create(order *models.Order) error
	AddMealToOrder(orderID, mealID uint, quantity uint) (*models.Order, error)
	AddReview(review *models.Review) error
	MarkCompleted(orderId, mealID uint) (*models.Order, error)
}
