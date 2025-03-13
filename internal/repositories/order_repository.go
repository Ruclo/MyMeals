package repositories

import (
	"errors"
	"fmt"
	"github.com/Ruclo/MyMeals/internal/models"
	"gorm.io/gorm"
	"time"
)

type OrderRepository interface {
	GetAllPendingOrders() ([]*models.Order, error)
	GetOrders(olderThan time.Time, pageSize uint) ([]models.Order, error)
	Create(order *models.Order) error
	AddMealToOrder(orderID uint, meal *models.Meal) error
	PostReview(orderID uint, review *models.Review) error
}

type orderRepository struct {
	db *gorm.DB
}

func (r *orderRepository) GetAllPendingOrders() ([]*models.Order, error) {
	var orders []*models.Order

	err := r.db.
		Distinct("orders.*").
		Joins("JOIN order_meals ON orders.id = order_meals.order_id").
		Where("order_meals.status = ?", models.StatusPending).
		Preload("Meals").
		Find(&orders).Error

	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *orderRepository) GetOrders(olderThan time.Time, pageSize uint) ([]models.Order, error) {
	var orders []models.Order

	query := r.db.Model(&models.Order{})

	if !olderThan.IsZero() {
		query = query.Where("created_at < ?", olderThan)
	}

	if pageSize != 0 {
		err := query.Limit(int(pageSize)).Find(&orders).Error
		if err != nil {
			return nil, err
		}
	}
	query = query.Order("created_at DESC")
	query = query.Preload("Meals").Preload("Meals.Meal").Preload("Review")
	if err := query.Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) AddMealToOrder(orderID, mealID uint, quantity int) (*models.Order, error) {
	var existingOrderMeal models.OrderMeal

	err := r.db.Where("order_id = ? AND meal_id = ?", orderID, mealID).First(&existingOrderMeal).Error

	if err == nil {
		existingOrderMeal.Quantity += quantity
		if err = r.db.Save(&existingOrderMeal).Error; err != nil {
			return nil, err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		err = r.db.Create(&models.OrderMeal{
			OrderID:  orderID,
			MealID:   mealID,
			Quantity: quantity,
		}).Error

		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	var updatedOrder models.Order
	if err = r.db.Preload("Meals").Preload("Meals.Meal").
		First(&updatedOrder, orderID).Error; err != nil {
		return nil, err
	}

	return &updatedOrder, nil

}

func (r *orderRepository) PostReview(orderID uint, review *models.Review) error {
	var order models.Order

	if err := r.db.First(&order, orderID).Error; err != nil {
		return err
	}

	if order.Review != nil {
		return fmt.Errorf("order %d already has a review", orderID)
	}

	return r.db.Create(review).Error
}
