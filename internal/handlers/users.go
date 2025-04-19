package handlers

import (
	"github.com/Ruclo/MyMeals/internal/dtos"
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UsersHandler struct {
	userService services.UserService
}

func NewUsersHandler(userService services.UserService) *UsersHandler {
	return &UsersHandler{userService: userService}
}

func (uh *UsersHandler) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.StaffMember

		err := c.ShouldBindJSON(&user)

		if err != nil {
			c.Error(errors.NewValidationErr("invalid request", err))
			return
		}

		err = uh.userService.Login(c, &user)
		if err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusOK)
	}
}

func (uh *UsersHandler) PostUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.StaffMember

		err := c.ShouldBindJSON(&user)
		if err != nil {
			c.Error(errors.NewValidationErr("Invalid request", err))
			return
		}

		err = uh.userService.Create(&user)
		if err != nil {
			c.Error(err)
			return
		}
		c.Status(http.StatusCreated)
	}
}

func (uh *UsersHandler) ChangePassword() gin.HandlerFunc {
	return func(c *gin.Context) {

		var changePasswordRequest dtos.ChangePasswordRequest

		err := c.ShouldBindJSON(&changePasswordRequest)
		if err != nil {
			c.Error(errors.NewValidationErr("Invalid request", err))
			return
		}

		username := c.MustGet("username")

		err = uh.userService.ChangePassword(username.(string),
			changePasswordRequest.OldPassword,
			changePasswordRequest.NewPassword)

		if err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusOK)
	}
}
