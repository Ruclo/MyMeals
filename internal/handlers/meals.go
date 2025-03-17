package handlers

import (
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type MealsHandler struct {
	mealRepository repositories.MealRepository
}

func NewMealsHandler(mealRepository repositories.MealRepository) *MealsHandler {
	return &MealsHandler{mealRepository: mealRepository}
}

func (mh *MealsHandler) GetMeals() gin.HandlerFunc {
	return func(c *gin.Context) {
		meals, err := mh.mealRepository.GetAll()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetched meals"})
			return
		}

		c.JSON(http.StatusOK, meals)
	}
}

func (mh *MealsHandler) PostMeal() gin.HandlerFunc {
	return func(c *gin.Context) {
		var meal models.Meal
		if err := c.ShouldBindJSON(&meal); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		err := mh.mealRepository.Create(&meal)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create meal"})
			//violations
			return
		}

		c.JSON(http.StatusCreated, meal)
	}
}

func (mh *MealsHandler) PutMeal() gin.HandlerFunc {
	return func(c *gin.Context) {
		var meal models.Meal
		if err := c.ShouldBindJSON(&meal); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		id := c.Param("mealID")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id parameter"})
			return
		}

		idUint, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id parameter"})
		}
		meal.ID = uint(idUint)

		err = mh.mealRepository.Update(&meal)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update meal"})
			//TODO: Check Invalid Request
			return
		}

		c.JSON(http.StatusOK, meal)
	}
}

func (mh *MealsHandler) DeleteMeal() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("mealID")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing id parameter"})
			return
		}

		idUint, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id parameter"})
			return
		}
		mealId := uint(idUint)

		err = mh.mealRepository.Delete(&models.Meal{ID: mealId})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete meal"})
			return
			//TODO: Check invalid meal
		}

		c.JSON(http.StatusOK, gin.H{"message": "Meal deleted"})

	}
}
