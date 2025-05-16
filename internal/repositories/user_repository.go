package repositories

import (
	"errors"
	"fmt"
	"github.com/Ruclo/MyMeals/internal/apperrors"
	"github.com/Ruclo/MyMeals/internal/models"
	"gorm.io/gorm"
)

// UserRepository provides an interface for CRUD operations on User entities and supports transactional operations.
// WithTransaction executes the given function within a database transaction.
// GetByUsername retrieves a User by their username.
// GetByRole retrieves all StaffMembers with a specific role.
// Create persists a new User to the database.
// Update updates an existing User in the database.
// Exists checks if a User with the specified username exists.
// DeleteByUsername removes a User from the database by their username.
type UserRepository interface {
	WithTransaction(fn func(txRepo UserRepository) error) error
	GetByUsername(username string) (*models.User, error)
	GetByRole(role models.Role) ([]*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	Exists(username string) (bool, error)
	DeleteByUsername(username string) error
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

type userRepositoryImpl struct {
	db *gorm.DB
}

func (r *userRepositoryImpl) WithTransaction(fn func(txRepo UserRepository) error) error {
	tx := r.db.Begin()

	if tx.Error != nil {
		return apperrors.NewInternalServerErr("Failed to start a transaction", tx.Error)
	}
	defer tx.Rollback()

	txRepo := &userRepositoryImpl{db: tx}

	if err := fn(txRepo); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return apperrors.NewInternalServerErr("Failed to commit transaction", err)
	}

	return nil
}

func (r *userRepositoryImpl) Exists(username string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).
		Where("username = ?", username).
		Count(&count).
		Error

	if err != nil {
		return false, apperrors.NewInternalServerErr("Failed to check if user exists", err)
	}

	return count > 0, nil
}

func (r *userRepositoryImpl) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.NewNotFoundErr(fmt.Sprintf("User with username %s doesnt exist", username), err)
	}

	if err != nil {
		return nil, apperrors.NewInternalServerErr(fmt.Sprintf("Failed to get user with username %s", username), err)
	}

	return &user, nil
}

func (r *userRepositoryImpl) Create(user *models.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return apperrors.NewInternalServerErr(fmt.Sprintf("Failed to create user with username %s", user.Username), err)
	}
	return nil
}

func (r *userRepositoryImpl) Update(user *models.User) error {
	res := r.db.Model(user).Updates(user)
	if err := res.Error; err != nil {
		return apperrors.NewInternalServerErr(fmt.Sprintf("Failed to update user with username %s", user.Username), err)
	}

	if res.RowsAffected == 0 {
		return apperrors.NewNotFoundErr(fmt.Sprintf("no user with username %s found", user.Username), nil)
	}

	return nil
}

func (r *userRepositoryImpl) GetByRole(role models.Role) ([]*models.User, error) {
	var users []*models.User
	err := r.db.Where("role = ?", role).Find(&users).Error
	if err != nil {
		return nil, apperrors.NewInternalServerErr(fmt.Sprintf("Failed to get users with role %s", role), err)
	}
	return users, nil
}

func (r *userRepositoryImpl) DeleteByUsername(username string) error {
	err := r.db.Where("username = ?", username).Delete(&models.User{}).Error
	if err != nil {
		return apperrors.NewInternalServerErr(fmt.Sprintf("Failed to delete user with username %s", username), err)
	}
	return nil
}
