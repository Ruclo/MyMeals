package handlers

import (
	"github.com/Ruclo/MyMeals/internal/auth"
	"github.com/Ruclo/MyMeals/internal/dtos"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type UsersHandler struct {
	userRepository repositories.UserRepository
}

func NewUsersHandler(userRepository repositories.UserRepository) *UsersHandler {
	return &UsersHandler{userRepository: userRepository}
}

func (uh *UsersHandler) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.StaffMember

		err := c.ShouldBindJSON(&user)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		foundUser, err := uh.userRepository.GetByUsername(user.Username)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
			return
			//TODO: User doesnt exist
		}

		err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(user.Password))

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		err = auth.SetStaffTokenCookie(foundUser.Username, foundUser.Role, c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		user.Password = string(hashedPassword)
		user.Role = models.AdminRole
		err = uh.userRepository.Create(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		username := c.MustGet("username")

		user, err := uh.userRepository.GetByUsername(username.(string))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(changePasswordRequest.OldPassword))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(changePasswordRequest.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		}

		err = uh.userRepository.Update(&models.StaffMember{
			Username: username.(string),
			Password: string(hashedPassword),
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		}

		c.Status(http.StatusOK)
	}
}
