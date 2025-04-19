package dtos

import (
	"github.com/Ruclo/MyMeals/internal/models"
	"time"
)

type CreateOrderRequest struct {
	TableNo    int                `json:"table_no" binding:"required, gte=1"`
	Notes      string             `json:"notes"`
	OrderMeals []OrderMealRequest `json:"items" binding:"required"`
}

type OrderMealRequest struct {
	MealID   uint `json:"meal_id" binding:"required"`
	Quantity uint `json:"quantity" binding:"required, gte=1"`
}

func (req *CreateOrderRequest) ToModel() *models.Order {
	order := &models.Order{
		TableNo:    req.TableNo,
		Notes:      req.Notes,
		OrderMeals: make([]models.OrderMeal, len(req.OrderMeals)),
	}

	for i, mealDTO := range req.OrderMeals {
		order.OrderMeals[i] = models.OrderMeal{
			MealID:    mealDTO.MealID,
			Quantity:  mealDTO.Quantity,
			Completed: 0, // Initialize as 0
		}
	}

	return order
}

type OrderResponse struct {
	ID         uint                `json:"id"`
	TableNo    int                 `json:"table_no"`
	Notes      string              `json:"notes"`
	CreatedAt  time.Time           `json:"created_at"`
	OrderMeals []OrderMealResponse `json:"order_meals"`
	Review     *models.Review      `json:"review,omitempty"`
}

type OrderMealResponse struct {
	MealID    uint   `json:"meal_id"`
	MealName  string `json:"meal_name"`
	Quantity  uint   `json:"quantity"`
	Completed uint   `json:"completed"`
}
