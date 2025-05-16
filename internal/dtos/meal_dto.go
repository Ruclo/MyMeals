package dtos

import (
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/shopspring/decimal"
)

type CreateMealRequest struct {
	Name        string              `form:"name" binding:"required,min=1"`
	Category    models.MealCategory `form:"category" binding:"required"`
	Description string              `form:"description" binding:"required,min=1"`
	Price       decimal.Decimal     `form:"price" binding:"required"`
}

func (req *CreateMealRequest) ToModel() *models.Meal {
	return &models.Meal{
		Name:        req.Name,
		Category:    req.Category,
		Description: req.Description,
		Price:       req.Price,
	}
}

type MealResponse struct {
	ID          uint                `json:"id"`
	Name        string              `json:"name"`
	Category    models.MealCategory `json:"category"`
	Description string              `json:"description"`
	ImageURL    string              `json:"image_url"`
	Price       decimal.Decimal     `json:"price"`
}

func ToMealResponse(meal *models.Meal) *MealResponse {
	return &MealResponse{
		ID:          meal.ID,
		Name:        meal.Name,
		Category:    meal.Category,
		Description: meal.Description,
		ImageURL:    meal.ImageURL,
		Price:       meal.Price,
	}
}

func ToMealResponses(meals []*models.Meal) []*MealResponse {
	var result []*MealResponse
	for _, m := range meals {
		result = append(result, ToMealResponse(m))
	}
	return result
}
