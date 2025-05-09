package dtos

import "github.com/Ruclo/MyMeals/internal/models"

type UserResponse struct {
	Username string      `json:"username"`
	Role     models.Role `json:"role"`
}

func ModelToUserResponse(user *models.StaffMember) *UserResponse {
	return &UserResponse{
		Username: user.Username,
		Role:     user.Role,
	}
}
