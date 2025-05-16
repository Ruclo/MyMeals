package dtos

import "github.com/Ruclo/MyMeals/internal/models"

type UserResponse struct {
	Username string      `json:"username"`
	Role     models.Role `json:"role"`
}

func ModelToUserResponse(user *models.User) *UserResponse {
	return &UserResponse{
		Username: user.Username,
		Role:     user.Role,
	}
}

func ModelToUserResponses(users []*models.User) []*UserResponse {
	var result []*UserResponse
	for _, u := range users {
		result = append(result, ModelToUserResponse(u))
	}
	return result
}
