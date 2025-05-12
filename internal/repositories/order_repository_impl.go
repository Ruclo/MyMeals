package repositories

import (
	stdErrors "errors"
	"fmt"
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"gorm.io/gorm"
)

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepositoryImpl{db: db}
}

type orderRepositoryImpl struct {
	db *gorm.DB
}

func (r *orderRepositoryImpl) WithTransaction(fn func(txRepo OrderRepository) error) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return errors.NewInternalServerErr("Failed to start a transaction", tx.Error)
	}
	defer tx.Rollback()

	txRepo := &orderRepositoryImpl{db: tx}

	if err := fn(txRepo); err != nil {
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return errors.NewInternalServerErr("Failed to commit transaction", err)
	}

	return nil
}

func (r *orderRepositoryImpl) GetByID(orderID uint) (*models.Order, error) {
	var order models.Order

	err := r.db.Model(&models.Order{}).Where("ID = ?", orderID).
		Preload("OrderMeals.Meal").
		Preload("Review").First(&order).Error

	if err == nil {
		return &order, nil
	}

	if stdErrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NewNotFoundErr(fmt.Sprintf("Order with id %d not found", orderID), err)
	}

	return nil, errors.NewInternalServerErr(fmt.Sprintf("Failed to get order with ID %d", orderID), err)
}

func (r *orderRepositoryImpl) GetOrders(params OrderQueryParams) ([]*models.Order, error) {
	var orders []*models.Order

	query := r.db.Model(&models.Order{})

	if params.OnlyPending {
		query = query.Distinct("orders.*").
			Joins("JOIN order_meals ON orders.id = order_meals.order_id").
			Where("order_meals.completed < order_meals.quantity")
	}

	if !params.OlderThan.IsZero() {
		query = query.Where("created_at < ?", params.OlderThan)
	}

	if params.PageSize > 0 {
		query = query.Limit(int(params.PageSize))
	}

	query = query.Order("created_at DESC")

	query = query.Preload("OrderMeals").Preload("Review")

	if err := query.Find(&orders).Error; err != nil {
		return nil, errors.NewInternalServerErr(fmt.Sprintf("Failed to get orders with params %+v", params), nil)
	}

	return orders, nil
}

func (r *orderRepositoryImpl) Create(order *models.Order) error {
	err := r.db.Create(order).Error
	if err == nil {
		return nil
	}

	return errors.NewInternalServerErr(fmt.Sprintf("Failed to create order %+v", order), nil)
}

func (r *orderRepositoryImpl) GetOrderMeal(orderID, mealID uint) (*models.OrderMeal, error) {
	var orderMeal models.OrderMeal

	err := r.db.Model(&models.OrderMeal{}).Where("order_id = ? AND meal_id = ?", orderID, mealID).First(&orderMeal).Error
	if err == nil {
		return &orderMeal, nil
	}

	if stdErrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NewNotFoundErr(fmt.Sprintf("Order %d does not have meal with id %d", orderID, mealID), err)
	}

	return nil, errors.NewInternalServerErr(fmt.Sprintf("Failed to get order meal order id: %d meal id: %d", orderID, mealID), err)
}

func (r *orderRepositoryImpl) CreateOrderMeal(orderMeal *models.OrderMeal) error {
	err := r.db.Create(orderMeal).Error
	if err == nil {
		return nil
	}

	return errors.NewInternalServerErr(fmt.Sprintf("Failed to create order meal %+v", orderMeal), err)

}

func (r *orderRepositoryImpl) UpdateOrderMeal(orderMeal *models.OrderMeal) error {
	res := r.db.Model(orderMeal).Updates(orderMeal)
	if res.Error != nil {
		return errors.NewInternalServerErr(fmt.Sprintf("Failed to update order meal %+v", orderMeal), res.Error)
	}
	if res.RowsAffected == 0 {
		return errors.NewNotFoundErr(fmt.Sprintf("Order meal %+v not found", orderMeal), res.Error)
	}

	return nil
}

func (r *orderRepositoryImpl) CreateReview(review *models.Review) error {
	err := r.db.Create(review).Error

	if err == nil {
		return nil
	}

	return errors.NewInternalServerErr(fmt.Sprintf("Failed to create a review %+v", review), nil)
}
