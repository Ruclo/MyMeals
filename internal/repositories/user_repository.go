package repositories

import (
	stdErrors "errors"
	"fmt"
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	WithTransaction(fn func(txRepo UserRepository) error) error
	GetByUsername(username string) (*models.StaffMember, error)
	GetByRole(role models.Role) ([]*models.StaffMember, error)
	Create(user *models.StaffMember) error
	Update(user *models.StaffMember) error
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
		return errors.NewInternalServerErr("Failed to start a transaction", tx.Error)
	}
	defer tx.Rollback()

	txRepo := &userRepositoryImpl{db: tx}

	if err := fn(txRepo); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return errors.NewInternalServerErr("Failed to commit transaction", err)
	}

	return nil
}

func (r *userRepositoryImpl) Exists(username string) (bool, error) {
	var count int64
	err := r.db.Model(&models.StaffMember{}).
		Where("username = ?", username).
		Count(&count).
		Error

	if err != nil {
		return false, errors.NewInternalServerErr("Failed to check if user exists", err)
	}

	return count > 0, nil
}

func (r *userRepositoryImpl) GetByUsername(username string) (*models.StaffMember, error) {
	var user models.StaffMember
	err := r.db.Where("username = ?", username).First(&user).Error

	if stdErrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NewNotFoundErr(fmt.Sprintf("User with username %s doesnt exist", username), err)
	}

	if err != nil {
		return nil, errors.NewInternalServerErr(fmt.Sprintf("Failed to get user with username %s", username), err)
	}

	return &user, nil
}

func (r *userRepositoryImpl) Create(user *models.StaffMember) error {
	if err := r.db.Create(user).Error; err != nil {
		return errors.NewInternalServerErr(fmt.Sprintf("Failed to create user with username %s", user.Username), err)
	}
	return nil
}

func (r *userRepositoryImpl) Update(user *models.StaffMember) error {
	res := r.db.Model(user).Updates(user)
	if err := res.Error; err != nil {
		return errors.NewInternalServerErr(fmt.Sprintf("Failed to update user with username %s", user.Username), err)
	}

	if res.RowsAffected == 0 {
		return errors.NewNotFoundErr(fmt.Sprintf("no user with username %s found", user.Username), nil)
	}

	return nil
}

func (r *userRepositoryImpl) GetByRole(role models.Role) ([]*models.StaffMember, error) {
	var users []*models.StaffMember
	err := r.db.Where("role = ?", role).Find(&users).Error
	if err != nil {
		return nil, errors.NewInternalServerErr(fmt.Sprintf("Failed to get users with role %s", role), err)
	}
	return users, nil
}

func (r *userRepositoryImpl) DeleteByUsername(username string) error {
	err := r.db.Where("username = ?", username).Delete(&models.StaffMember{}).Error
	if err != nil {
		return errors.NewInternalServerErr(fmt.Sprintf("Failed to delete user with username %s", username), err)
	}
	return nil
}
