package dtos

import (
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/shopspring/decimal"
)

type CreateMealRequest struct {
	Name        string              `form:"name" binding:"required, min=1"`
	Category    models.MealCategory `form:"category" binding:"required"`
	Description string              `form:"description" binding:"required, min=1"`
	Price       decimal.Decimal     `form:"price" binding:"required"`
}
