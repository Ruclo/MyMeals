package services

import (
	stdErrors "errors"
	"fmt"
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"github.com/Ruclo/MyMeals/internal/storage"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mime/multipart"
	"time"
)

type OrderService interface {
	GetOrders(olderThan time.Time, pageSize uint) ([]*models.Order, error)
	GetAllPendingOrders() ([]*models.Order, error)
	Create(order *models.Order) error
	AddMealToOrder(orderID, mealID, quantity uint) (*models.Order, error)
	CreateReview(c *gin.Context, review *models.Review, photos []*multipart.FileHeader) error
	MarkCompleted(orderID, mealID uint) (*models.Order, error)
}

type orderService struct {
	orderRepository repositories.OrderRepository
	mealRepository  repositories.MealRepository
	imageStorage    storage.ImageStorage
}

func NewOrderService(orderRepository repositories.OrderRepository,
	mealRepository repositories.MealRepository,
	imageStorage storage.ImageStorage) OrderService {
	return &orderService{
		orderRepository: orderRepository,
		mealRepository:  mealRepository,
		imageStorage:    imageStorage,
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
	return os.orderRepository.Create(order)
	//TODO: broadcast
}

func (os *orderService) AddMealToOrder(orderID, mealID, quantity uint) (*models.Order, error) {

	_, err := os.mealRepository.GetByID(mealID)
	if err != nil {
		return nil, err
	}

	var order *models.Order

	err = os.orderRepository.WithTransaction(func(tx repositories.OrderRepository) error {

		orderMeal, err := tx.GetOrderMeal(orderID, mealID)
		if err != nil && !stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NewInternalServerErr(fmt.Sprintf("Failed to get order meal with order id %d and meal id %d", orderID, mealID), err)
		}

		if orderMeal != nil {
			orderMeal.Quantity += quantity
			return tx.UpdateOrderMeal(orderMeal)
		}

		orderMeal = &models.OrderMeal{
			OrderID:   orderID,
			MealID:    mealID,
			Quantity:  quantity,
			Completed: 0,
		}

		err = tx.CreateOrderMeal(orderMeal)
		if err != nil {
			return err
		}

		order, err = tx.GetByID(orderID)
		return err
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}

func (os *orderService) CreateReview(c *gin.Context, review *models.Review, photos []*multipart.FileHeader) error {
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

	return order, err

}
