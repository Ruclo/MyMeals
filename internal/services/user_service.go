package services

import (
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Create(user *models.StaffMember) error
	Login(username, password string) (*models.StaffMember, error)
	ChangePassword(username, oldPassword, newPassword string) error
}

type userService struct {
	userRepository repositories.UserRepository
}

func NewUserService(userRepository repositories.UserRepository) UserService {
	return &userService{userRepository: userRepository}
}

func (us *userService) Create(user *models.StaffMember) error {

	// Check if user with username already exists
	exists, err := us.userRepository.Exists(user.Username)
	if err != nil {
		return err
	}

	if exists {
		return errors.NewAlreadyExistsErr("User already exists", nil)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.NewInternalServerErr("failed to generate a hashed password", err)
	}

	user.Password = string(hashedPassword)
	user.Role = models.RegularStaffRole
	return us.userRepository.Create(user)
}

func (us *userService) Login(username, password string) (*models.StaffMember, error) {
	foundUser, err := us.userRepository.GetByUsername(username)
	if err != nil {
		return nil, errors.NewUnauthorizedErr("Failed to authorize", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(password))
	if err != nil {
		return nil, errors.NewUnauthorizedErr("Failed to authorize", err)
	}

	return foundUser, nil

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
