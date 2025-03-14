package repositories

import (
	"errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"gorm.io/gorm"
	"time"
)

type OrderRepository interface {
	GetAllPendingOrders() ([]*models.Order, error)
	GetOrders(olderThan time.Time, pageSize uint) ([]*models.Order, error)
	Create(order *models.Order) error
	AddMealToOrder(orderID, mealID uint, quantity int) (*models.Order, error)
	AddReview(review *models.Review) error
	UpdateStatus(orderId, mealID uint, status models.OrderStatus) error
}

func NewOrderRepository(db *gorm.DB) *OrderRepositoryImpl {
	return &OrderRepositoryImpl{db: db}
}

type OrderRepositoryImpl struct {
	db *gorm.DB
}

func (r *OrderRepositoryImpl) GetAllPendingOrders() ([]*models.Order, error) {
	var orders []*models.Order

	err := r.db.
		Distinct("orders.*").
		Joins("JOIN order_meals ON orders.id = order_meals.order_id").
		Where("order_meals.status = ?", models.StatusPending).
		Preload("OrderMeals").
		Preload("OrderMeals.Meal").
		Find(&orders).Error

	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *OrderRepositoryImpl) GetOrders(olderThan time.Time, pageSize uint) ([]*models.Order, error) {
	var orders []*models.Order

	query := r.db.Model(&models.Order{})

	if !olderThan.IsZero() {
		query = query.Where("created_at < ?", olderThan)
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	err := query.Limit(int(pageSize)).Find(&orders).Error
	if err != nil {
		return nil, err
	}

	query = query.Order("created_at DESC")
	query = query.Preload("OrderMeals").Preload("OrderMeals.Meal").Preload("Review")
	if err := query.Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepositoryImpl) Create(order *models.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(order).Error
	})

}

func (r *OrderRepositoryImpl) AddMealToOrder(orderID, mealID uint, quantity int) (*models.Order, error) {
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
	if err = r.db.Preload("OrderMeals").
		First(&updatedOrder, orderID).Error; err != nil {
		return nil, err
	}

	return &updatedOrder, nil

}

func (r *OrderRepositoryImpl) AddReview(review *models.Review) error {
	return r.db.Create(review).Error
}

func (r *OrderRepositoryImpl) UpdateStatus(orderId, mealId uint, status models.OrderStatus) error {
	return r.db.Model(&models.OrderMeal{}).Where("order_id = ? AND meal_id = ?", orderId, mealId).Update("status", status).Error
}
