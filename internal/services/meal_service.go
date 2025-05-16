package services

import (
	"context"
	"github.com/Ruclo/MyMeals/internal/apperrors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"github.com/Ruclo/MyMeals/internal/storage"
	"mime/multipart"
)

const MealPhotoSize = 1000

// MealService defines an interface for managing meal operations, including creation, updating, deletion, and retrieval.
type MealService interface {
	Create(context.Context, *models.Meal, *multipart.FileHeader) error
	Replace(context.Context, *models.Meal, *multipart.FileHeader) error
	Delete(uint) error
	GetAll() ([]*models.Meal, error)
	GetAllWithDeleted() ([]*models.Meal, error)
}

type mealService struct {
	mealRepository repositories.MealRepository
	imageStorage   storage.ImageStorage
}

func NewMealService(mealRepository repositories.MealRepository, imageStorage storage.ImageStorage) MealService {
	return &mealService{
		mealRepository: mealRepository,
		imageStorage:   imageStorage,
	}
}

// Create uploads a meal photo, sets the meal's image URL, and stores the meal in the database.
func (ms *mealService) Create(c context.Context,
	meal *models.Meal,
	photo *multipart.FileHeader) error {
	result, err := ms.imageStorage.UploadCropped(c, photo, MealPhotoSize, MealPhotoSize)

	if err != nil {
		return apperrors.NewInternalServerErr("Failed to upload photo", err)
	}

	meal.ImageURL = result.URL

	if err = ms.mealRepository.Create(meal); err != nil {
		ms.imageStorage.Delete(c, result.PublicID)
		return err
	}

	return nil
}

// GetAll retrieves all meal records from the repository and returns them along with any encountered apperrors.
func (ms *mealService) GetAll() ([]*models.Meal, error) {
	return ms.mealRepository.GetAll()
}

// GetAllWithDeleted retrieves all Meal records, including those that have been soft-deleted, from the repository.
func (ms *mealService) GetAllWithDeleted() ([]*models.Meal, error) {
	return ms.mealRepository.GetAllWithDeleted()
}

// Replace modifies an existing meal, optionally updates its image, and replaces it in the repository
// within a transaction context.
// The old meal gets soft deleted. A new meal gets created.
func (ms *mealService) Replace(c context.Context, meal *models.Meal, photo *multipart.FileHeader) error {

	existingMeal, err := ms.mealRepository.GetByID(meal.ID)
	if err != nil {
		return err
	}

	newMeal := models.Meal{
		Name:        meal.Name,
		Price:       meal.Price,
		Description: meal.Description,
		Category:    meal.Category,
		ImageURL:    existingMeal.ImageURL,
	}

	err = ms.mealRepository.WithTransaction(func(tx repositories.MealRepository) error {
		if photo != nil {
			result, err := ms.imageStorage.UploadCropped(c, photo, MealPhotoSize, MealPhotoSize)
			if err != nil {
				return err
			}

			err = ms.imageStorage.Delete(c, meal.ImageURL)
			if err != nil {
				return err
			}

			newMeal.ImageURL = result.URL
		}

		if err := tx.Create(&newMeal); err != nil {
			return err
		}

		return tx.Delete(existingMeal)

	})

	*meal = newMeal
	return err
}

// Delete performs a soft delete of a meal identified by the given ID using the meal repository.
// Returns an error if any occurs.
func (ms *mealService) Delete(id uint) error {
	return ms.mealRepository.Delete(&models.Meal{ID: id})
}
