package repositories

import (
	"fmt"
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"gorm.io/gorm"
)

type MealRepository interface {
	GetAll() ([]models.Meal, error)
	Create(meal *models.Meal) error
	Update(meal *models.Meal) error
	Delete(meal *models.Meal) error
}

func NewMealRepository(db *gorm.DB) MealRepository {
	return &mealRepositoryImpl{db: db}
}

type mealRepositoryImpl struct {
	db *gorm.DB
}

func (r *mealRepositoryImpl) GetAll() ([]models.Meal, error) {
	var meals []models.Meal

	if err := r.db.Find(&meals).Error; err != nil {
		return nil, err
	}

	return meals, nil
}

func (r *mealRepositoryImpl) Create(meal *models.Meal) error {
	return r.db.Create(meal).Error
}

func (r *mealRepositoryImpl) Update(meal *models.Meal) error {
	result := r.db.Model(meal).Select("*").Updates(meal)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		err := fmt.Errorf("meal %s with id %d not found", meal.Name, meal.ID)
		return errors.NewNotFoundErr(err.Error(), err)
	}

	return nil
}

func (r *mealRepositoryImpl) Delete(meal *models.Meal) error {
	return r.db.Delete(meal).Error
}
