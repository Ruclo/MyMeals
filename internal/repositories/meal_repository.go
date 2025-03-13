package repositories

import (
	"fmt"
	"github.com/Ruclo/MyMeals/internal/models"
	"gorm.io/gorm"
)

type MealRepository interface {
	GetAll() ([]models.Meal, error)
	Create(meal models.Meal) error
	Update(meal models.Meal) error
}

type mealRepository struct {
	db *gorm.DB
}

func (r *mealRepository) GetAll() ([]models.Meal, error) {
	var meals []models.Meal

	if err := r.db.Find(&meals).Error; err != nil {
		return nil, err
	}

	return meals, nil
}

func (r *mealRepository) Create(meal *models.Meal) error {
	return r.db.Create(meal).Error
}

func (r *mealRepository) Update(meal *models.Meal) error {
	result := r.db.Model(meal).Updates(meal)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected != 0 {
		return fmt.Errorf("meal %s with id %d not found", meal.Name, meal.ID)
	}

	return nil
}
