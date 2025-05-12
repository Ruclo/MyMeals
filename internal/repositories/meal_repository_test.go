package repositories_test

import (
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	testinghelpers "github.com/Ruclo/MyMeals/internal/testing"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestMealRepository_GetAll(t *testing.T) {
	// Setup
	db := testinghelpers.NewTestDB(t)
	defer testinghelpers.CleanupTestDB(t, db)

	repo := repositories.NewMealRepository(db)

	meals, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, meals, 0)

	// Create test data
	testMeals := []models.Meal{
		{Name: "Test Meal 1", Price: decimal.NewFromFloat(10.99), Category: "Main Courses", Description: "Test Description 1", ImageURL: "http://example.com/1.jpg"},
		{Name: "Test Meal 2", Price: decimal.NewFromFloat(7.99), Category: "Desserts", Description: "Test Description 2", ImageURL: "http://example.com/2.jpg"},
	}

	for i := range testMeals {
		assert.NoError(t, db.Create(&testMeals[i]).Error)
	}

	// Execute
	meals, err = repo.GetAll()

	// Verify
	assert.NoError(t, err)
	assert.Len(t, meals, 2)

	for _, meal := range meals {
		found := false
		for _, inputMeal := range testMeals {
			if meal.Name == inputMeal.Name {
				found = true
				assert.NotZero(t, meal.ID)
			}
		}
		assert.True(t, found)
	}
}

func TestMealRepository_GetByID(t *testing.T) {
	// Setup
	db := testinghelpers.NewTestDB(t)
	defer testinghelpers.CleanupTestDB(t, db)

	repo := repositories.NewMealRepository(db)

	testMeal := models.Meal{
		Name:        "Test Meal",
		Price:       decimal.NewFromFloat(9.99),
		Category:    "Main Courses",
		Description: "Test Description",
		ImageURL:    "http://example.com/image.jpg",
	}

	assert.NoError(t, db.Create(&testMeal).Error)

	// Execute
	meal, err := repo.GetByID(testMeal.ID)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, meal)
	assert.Equal(t, testMeal.ID, meal.ID)
	assert.Equal(t, testMeal.Name, meal.Name)

	// Non existant meal
	meal, err = repo.GetByID(999)
	assert.Error(t, err)

	var appErr *errors.AppError
	if assert.ErrorAs(t, err, &appErr) {
		assert.Equal(t, http.StatusNotFound, appErr.StatusCode) // TODO: Better error checking
	}

	assert.Nil(t, meal)
}

func TestMealRepository_Create(t *testing.T) {

	t.Run("valid meal", func(t *testing.T) {
		db := testinghelpers.NewTestDB(t)
		defer testinghelpers.CleanupTestDB(t, db)

		repo := repositories.NewMealRepository(db)
		meal := getTestMeal()
		err := repo.Create(meal)
		assert.NoError(t, err)
		assert.NotZero(t, meal.ID)

		var meals []*models.Meal
		assert.NoError(t, db.Model(&models.Meal{}).Find(&meals).Error)
		assert.Len(t, meals, 1)

		assertMealEquals(t, meal, meals[0])
	})

	t.Run("invalid meal", func(t *testing.T) {
		db := testinghelpers.NewTestDB(t)
		defer testinghelpers.CleanupTestDB(t, db)

		repo := repositories.NewMealRepository(db)
		meal := getTestMeal()
		meal.Price = decimal.NewFromInt(0)

		err := repo.Create(meal)
		assert.Error(t, err)
	})
}

/*func TestMealRepository_Update(t *testing.T) {
	t.Run("existing meal", func(t *testing.T) {
		db := testinghelpers.NewTestDB(t)
		defer testinghelpers.CleanupTestDB(t, db)
		repo := repositories.NewMealRepository(db)

		meal := getTestMeal()
		err := db.Create(&meal).Error
		require.NoError(t, err)
		assert.NotZero(t, meal.ID)

		meal.Name = "Updated name"
		err = repo.Update(meal)
		assert.NoError(t, err)

		var fetchedMeal models.Meal
		err = db.Model(&models.Meal{}).Where("ID = ?", meal.ID).First(&fetchedMeal).Error
		assert.NoError(t, err)
		assertMealEquals(t, meal, &fetchedMeal)

		fetchedMeal.Category = "Invalid Category"
		assert.Error(t, repo.Update(&fetchedMeal))

	})
	t.Run("non-existent meal", func(t *testing.T) {
		db := testinghelpers.NewTestDB(t)
		defer testinghelpers.CleanupTestDB(t, db)
		repo := repositories.NewMealRepository(db)

		meal := getTestMeal()
		meal.ID = 1
		err := repo.Update(meal)
		assert.Error(t, err)

		var appErr *errors.AppError
		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusNotFound, appErr.StatusCode)
		}
	})
}*/

