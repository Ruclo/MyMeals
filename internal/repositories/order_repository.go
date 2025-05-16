package repositories

import (
	"github.com/Ruclo/MyMeals/internal/models"
	"time"
)

// OrderQueryParams defines parameters for querying orders in the data store.
type OrderQueryParams struct {
	OlderThan   time.Time
	PageSize    uint
	OnlyPending bool
}

// OrderRepository provides an interface for CRUD operations on Order and OrderMeal entities.
// WithTransaction executes a function within a database transaction.
// GetOrders retrieves a list of orders based on the specified query parameters.
// GetByID fetches a single order by its unique identifier.
// Create adds a new order to the data store.
// GetOrderMeal retrieves a specific meal associated with an order.
// CreateOrderMeal adds a new meal to an order in the data store.
// UpdateOrderMeal updates an existing meal tied to an order.
// CreateReview creates a new review associated with an order.
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
