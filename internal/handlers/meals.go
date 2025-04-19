package handlers

import (
	"github.com/Ruclo/MyMeals/internal/dtos"
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type MealsHandler struct {
	mealService services.MealService
}

func NewMealsHandler(mealService services.MealService) *MealsHandler {
	return &MealsHandler{mealService: mealService}
}

func (mh *MealsHandler) GetMeals() gin.HandlerFunc {
	return func(c *gin.Context) {
		meals, err := mh.mealService.GetAll()
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, meals)
	}
}

func (mh *MealsHandler) PostMeal() gin.HandlerFunc {
	return func(c *gin.Context) {
		var createMealRequest dtos.CreateMealRequest
		if err := c.ShouldBind(&createMealRequest); err != nil {
			c.Error(errors.NewValidationErr("Invalid request", err))
			return
		}

		photo, err := c.FormFile("photo")
		if err != nil {
			c.Error(errors.NewValidationErr("photo not provided", err))
			return
		}

		var meal *models.Meal
		if meal, err = mh.mealService.Create(c, createMealRequest, photo); err != nil {
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
			c.Error(errors.NewValidationErr("invalid request", err))
			return
		}

		id := c.Param("mealID")
		idUint, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.Error(errors.NewValidationErr("Invalid meal id", err))
		}

		meal.ID = uint(idUint)
		err = mh.mealService.Update(&meal)

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
			c.Error(errors.NewValidationErr("id is invalid", nil))
			return
		}

		err = mh.mealService.Delete(uint(idUint))

		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Meal deleted"})

	}
}
