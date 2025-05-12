package handlers

import (
	"github.com/Ruclo/MyMeals/internal/auth"
	"github.com/Ruclo/MyMeals/internal/dtos"
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
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

		loggedUser, err := uh.userService.Login(user.Username, user.Password)
		if err != nil {
			c.Error(err)
			return
		}

		token, expiration, err := auth.GenerateStaffJWT(loggedUser.Username, loggedUser.Role)
		if err != nil {
			c.Error(err)
			return
		}

		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie("token", token, int(time.Until(expiration).Seconds()), "/", "", true, true)

		response := dtos.ModelToUserResponse(loggedUser)
		c.JSON(http.StatusOK, response)
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

		resp := dtos.ModelToUserResponse(&user)
		c.JSON(http.StatusCreated, resp)

	}
}

func (uh *UsersHandler) GetMe() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func (uh *UsersHandler) GetStaff() gin.HandlerFunc {
	return func(c *gin.Context) {
		staff, err := uh.userService.GetStaff()
		if err != nil {
			c.Error(err)
			return
		}

		var response []*dtos.UserResponse
		for _, user := range staff {
			response = append(response, dtos.ModelToUserResponse(user))
		}

		c.JSON(http.StatusOK, response)
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

func (uh *UsersHandler) DeleteUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Param("username")
		if username == "" {
			c.Error(errors.NewValidationErr("Invalid username", nil))
			return
		}

		err := uh.userService.DeleteUser(username)
		if err != nil {
			c.Error(err)
			return
		}
		c.Status(http.StatusNoContent)

	}
}
