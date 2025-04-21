package services

import (
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"github.com/Ruclo/MyMeals/internal/storage"
	"github.com/gin-gonic/gin"
	"mime/multipart"
)

type MealService interface {
	Create(*gin.Context, *models.Meal, *multipart.FileHeader) error
	Update(*models.Meal) error
	Delete(uint) error
	GetAll() ([]models.Meal, error)
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

func (ms *mealService) Create(c *gin.Context,
	meal *models.Meal,
	photo *multipart.FileHeader) error {
	result, err := ms.imageStorage.UploadCropped(c, photo, 1000, 1000)

	if err != nil {
		return errors.NewInternalServerErr("Failed to upload photo", err)
	}

	meal.ImageURL = result.URL

	if err = ms.mealRepository.Create(meal); err != nil {
		ms.imageStorage.Delete(c, result.PublicID)
		return err
	}

	return nil
}

func (ms *mealService) GetAll() ([]models.Meal, error) {
	return ms.mealRepository.GetAll()
}

func (ms *mealService) Update(meal *models.Meal) error {
	return ms.mealRepository.Update(meal)
}

func (ms *mealService) Delete(id uint) error {
	return ms.mealRepository.Delete(&models.Meal{ID: id})
}
