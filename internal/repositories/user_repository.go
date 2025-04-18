package repositories

import (
	stdErrors "errors"
	"fmt"
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetByUsername(username string) (*models.StaffMember, error)
	Create(user *models.StaffMember) error
	Update(user *models.StaffMember) error
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

type userRepositoryImpl struct {
	db *gorm.DB
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
	tx := r.db.Begin()
	defer tx.Rollback()

	var count int64
	err := tx.Model(&user).Where("username = ?", user.Username).Count(&count).Error
	if err != nil {
		return errors.NewInternalServerErr(fmt.Sprintf("Failed to check user count with username %s", user.Username), err)
	}

	if count > 0 {
		return errors.NewAlreadyExistsErr(fmt.Sprintf("User with username %s already exists", user.Username), nil)
	}

	if err = tx.Create(user).Error; err != nil {
		return errors.NewInternalServerErr(fmt.Sprintf("Failed to create user with username %s", user.Username), err)
	}

	if err = tx.Commit().Error; err != nil {
		return errors.NewInternalServerErr(fmt.Sprintf("Failed to commit user with username %s", user.Username), err)
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
