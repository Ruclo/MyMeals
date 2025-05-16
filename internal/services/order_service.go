package services

import (
	"context"
	"fmt"
	"github.com/Ruclo/MyMeals/internal/apperrors"
	"github.com/Ruclo/MyMeals/internal/events"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"github.com/Ruclo/MyMeals/internal/storage"
	"mime/multipart"
	"time"
)

// OrderService defines operations for managing orders, adding meals, creating reviews, and marking orders as completed.
type OrderService interface {
	GetByID(id uint) (*models.Order, error)
	GetOrders(olderThan time.Time, pageSize uint) ([]*models.Order, error)
	GetAllPendingOrders() ([]*models.Order, error)
	Create(order *models.Order) error
	AddMealsToOrder(meals *[]models.OrderMeal) (*models.Order, error)
	CreateReview(c context.Context, review *models.Review, photos []*multipart.FileHeader) error
	MarkCompleted(orderID, mealID uint) (*models.Order, error)
}

type orderService struct {
	orderRepository  repositories.OrderRepository
	mealRepository   repositories.MealRepository
	imageStorage     storage.ImageStorage
	orderBroadcaster events.OrderBroadcaster
}

func NewOrderService(orderRepository repositories.OrderRepository,
	mealRepository repositories.MealRepository,
	imageStorage storage.ImageStorage,
	orderBroadcaster events.OrderBroadcaster) OrderService {
	return &orderService{
		orderRepository:  orderRepository,
		mealRepository:   mealRepository,
		imageStorage:     imageStorage,
		orderBroadcaster: orderBroadcaster,
	}
}

// GetByID retrieves an order by its unique identifier from the order repository.
// Returns the order or an error if not found.
func (os *orderService) GetByID(id uint) (*models.Order, error) {
	return os.orderRepository.GetByID(id)
}

// GetOrders retrieves a list of orders older than the specified time,
// with a maximum number of results determined by pageSize.
func (os *orderService) GetOrders(olderThan time.Time, pageSize uint) ([]*models.Order, error) {
	const MaximumPageSize = 100

	if pageSize > MaximumPageSize {
		pageSize = MaximumPageSize
	}

	params := repositories.OrderQueryParams{
		OlderThan:   olderThan,
		PageSize:    pageSize,
		OnlyPending: false,
	}

	return os.orderRepository.GetOrders(params)
}

// GetAllPendingOrders retrieves all orders with a pending status from the order repository.
func (os *orderService) GetAllPendingOrders() ([]*models.Order, error) {
	params := repositories.OrderQueryParams{
		OlderThan:   time.Time{},
		PageSize:    0,
		OnlyPending: true,
	}

	return os.orderRepository.GetOrders(params)
}

// Create handles the creation of a new order. Broadcasts the newly created order via OrderBroadcaster.
func (os *orderService) Create(order *models.Order) error {
	return os.orderRepository.WithTransaction(func(tx repositories.OrderRepository) error {
		err := tx.Create(order)
		if err != nil {
			return err
		}

		foundOrder, err := tx.GetByID(order.ID)
		if err != nil {
			return err
		}

		if err = os.orderBroadcaster.BroadcastOrder(order); err != nil {
			return err
		}

		*order = *foundOrder
		return nil
	})
}

// AddMealsToOrder adds one or more meals to an existing order, updating quantities if meals already exist in the order.
// It validates the existence of each meal and returns the updated order or an error in case of failure.
func (os *orderService) AddMealsToOrder(meals *[]models.OrderMeal) (*models.Order, error) {

	if len(*meals) == 0 {
		return nil, apperrors.NewValidationErr("No meals attached", nil)
	}

	for _, orderMeal := range *meals {
		_, err := os.mealRepository.GetByID(orderMeal.MealID)
		if err != nil {
			return nil, err
		}
	}

	var order *models.Order

	err := os.orderRepository.WithTransaction(func(tx repositories.OrderRepository) error {

		for _, orderMeal := range *meals {
			foundOrderMeal, err := tx.GetOrderMeal(orderMeal.OrderID, orderMeal.MealID)
			if err != nil && !apperrors.IsNotFoundErr(err) {
				return err
			}

			if foundOrderMeal != nil {
				foundOrderMeal.Quantity += orderMeal.Quantity
				err = tx.UpdateOrderMeal(foundOrderMeal)
				if err != nil {
					return err
				}
			} else {
				orderMeal.Completed = 0

				err = tx.CreateOrderMeal(&orderMeal)
				if err != nil {
					return err
				}
			}
		}
		foundOrder, err := tx.GetByID((*meals)[0].OrderID)
		if err != nil {
			return err
		}
		if err = os.orderBroadcaster.BroadcastOrder(foundOrder); err != nil {
			return err
		}
		order = foundOrder
		return nil
	})

	if err != nil {
		return nil, err
	}
	return order, nil
}

// CreateReview handles the creation of a review for a specified order, uploads photos
// and broadcasts the updated order.
func (os *orderService) CreateReview(c context.Context, review *models.Review, photos []*multipart.FileHeader) error {
	const MaxReviewPhotos = 3
	if len(photos) > MaxReviewPhotos {
		return apperrors.NewValidationErr("Too many review photos attached", nil)
	}

	order, err := os.orderRepository.GetByID(review.OrderID)
	if err != nil {
		return err
	}

	if order.Review != nil {
		return apperrors.NewAlreadyExistsErr("Order already has a review", nil)
	}

	var results []*storage.ImageResult

	for _, photo := range photos {
		var result *storage.ImageResult
		result, err = os.imageStorage.Upload(c, photo)
		if err != nil {
			break
		}

		results = append(results, result)
	}

	// Try to delete photos on error
	if err != nil {
		for _, result := range results {
			if os.imageStorage.Delete(c, result.PublicID) != nil {
				fmt.Println("Failed to delete photo with public ID:", result.PublicID)
			}
		}
		return err

	}

	var photoUrls []string

	for _, result := range results {
		photoUrls = append(photoUrls, result.URL)
	}

	review.PhotoURLs = photoUrls
	os.orderBroadcaster.BroadcastOrder(order) //:c
	return os.orderRepository.CreateReview(review)
}

// MarkCompleted marks an order meal as fully completed in terms of quantity and
// updates the associated order in the database.
func (os *orderService) MarkCompleted(orderID, mealID uint) (*models.Order, error) {

	var order *models.Order
	err := os.orderRepository.WithTransaction(func(tx repositories.OrderRepository) error {
		orderMeal, err := tx.GetOrderMeal(orderID, mealID)
		if err != nil {
			return err
		}

		orderMeal.Completed = orderMeal.Quantity

		err = tx.UpdateOrderMeal(orderMeal)
		if err != nil {
			return err
		}

		order, err = tx.GetByID(orderID)
		if err != nil {
			return err
		}
		if err = os.orderBroadcaster.BroadcastOrder(order); err != nil {
			return err
		}
		return nil

	})

	if err != nil {
		return nil, err
	}

	return order, err

}
