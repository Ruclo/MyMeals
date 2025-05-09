package dtos

import (
	"github.com/Ruclo/MyMeals/internal/models"
	"time"
)

type CreateOrderRequest struct {
	TableNo int                `json:"table_no" binding:"required,gte=1"`
	Notes   string             `json:"notes"`
	Items   []OrderMealRequest `json:"items" binding:"required"`
}

type OrderMealRequest struct {
	MealID   uint `json:"meal_id" binding:"required"`
	Quantity uint `json:"quantity" binding:"required,gte=1"`
}

func (req *CreateOrderRequest) ToModel() *models.Order {
	order := &models.Order{
		TableNo:    req.TableNo,
		Notes:      req.Notes,
		OrderMeals: make([]models.OrderMeal, len(req.Items)),
	}

	for i, mealDTO := range req.Items {
		order.OrderMeals[i] = models.OrderMeal{
			MealID:    mealDTO.MealID,
			Quantity:  mealDTO.Quantity,
			Completed: 0, // Initialize as 0
		}
	}

	return order
}

type OrderResponse struct {
	ID        uint                `json:"id"`
	TableNo   int                 `json:"table_no"`
	Notes     string              `json:"notes"`
	CreatedAt time.Time           `json:"created_at"`
	Items     []OrderMealResponse `json:"items"`
	Review    *models.Review      `json:"review,omitempty"`
}

type OrderMealResponse struct {
	MealID    uint   `json:"meal_id"`
	MealName  string `json:"meal_name"`
	Quantity  uint   `json:"quantity"`
	Completed uint   `json:"completed"`
}

func ToOrderResponse(order *models.Order) *OrderResponse {
	orderResponse := &OrderResponse{
		ID:        order.ID,
		TableNo:   order.TableNo,
		Notes:     order.Notes,
		CreatedAt: order.CreatedAt,
		Items:     make([]OrderMealResponse, len(order.OrderMeals)),
	}

	for i, orderMeal := range order.OrderMeals {
		orderResponse.Items[i] = *ToOrderMealResponse(&orderMeal)
	}

	return orderResponse
}

func ToOrderMealResponse(orderMeal *models.OrderMeal) *OrderMealResponse {
	return &OrderMealResponse{
		MealID:    orderMeal.MealID,
		MealName:  orderMeal.Meal.Name,
		Quantity:  orderMeal.Quantity,
		Completed: orderMeal.Completed,
	}
}

func ToOrderReponseList(orders []*models.Order) []*OrderResponse {
	orderResponses := make([]*OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = ToOrderResponse(order)
	}
	return orderResponses
}
