package repositories

import (
	"errors"
	"fmt"
	"github.com/Ruclo/MyMeals/internal/apperrors"
	"github.com/Ruclo/MyMeals/internal/models"
	"gorm.io/gorm"
)

// MealRepository provides an interface for CRUD operations on Meal entities and supports transactional operations.
// WithTransaction executes a function within a database transaction and rolls back if an error occurs.
// GetAll retrieves allMeal records from the database.
// GetAllWithDeleted retrieves all Meal records, including soft-deleted ones, from the database.
// GetByID retrieves a specific Meal by its ID from the database.
// Create adds a new Meal record to the database.
// Delete performs a soft delete on a Meal record in the database.
type MealRepository interface {
	WithTransaction(fn func(txRepo MealRepository) error) error
	GetAll() ([]*models.Meal, error)
	GetAllWithDeleted() ([]*models.Meal, error)
	GetByID(ID uint) (*models.Meal, error)
	Create(meal *models.Meal) error
	Delete(meal *models.Meal) error
}

func NewMealRepository(db *gorm.DB) MealRepository {
	return &mealRepositoryImpl{db: db}
}

type mealRepositoryImpl struct {
	db *gorm.DB
}

func (r *mealRepositoryImpl) WithTransaction(fn func(txRepo MealRepository) error) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return apperrors.NewInternalServerErr("Failed to start a transaction", tx.Error)
	}
	defer tx.Rollback()

	txRepo := &mealRepositoryImpl{db: tx}

	if err := fn(txRepo); err != nil {
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return apperrors.NewInternalServerErr("Failed to commit transaction", err)
	}
	return nil
}

func (r *mealRepositoryImpl) GetAll() ([]*models.Meal, error) {
	var meals []*models.Meal

	if err := r.db.Find(&meals).Error; err != nil {
		return nil, apperrors.NewInternalServerErr("Failed to get all meals", err)
	}

	return meals, nil
}

func (r *mealRepositoryImpl) GetAllWithDeleted() ([]*models.Meal, error) {
	var meals []*models.Meal

	if err := r.db.Unscoped().Find(&meals).Error; err != nil {
		return nil, apperrors.NewInternalServerErr("Failed to get all meals including deleted", err)
	}

	return meals, nil
}

func (r *mealRepositoryImpl) GetByID(ID uint) (*models.Meal, error) {
	var meal models.Meal
	err := r.db.Model(&models.Meal{}).Where("ID = ?", ID).First(&meal).Error

	if err == nil {
		return &meal, nil

	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.NewNotFoundErr(fmt.Sprintf("Meal with ID %d not found", ID), err)
	}

	return nil, apperrors.NewInternalServerErr("Failed to get meal", err)

}

func (r *mealRepositoryImpl) Create(meal *models.Meal) error {
	if err := r.db.Create(meal).Error; err != nil {
		return apperrors.NewInternalServerErr("Failed to create meal", err)
	}
	return nil
}

func (r *mealRepositoryImpl) Delete(meal *models.Meal) error {
	result := r.db.Delete(meal)
	if err := result.Error; err != nil {
		return apperrors.NewInternalServerErr("Failed to delete meal", err)
	}

	if result.RowsAffected == 0 {
		return apperrors.NewNotFoundErr(fmt.Sprintf("meal %s not found", meal.Name), nil)
	}

	return nil
}