func TestMealRepository_Delete(t *testing.T) {
	t.Run("existing meal", func(t *testing.T) {
		db := testinghelpers.NewTestDB(t)
		defer testinghelpers.CleanupTestDB(t, db)
		repo := repositories.NewMealRepository(db)

		meal := getTestMeal()
		err := db.Create(meal).Error
		assert.NoError(t, err)

		err = repo.Delete(meal)
		assert.NoError(t, err)

		var meals []*models.Meal

		assert.NoError(t, db.Model(&models.Meal{}).Find(&meals).Error)
		assert.Empty(t, meals)

		var foundMeal models.Meal
		assert.NoError(t, db.Unscoped().First(&foundMeal, meal.ID).Error)
		assertMealEquals(t, meal, &foundMeal)

	})
	t.Run("non-existent meal", func(t *testing.T) {
		db := testinghelpers.NewTestDB(t)
		defer testinghelpers.CleanupTestDB(t, db)
		repo := repositories.NewMealRepository(db)

		err := repo.Delete(&models.Meal{ID: 1})
		assert.Error(t, err)

		var appErr *errors.AppError

		if assert.ErrorAs(t, err, &appErr) {
			assert.Equal(t, http.StatusNotFound, appErr.StatusCode)
		}
	})
}

func getTestMeal() *models.Meal {
	return &models.Meal{
		Name:        "Test Meal",
		Description: "Test Description",
		Category:    models.MainCourses,
		ImageURL:    "http://image.com",
		Price:       decimal.NewFromFloat(9.99),
	}
}

func TestMealRepository_IntegrationFlow(t *testing.T) {
	db := testinghelpers.NewTestDB(t)
	repo := repositories.NewMealRepository(db)

	// 1. Try operations on non-existent data
	t.Run("operations on non-existent data", func(t *testing.T) {
		// Try to get by ID
		_, err := repo.GetByID(999)
		assert.Error(t, err)

		// Try to delete
		deleteErr := repo.Delete(&models.Meal{ID: 999})
		assert.Error(t, deleteErr)
	})

	// 2. Create and verify data
	var mealID uint
	t.Run("create and verify", func(t *testing.T) {
		// Check empty state
		meals, err := repo.GetAll()
		assert.NoError(t, err)
		initialCount := len(meals)

		// Create a meal
		meal := getTestMeal()
		expected := getTestMeal()

		err = repo.Create(meal)
		assert.NoError(t, err)
		assert.NotZero(t, meal.ID)
		mealID = meal.ID
		expected.ID = meal.ID

		fetchedMeal, err := repo.GetByID(mealID)
		assert.NoError(t, err)
		assertMealEquals(t, expected, fetchedMeal)

		meals, err = repo.GetAll()
		assert.NoError(t, err)
		assert.Equal(t, initialCount+1, len(meals))
	})

	// 4. Delete and verify deletion
	t.Run("delete and verify", func(t *testing.T) {
		// Get count before deletion
		meals, err := repo.GetAll()
		assert.NoError(t, err)
		countBeforeDelete := len(meals)

		// Delete the meal
		err = repo.Delete(&models.Meal{ID: mealID})
		assert.NoError(t, err)

		// Verify with GetByID
		_, err = repo.GetByID(mealID)
		assert.Error(t, err)

		// Verify with GetAll
		meals, err = repo.GetAll()
		assert.NoError(t, err)
		assert.Equal(t, countBeforeDelete-1, len(meals))

		// Try updating deleted meal
	})
}

func assertMealEquals(t *testing.T, expected, actual *models.Meal) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Description, actual.Description)
	assert.Equal(t, expected.Category, actual.Category)
	assert.Equal(t, expected.Price, actual.Price)
	assert.Equal(t, expected.ImageURL, actual.ImageURL)
}
