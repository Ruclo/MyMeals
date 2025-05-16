package handlers

import (
	"github.com/Ruclo/MyMeals/internal/apperrors"
	"github.com/Ruclo/MyMeals/internal/auth"
	"github.com/Ruclo/MyMeals/internal/dtos"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// UsersHandler handles HTTP requests related to staff member actions such as
// login, user management, and password changes.
type UsersHandler struct {
	userService services.UserService
}

func NewUsersHandler(userService services.UserService) *UsersHandler {
	return &UsersHandler{userService: userService}
}

// Login handles the HTTP POST request to log in.
// It includes a JWT used for further authentication in a response cookie.
func (uh *UsersHandler) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User

		err := c.ShouldBindJSON(&user)

		if err != nil {
			c.Error(apperrors.NewValidationErr("invalid request", err))
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
		c.SetCookie("token", token, int(time.Until(expiration).Seconds()),
			"/", "", true, true)

		c.JSON(http.StatusOK, dtos.ModelToUserResponse(loggedUser))
	}
}

// PostUser handles the HTTP POST request to create a new user.
func (uh *UsersHandler) PostUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User

		err := c.ShouldBindJSON(&user)
		if err != nil {
			c.Error(apperrors.NewValidationErr("Invalid request", err))
			return
		}

		err = uh.userService.Create(&user)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusCreated, dtos.ModelToUserResponse(&user))

	}
}

// GetMe handles the HTTP GET request to retrieve information about the authenticated user.
func (uh *UsersHandler) GetMe() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.Error(apperrors.NewUnauthorizedErr("You are not authenticated", nil))
			return
		}
		user, err := uh.userService.GetByUsername(username.(string))
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, dtos.ModelToUserResponse(user))
	}
}

// GetStaff handles the HTTP GET request to retrieve a list of all staff members.
func (uh *UsersHandler) GetStaff() gin.HandlerFunc {
	return func(c *gin.Context) {
		staff, err := uh.userService.GetStaff()
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, dtos.ModelToUserResponses(staff))
	}
}

// ChangePassword handles the HTTP PUT request for updating the authenticated user's password.
func (uh *UsersHandler) ChangePassword() gin.HandlerFunc {
	return func(c *gin.Context) {

		var changePasswordRequest dtos.ChangePasswordRequest

		err := c.ShouldBindJSON(&changePasswordRequest)
		if err != nil {
			c.Error(apperrors.NewValidationErr("Invalid request", err))
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

// DeleteUser handles the HTTP DELETE request to delete a user by username.
func (uh *UsersHandler) DeleteUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Param("username")
		if username == "" {
			c.Error(apperrors.NewValidationErr("Invalid username", nil))
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
