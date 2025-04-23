package services

import (
	"context"
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"github.com/Ruclo/MyMeals/internal/storage"
	"mime/multipart"
)

const MealPhotoSize = 1000

type MealService interface {
	Create(context.Context, *models.Meal, *multipart.FileHeader) error
	Update(context.Context, *models.Meal, *multipart.FileHeader) error
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

func (ms *mealService) Create(c context.Context,
	meal *models.Meal,
	photo *multipart.FileHeader) error {
	result, err := ms.imageStorage.UploadCropped(c, photo, MealPhotoSize, MealPhotoSize)

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

func (ms *mealService) Update(c context.Context, meal *models.Meal, photo *multipart.FileHeader) error {

	if photo != nil {
		result, err := ms.imageStorage.UploadCropped(c, photo, MealPhotoSize, MealPhotoSize)
		if err != nil {
			return errors.NewInternalServerErr("Failed to upload photo", err)
		}

		err = ms.imageStorage.Delete(c, meal.ImageURL)
		if err != nil {
			return errors.NewInternalServerErr("Failed to delete old photo", err)
		}

		meal.ImageURL = result.URL
	}

	return ms.mealRepository.Update(meal)
}

func (ms *mealService) Delete(id uint) error {
	return ms.mealRepository.Delete(&models.Meal{ID: id})
}
