package dtos

import "github.com/Ruclo/MyMeals/internal/models"

type UpdateStatusRequest struct {
	Status models.OrderStatus `json:"status" binding:"required"`
}
