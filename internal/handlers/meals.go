package handlers

import (
	stdErrors "errors"
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type MealsHandler struct {
	mealRepository repositories.MealRepository
	cloudinary     *cloudinary.Cloudinary
}

func NewMealsHandler(mealRepository repositories.MealRepository, cloudinary *cloudinary.Cloudinary) *MealsHandler {
	return &MealsHandler{mealRepository: mealRepository, cloudinary: cloudinary}
}

func (mh *MealsHandler) GetMeals() gin.HandlerFunc {
	return func(c *gin.Context) {
		meals, err := mh.mealRepository.GetAll()

		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, meals)
	}
}

func (mh *MealsHandler) PostMeal() gin.HandlerFunc {
	return func(c *gin.Context) {
		var meal models.Meal
		if err := c.ShouldBind(&meal); err != nil {
			c.Error(err)
			return
		}

		photo, err := c.FormFile("photo")
		if err != nil {
			c.Error(err)
			return
		}

		result, err := mh.cloudinary.Upload.Upload(c, photo,
			uploader.UploadParams{Transformation: "c_crop,h_1000,w_1000"})
		if err != nil {
			c.Error(err)
			return
		}

		meal.ImageURL = result.SecureURL

		err = mh.mealRepository.Create(&meal)

		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusCreated, meal)
	}
}

func (mh *MealsHandler) PutMeal() gin.HandlerFunc {
	return func(c *gin.Context) {
		var meal models.Meal
		if err := c.ShouldBindJSON(&meal); err != nil {
			c.Error(err)
			return
		}

		id := c.Param("mealID")

		idUint, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.Error(err)
		}
		meal.ID = uint(idUint)

		err = mh.mealRepository.Update(&meal)

		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, meal)
	}
}

func (mh *MealsHandler) DeleteMeal() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("mealID")

		idUint, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			err := stdErrors.New("id is invalid")
			c.Error(errors.NewValidationErr(err.Error(), err))
			return
		}
		mealId := uint(idUint)

		err = mh.mealRepository.Delete(&models.Meal{ID: mealId})

		if err != nil {
			c.Error(err)
			return
			//TODO: Check invalid meal
		}

		c.JSON(http.StatusOK, gin.H{"message": "Meal deleted"})

	}
}
