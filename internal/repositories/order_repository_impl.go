package repositories

import (
	"errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"gorm.io/gorm"
	"time"
)

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepositoryImpl{db: db}
}

type orderRepositoryImpl struct {
	db *gorm.DB
}

func (r *orderRepositoryImpl) GetAllPendingOrders() ([]*models.Order, error) {
	var orders []*models.Order

	err := r.db.
		Distinct("orders.*").
		Joins("JOIN order_meals ON orders.id = order_meals.order_id").
		Where("order_meals.completed < order_meals.quantity ").
		Preload("OrderMeals.Meal").
		Find(&orders).Error

	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *orderRepositoryImpl) GetOrders(olderThan time.Time, pageSize uint) ([]*models.Order, error) {
	var orders []*models.Order

	query := r.db.Model(&models.Order{})

	if !olderThan.IsZero() {
		query = query.Where("created_at < ?", olderThan)
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	query = query.Limit(int(pageSize))

	query = query.Order("created_at DESC")
	query = query.Preload("OrderMeals.Meal").Preload("Review")
	if err := query.Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepositoryImpl) Create(order *models.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(order).Error
		if err != nil {
			return err
		}

		err = tx.Preload("OrderMeals.Meal").First(order, order.ID).Error
		if err != nil {
			return err
		}
		return nil
	})

}

func (r *orderRepositoryImpl) AddMealToOrder(orderID, mealID uint, quantity uint) (*models.Order, error) {
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
	if err = r.db.Preload("OrderMeals.Meal").
		First(&updatedOrder, orderID).Error; err != nil {
		return nil, err
	}

	return &updatedOrder, nil

}

func (r *orderRepositoryImpl) AddReview(review *models.Review) error {
	return r.db.Create(review).Error
}

func (r *orderRepositoryImpl) MarkCompleted(orderId, mealId uint) (*models.Order, error) {
	var orderMeal models.OrderMeal
	err := r.db.Where("order_id = ? AND meal_id = ?", orderId, mealId).First(&orderMeal).Error
	if err != nil {
		return nil, err
	}

	orderMeal.Completed = orderMeal.Quantity
	err = r.db.Model(orderMeal).Updates(orderMeal).Error

	if err != nil {
		return nil, err
	}

	var order models.Order

	if err = r.db.Preload("OrderMeals.Meal").First(&order, orderId).Error; err != nil {
		return nil, err
	}

	return &order, nil
}
