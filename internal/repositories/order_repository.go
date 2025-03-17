package repositories

import (
	"github.com/Ruclo/MyMeals/internal/models"
	"time"
)

type OrderRepository interface {
	GetAllPendingOrders() ([]*models.Order, error)
	GetOrders(olderThan time.Time, pageSize uint) ([]*models.Order, error)
	Create(order *models.Order) error
	AddMealToOrder(orderID, mealID uint, quantity int) (*models.Order, error)
	AddReview(review *models.Review) error
	UpdateStatus(orderId, mealID uint, status models.OrderStatus) (*models.Order, error)
}
