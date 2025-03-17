package dtos

import "github.com/Ruclo/MyMeals/internal/models"

type OrderResponse struct {
	models.Order
	Jwt string
}
