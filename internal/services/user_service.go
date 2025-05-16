package services

import (
	"github.com/Ruclo/MyMeals/internal/apperrors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

// UserService defines methods for managing users such as creation, deletion, login and password management.
type UserService interface {
	GetByUsername(username string) (*models.User, error)
	Create(user *models.User) error
	Login(username, password string) (*models.User, error)
	ChangePassword(username, oldPassword, newPassword string) error
	GetStaff() ([]*models.User, error)
	DeleteUser(username string) error
}

type userService struct {
	userRepository repositories.UserRepository
}

func NewUserService(userRepository repositories.UserRepository) UserService {
	return &userService{userRepository: userRepository}
}

// Create attempts to add a new user to the repository, hashing the password and assigning the default role.
func (us *userService) Create(user *models.User) error {

	exists, err := us.userRepository.Exists(user.Username)
	if err != nil {
		return err
	}
	if exists {
		return apperrors.NewAlreadyExistsErr("User already exists", nil)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return apperrors.NewInternalServerErr("failed to generate a hashed password", err)
	}

	user.Password = string(hashedPassword)
	user.Role = models.RegularStaffRole
	return us.userRepository.Create(user)
}

// Login authenticates a user by verifying the provided username and password, returning the user or an error.
func (us *userService) Login(username, password string) (*models.User, error) {
	foundUser, err := us.userRepository.GetByUsername(username)
	if err != nil {
		return nil, apperrors.NewUnauthorizedErr("Failed to authorize", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(password))
	if err != nil {
		return nil, apperrors.NewUnauthorizedErr("Failed to authorize", err)
	}

	return foundUser, nil

}

// GetByUsername retrieves a user by their username from the repository. Returns the user or an error if not found.
func (us *userService) GetByUsername(username string) (*models.User, error) {
	return us.userRepository.GetByUsername(username)
}

// ChangePassword updates a user's password after verifying the provided old password.
// Returns an error if validation fails.
func (us *userService) ChangePassword(username, oldPassword, newPassword string) error {
	user, err := us.userRepository.GetByUsername(username)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return apperrors.NewUnauthorizedErr("Old password is incorrect", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperrors.NewInternalServerErr("failed to generate a hashed password", err)
	}

	return us.userRepository.Update(&models.User{
		Username: username,
		Password: string(hashedPassword),
	})
}

// GetStaff retrieves all users with the role of Regular Staff from the user repository.
// Returns a slice of users or an error.
func (us *userService) GetStaff() ([]*models.User, error) {
	return us.userRepository.GetByRole(models.RegularStaffRole)
}

// DeleteUser removes a user by their username if they exist, returning an error if the user is not found or on failure.
func (us *userService) DeleteUser(username string) error {
	exists, err := us.userRepository.Exists(username)
	if err != nil {
		return err
	}

	if !exists {
		return apperrors.NewNotFoundErr("User not found", nil)
	}

	return us.userRepository.DeleteByUsername(username)
}
