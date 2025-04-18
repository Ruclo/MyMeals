package repositories

import (
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

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) Create(user *models.StaffMember) error {
	return r.db.Create(user).Error
}

func (r *userRepositoryImpl) Update(user *models.StaffMember) error {
	res := r.db.Model(user).Updates(user)
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		err := fmt.Errorf("no user with username %s found", user.Username)
		return errors.NewNotFoundErr(err.Error(), err)
	}

	return nil
}
