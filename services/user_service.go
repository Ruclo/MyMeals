package services

import (
	"github.com/Ruclo/MyMeals/internal/auth"
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Create(user *models.StaffMember) error
	Login(ctx *gin.Context, user *models.StaffMember) error
	ChangePassword(username, oldPassword, newPassword string) error
}

type userService struct {
	userRepository repositories.UserRepository
}

func NewUserService(userRepository repositories.UserRepository) UserService {
	return &userService{userRepository: userRepository}
}

func (us *userService) Create(user *models.StaffMember) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.NewInternalServerErr("failed to generate a hashed password", err)
	}

	user.Password = string(hashedPassword)
	user.Role = models.AdminRole
	return us.userRepository.Create(user)
}

func (us *userService) Login(c *gin.Context, user *models.StaffMember) error {
	foundUser, err := us.userRepository.GetByUsername(user.Username)
	if err != nil {
		return errors.NewUnauthorizedErr("Failed to authorize", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(user.Password))
	if err != nil {
		return errors.NewUnauthorizedErr("Failed to authorize", err)
	}

	err = auth.SetStaffTokenCookie(foundUser.Username, foundUser.Role, c)
	if err != nil {
		return errors.NewInternalServerErr("Failed to create a cookie", err)
	}

	return nil
}

func (us *userService) ChangePassword(username, oldPassword, newPassword string) error {
	user, err := us.userRepository.GetByUsername(username)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return errors.NewUnauthorizedErr("Old password is incorrect", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.NewInternalServerErr("failed to generate a hashed password", err)
	}

	return us.userRepository.Update(&models.StaffMember{
		Username: username,
		Password: string(hashedPassword),
	})
}
