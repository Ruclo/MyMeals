package repositories

import (
	"github.com/Ruclo/MyMeals/internal/models"
	"time"
)

type OrderQueryParams struct {
	OlderThan   time.Time
	PageSize    uint
	OnlyPending bool
}

type OrderRepository interface {
	WithTransaction(fn func(tx OrderRepository) error) error
	GetOrders(params OrderQueryParams) ([]*models.Order, error)
	GetByID(orderID uint) (*models.Order, error)
	Create(order *models.Order) error
	GetOrderMeal(orderID, mealID uint) (*models.OrderMeal, error)
	CreateOrderMeal(orderMeal *models.OrderMeal) error
	UpdateOrderMeal(orderMeal *models.OrderMeal) error
	CreateReview(review *models.Review) error
}
