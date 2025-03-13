package dtos

import "github.com/Ruclo/MyMeals/internal/models"

type OrderDTO struct {
	ID      uint           `json:"id"`
	TableNo int            `json:"tableNo"`
	Name    string         `json:"name,omitempty"`  // Will be included in all responses but can be empty
	Notes   string         `json:"notes,omitempty"` // Will be included in all responses but can be empty
	Items   []OrderItemDTO `json:"items"`
	Review  *ReviewDTO     `json:"review,omitempty"` // Only included if exists
}

type OrderItemDTO struct {
	MealID   uint               `json:"mealID"`
	MealName string             `json:"mealName,omitempty"`
	Quantity int                `json:"quantity"`
	Status   models.OrderStatus `json:"status,omitempty"` // Always included, frontend can ignore if needed
}

type ReviewDTO struct {
	Rating    int      `json:"rating"`
	Comment   string   `json:"comment"`
	PhotoURLs []string `json:"photoURLs"`
}
