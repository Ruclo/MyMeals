package repositories

import (
	"fmt"
	"github.com/Ruclo/MyMeals/internal/database"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

func TestMealRepoSuite(t *testing.T) {
	suite.Run(t, new(MealRepoTestSuite))
}

type MealRepoTestSuite struct {
	suite.Suite
	repo MealRepository
	db   *gorm.DB
}

func (s *MealRepoTestSuite) SetupSuite() {
	s.db = database.CreateConnection()
	s.repo = NewMealRepository(s.db)
}

func (s *MealRepoTestSuite) SetupTest() {
	database.WipeDB(s.db)
}

func CreateValidMeal() *models.Meal {
	return &models.Meal{
		ID:          0, // Should be auto-assigned
		Name:        "Test Pizza",
		Category:    models.MainCourses,
		Description: "Delicious test pizza with extra cheese",
		ImageURL:    "https://example.com/pizza.jpg",
		Price:       decimal.NewFromFloat(12.99),
	}
}

func (s *MealRepoTestSuite) TestCreate_ValidMeal() {
	s.SetupTest()
	meal := CreateValidMeal()

	err := s.repo.Create(meal)

	require.NoError(s.T(), err)
	assert.NotZero(s.T(), meal.ID, "ID should be auto-assigned")
	assert.Equal(s.T(), "Test Pizza", meal.Name)
	assert.Equal(s.T(), models.MainCourses, meal.Category)
	assert.Equal(s.T(), "Delicious test pizza with extra cheese", meal.Description)
	assert.Equal(s.T(), "https://example.com/pizza.jpg", meal.ImageURL)
	assert.Equal(s.T(), decimal.NewFromFloat(12.99), meal.Price)

	var found models.Meal
	result := s.db.First(&found, meal.ID)
	assert.NoError(s.T(), result.Error)
	assert.Equal(s.T(), meal.Name, found.Name)
	assert.Equal(s.T(), meal.Category, found.Category)
	assert.Equal(s.T(), meal.Description, found.Description)
	assert.Equal(s.T(), meal.ImageURL, found.ImageURL)
	assert.Equal(s.T(), meal.Price, found.Price)

}

func (s *MealRepoTestSuite) TestCreate_NonZeroID() {
	s.SetupTest()
	meal := CreateValidMeal()
	meal.ID = 42 // Non-zero ID

	err := s.repo.Create(meal)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Test Pizza", meal.Name)
	assert.Equal(s.T(), models.MainCourses, meal.Category)
	assert.Equal(s.T(), "Delicious test pizza with extra cheese", meal.Description)
	assert.Equal(s.T(), "https://example.com/pizza.jpg", meal.ImageURL)
	assert.Equal(s.T(), decimal.NewFromFloat(12.99), meal.Price)

}

// Create Tests
func (s *MealRepoTestSuite) TestCreate_EmptyName() {
	s.SetupTest()

	meal := CreateValidMeal()
	meal.Name = ""
	err := s.repo.Create(meal)
	assert.Error(s.T(), err, "Empty name should not be allowed")
}

func (s *MealRepoTestSuite) TestCreate_InvalidCategory() {
	s.SetupTest()
	meal := CreateValidMeal()
	meal.Category = "RandomCategory" // Invalid category

	// Act
	err := s.repo.Create(meal)

	// Assert
	assert.Error(s.T(), err, "Invalid category should not be allowed")
}

func (s *MealRepoTestSuite) TestCreate_EmptyDescription() {
	s.SetupTest()
	meal := CreateValidMeal()
	meal.Description = "" // Empty description

	// Act
	err := s.repo.Create(meal)
	fmt.Println(meal)
	// Assert
	assert.Error(s.T(), err, "Empty description should not be allowed")
}

func (s *MealRepoTestSuite) TestCreate_EmptyImageURL() {
	s.SetupTest()
	meal := CreateValidMeal()
	meal.ImageURL = "" // Empty URL

	// Act
	err := s.repo.Create(meal)

	// Assert
	assert.Error(s.T(), err, "Empty image URL should not be allowed")

}

func (s *MealRepoTestSuite) TestCreate_NegativePrice() {
	s.SetupTest()
	meal := CreateValidMeal()
	meal.Price = decimal.NewFromFloat(-5.99)

	// Act
	err := s.repo.Create(meal)

	// Assert
	assert.Error(s.T(), err, "Negative price should not be allowed")
}

//Update Tests

func (s *MealRepoTestSuite) TestUpdate_ValidMeal() {
	s.SetupTest()
	meal := CreateValidMeal()

	err := s.repo.Create(meal)
	require.NoError(s.T(), err)
	require.NotZero(s.T(), meal.ID)

	// Prepare update
	meal.Name = "Updated Pizza"
	meal.Category = models.Drinks
	meal.Description = "Updated description"
	meal.ImageURL = "https://example.com/updated-pizza.jpg"
	meal.Price = decimal.NewFromFloat(14.99)

	// Act
	err = s.repo.Update(meal)

	// Assert
	require.NoError(s.T(), err)

	var updated models.Meal
	result := s.db.First(&updated, meal.ID)
	assert.NoError(s.T(), result.Error)
	assert.Equal(s.T(), "Updated Pizza", updated.Name)
	assert.Equal(s.T(), models.Drinks, updated.Category)
	assert.Equal(s.T(), "Updated description", updated.Description)
	assert.Equal(s.T(), "https://example.com/updated-pizza.jpg", updated.ImageURL)
	assert.Equal(s.T(), decimal.NewFromFloat(14.99), updated.Price)

}

func (s *MealRepoTestSuite) TestUpdate_NonExistentMeal() {
	s.SetupTest()
	// Arrange - create a meal with non-existent ID
	meal := CreateValidMeal()
	meal.ID = 9999 // ID that doesn't exist

	// Act
	err := s.repo.Update(meal)

	// Assert
	assert.Error(s.T(), err, "Meal with non-existent ID should not be updated")
}

func (s *MealRepoTestSuite) TestUpdate_EmptyName() {
	s.SetupTest()
	meal := CreateValidMeal()
	err := s.repo.Create(meal)
	require.NoError(s.T(), err)

	// Prepare invalid update
	meal.Name = "" // Empty name

	// Act
	err = s.repo.Update(meal)

	assert.Error(s.T(), err, "Meal with empty name should not be updated")
}

func (s *MealRepoTestSuite) TestGetAll_NoEntries() {
	s.SetupTest()
	meals, err := s.repo.GetAll()
	require.NoError(s.T(), err)
	assert.Empty(s.T(), meals, "No meals should be returned")
}

func (s *MealRepoTestSuite) TestGetAll_WithEntries() {
	s.SetupTest()
	meals := []models.Meal{*CreateValidMeal(), *CreateValidMeal()}

	for _, meal := range meals {
		err := s.repo.Create(&meal)
		require.NoError(s.T(), err)
	}

	retrievedMeals, err := s.repo.GetAll()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), len(meals), len(retrievedMeals), "Number of meals should match")
}
