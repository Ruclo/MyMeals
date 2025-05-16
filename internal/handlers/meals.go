package handlers

import (
	"errors"
	"github.com/Ruclo/MyMeals/internal/apperrors"
	"github.com/Ruclo/MyMeals/internal/dtos"
	"github.com/Ruclo/MyMeals/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// MealsHandler handles HTTP requests related to meal operations.
type MealsHandler struct {
	mealService services.MealService
}

func NewMealsHandler(mealService services.MealService) *MealsHandler {
	return &MealsHandler{mealService: mealService}
}

// GetMeals handles the HTTP GET request to retrieve all meals and returns them as a JSON response.
func (mh *MealsHandler) GetMeals() gin.HandlerFunc {
	return func(c *gin.Context) {
		meals, err := mh.mealService.GetAll()
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, dtos.ToMealResponses(meals))
	}
}

// GetMealsWithDeleted handles the HTTP GET request to retrieve all meals, including soft deleted ones
// and returns them as a JSON response.
func (mh *MealsHandler) GetMealsWithDeleted() gin.HandlerFunc {
	return func(c *gin.Context) {
		meals, err := mh.mealService.GetAllWithDeleted()
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, dtos.ToMealResponses(meals))
	}
}

// PostMeal handles the HTTP POST request to create a new meal with the provided details and photo.
func (mh *MealsHandler) PostMeal() gin.HandlerFunc {
	return func(c *gin.Context) {
		var createMealRequest dtos.CreateMealRequest
		if err := c.ShouldBind(&createMealRequest); err != nil {
			c.Error(apperrors.NewValidationErr("Invalid request", err))
			return
		}

		photo, err := c.FormFile("photo")
		if err != nil {
			c.Error(apperrors.NewValidationErr("photo not provided", err))
			return
		}

		meal := createMealRequest.ToModel()

		if err = mh.mealService.Create(c, meal, photo); err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusCreated, dtos.ToMealResponse(meal))
	}
}

// PostMealReplace handles the HTTP POST request to replace an existing meal
// by its ID with new details and an optional photo.
func (mh *MealsHandler) PostMealReplace() gin.HandlerFunc {
	return func(c *gin.Context) {
		var mealRequest dtos.CreateMealRequest
		if err := c.ShouldBind(&mealRequest); err != nil {
			c.Error(apperrors.NewValidationErr("invalid request", err))
			return
		}

		id := c.Param("mealID")
		idUint, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.Error(apperrors.NewValidationErr("Invalid meal id", err))
			return
		}

		photo, err := c.FormFile("photo")
		if err != nil && !errors.Is(err, http.ErrMissingFile) {
			c.Error(apperrors.NewValidationErr("error processing the photo", err))
			return
		}

		meal := mealRequest.ToModel()
		meal.ID = uint(idUint)

		err = mh.mealService.Replace(c, meal, photo)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, dtos.ToMealResponse(meal))
	}
}

// DeleteMeal handles the HTTP DELETE request to remove a meal by its ID.
func (mh *MealsHandler) DeleteMeal() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("mealID")

		idUint, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.Error(apperrors.NewValidationErr("id is invalid", nil))
			return
		}

		err = mh.mealService.Delete(uint(idUint))

		if err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusNoContent)

	}
}
