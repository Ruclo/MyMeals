package repositories

import (
	stdErrors "errors"
	"fmt"
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"gorm.io/gorm"
)

type MealRepository interface {
	GetAll() ([]models.Meal, error)
	GetByID(ID uint) (*models.Meal, error)
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
		return nil, errors.NewInternalServerErr("Failed to get all meals", err)
	}

	return meals, nil
}

func (r *mealRepositoryImpl) GetByID(ID uint) (*models.Meal, error) {
	var meal models.Meal
	err := r.db.Model(&models.Meal{}).Where("ID = ?").First(&meal).Error

	if err == nil {
		return &meal, nil

	}

	if stdErrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NewNotFoundErr(fmt.Sprintf("Meal with ID %d not found", ID), err)
	}

	return nil, errors.NewInternalServerErr("Failed to get meal", err)

}

func (r *mealRepositoryImpl) Create(meal *models.Meal) error {
	if err := r.db.Create(meal).Error; err != nil {
		return errors.NewInternalServerErr("Failed to create meal", err)
	}
	return nil
}

func (r *mealRepositoryImpl) Update(meal *models.Meal) error {
	result := r.db.Model(meal).Select("*").Updates(meal)
	if result.Error != nil {
		return errors.NewInternalServerErr("Failed to update meal", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewNotFoundErr(fmt.Sprintf("meal %s with id %d not found", meal.Name, meal.ID), nil)
	}

	return nil
}

func (r *mealRepositoryImpl) Delete(meal *models.Meal) error {
	result := r.db.Delete(meal)
	if err := result.Error; err != nil {
		return errors.NewInternalServerErr("Failed to delete meal", err)
	}

	if result.RowsAffected == 0 {
		return errors.NewNotFoundErr(fmt.Sprintf("meal %s not found", meal.Name), nil)
	}

	return nil
}
