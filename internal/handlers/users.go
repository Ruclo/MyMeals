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
			//TODO: User doesnt exist
		}

		err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(user.Password))

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		}

		jwt, err := auth.GenerateStaffToken(foundUser.Username, foundUser.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		}

		c.JSON(http.StatusOK, gin.H{"token": jwt})
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

		username, exists := c.Get("username")

		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get username"})
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
