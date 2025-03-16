package repositories

import (
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
		return &user, nil
	}
	return nil, err
}

func (r *userRepositoryImpl) Create(user *models.StaffMember) error {
	return r.db.Create(user).Error
}

func (r *userRepositoryImpl) Update(user *models.StaffMember) error {
	return r.db.Model(user).Updates(user).Error
}
