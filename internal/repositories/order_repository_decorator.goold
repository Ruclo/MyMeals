package repositories

import (
	"github.com/Ruclo/MyMeals/internal/models"
)

func NewBroadcastingOrderRepository(repo OrderRepository, ch chan models.Order) OrderRepository {
	return &orderRepositoryDecorator{
		repo: repo,
		ch:   ch,
	}
}

type orderRepositoryDecorator struct {
	repo OrderRepository
	ch   chan models.Order
}

func (d *orderRepositoryDecorator) GetOrders(params OrderQueryParams) ([]*models.Order, error) {
	return d.repo.GetOrders(params)
}

// Create implements OrderRepository interface with order broadcasting
func (d *orderRepositoryDecorator) Create(order *models.Order) error {
	err := d.repo.Create(order)
	if err == nil {
		d.ch <- *order
	}
	return err
}

// AddMealToOrder implements OrderRepository interface with order broadcasting
func (d *orderRepositoryDecorator) AddMealToOrder(orderID, mealID uint, quantity uint) (*models.Order, error) {
	order, err := d.repo.AddMealToOrder(orderID, mealID, quantity)
	if err == nil {
		d.ch <- *order
	}
	return order, err
}

// AddReview implements OrderRepository interface
func (d *orderRepositoryDecorator) AddReview(review *models.Review) error {
	return d.repo.CreateReview(review)
}

func (d *orderRepositoryDecorator) MarkCompleted(orderId, mealId uint) (*models.Order, error) {
	order, err := d.repo.MarkCompleted(orderId, mealId)
	if err == nil {
		d.ch <- *order
	}
	return order, err
}
