package services

import (
	"context"
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/events"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"github.com/Ruclo/MyMeals/internal/storage"
	"mime/multipart"
	"time"
)

type OrderService interface {
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

func (os *orderService) GetAllPendingOrders() ([]*models.Order, error) {
	params := repositories.OrderQueryParams{
		OlderThan:   time.Time{},
		PageSize:    0,
		OnlyPending: true,
	}

	return os.orderRepository.GetOrders(params)
}

func (os *orderService) Create(order *models.Order) error {
	err := os.orderRepository.WithTransaction(func(tx repositories.OrderRepository) error {
		err := tx.Create(order)
		if err != nil {
			return err
		}

		foundOrder, err := tx.GetByID(order.ID)
		if err != nil {
			return err
		}
		*order = *foundOrder
		return err
	})
	if err != nil {
		return nil
	}

	// No errors for now
	os.orderBroadcaster.BroadcastOrder(order)
	return nil
}

func (os *orderService) AddMealsToOrder(meals *[]models.OrderMeal) (*models.Order, error) {

	if len(*meals) == 0 {
		return nil, errors.NewValidationErr("No meals attached", nil)
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
			if err != nil && !errors.IsNotFoundErr(err) {
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
		var err error
		order, err = tx.GetByID((*meals)[0].OrderID)
		return err
	})

	if err != nil {
		return nil, err
	}

	os.orderBroadcaster.BroadcastOrder(order)
	return order, nil
}

func (os *orderService) CreateReview(c context.Context, review *models.Review, photos []*multipart.FileHeader) error {
	const MaxReviewPhotos = 3
	if len(photos) > MaxReviewPhotos {
		return errors.NewValidationErr("Too many review photos attached", nil)
	}

	order, err := os.orderRepository.GetByID(review.OrderID)
	if err != nil {
		return err
	}

	if order.Review != nil {
		return errors.NewDuplicateErr("Order already has a review", nil)
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

	if err != nil {
		for _, result := range results {
			os.imageStorage.Delete(c, result.PublicID)
		}
		return err

	}

	var photoUrls []string

	for _, result := range results {
		photoUrls = append(photoUrls, result.URL)
	}

	review.PhotoURLs = photoUrls
	os.orderBroadcaster.BroadcastOrder(order)
	return os.orderRepository.CreateReview(review)
}

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
		return err

	})

	if err != nil {
		return nil, err
	}

	os.orderBroadcaster.BroadcastOrder(order)
	return order, err

}
