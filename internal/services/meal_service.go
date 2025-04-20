package services

import (
	"github.com/Ruclo/MyMeals/internal/dtos"
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"mime/multipart"
)

type MealService interface {
	Create(*gin.Context, dtos.CreateMealRequest, *multipart.FileHeader) (*models.Meal, error)
	Update(*models.Meal) error
	Delete(uint) error
	GetAll() ([]models.Meal, error)
}

type mealService struct {
	mealRepository repositories.MealRepository
	cloudinary     *cloudinary.Cloudinary
}

func NewMealService(mealRepository repositories.MealRepository, cloudinary *cloudinary.Cloudinary) MealService {
	return &mealService{
		mealRepository: mealRepository,
		cloudinary:     cloudinary,
	}
}

func (ms *mealService) Create(c *gin.Context,
	request dtos.CreateMealRequest,
	photo *multipart.FileHeader) (*models.Meal, error) {
	result, err := ms.cloudinary.Upload.Upload(c, photo,
		uploader.UploadParams{Transformation: "c_crop,h_1000,w_1000"})

	if err != nil {
		return nil, errors.NewInternalServerErr("Failed to upload photo", err)
	}

	meal := models.Meal{
		Name:        request.Name,
		Category:    request.Category,
		Description: request.Description,
		ImageURL:    result.SecureURL,
		Price:       request.Price,
	}

	if err = ms.mealRepository.Create(&meal); err != nil {
		ms.cloudinary.Upload.Destroy(c, uploader.DestroyParams{PublicID: result.PublicID})
		return nil, err
	}

	return &meal, nil
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
