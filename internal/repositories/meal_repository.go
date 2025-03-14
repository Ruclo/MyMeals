package repositories

import (
	"fmt"
	"github.com/Ruclo/MyMeals/internal/models"
	"gorm.io/gorm"
)

type MealRepository interface {
	GetAll() ([]models.Meal, error)
	Create(meal *models.Meal) error
	Update(meal *models.Meal) error
}

func NewMealRepository(db *gorm.DB) *MealRepositoryImpl {
	return &MealRepositoryImpl{db: db}
}

type MealRepositoryImpl struct {
	db *gorm.DB
}

func (r *MealRepositoryImpl) GetAll() ([]models.Meal, error) {
	var meals []models.Meal

	if err := r.db.Find(&meals).Error; err != nil {
		return nil, err
	}

	return meals, nil
}

func (r *MealRepositoryImpl) Create(meal *models.Meal) error {
	return r.db.Create(meal).Error
}

func (r *MealRepositoryImpl) Update(meal *models.Meal) error {
	result := r.db.Model(meal).Select("*").Updates(meal)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("meal %s with id %d not found", meal.Name, meal.ID)
	}

	return nil
}
