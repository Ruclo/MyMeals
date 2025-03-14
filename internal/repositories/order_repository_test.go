package repositories

import (
	"github.com/Ruclo/MyMeals/internal/database"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

func TestOrderRepoSuite(t *testing.T) {
	suite.Run(t, new(OrderRepoTestSuite))
}

type OrderRepoTestSuite struct {
	suite.Suite
	repo OrderRepository
	db   *gorm.DB
}

func (s *OrderRepoTestSuite) SetupSuite() {
	s.db = database.CreateConnection()
	s.repo = NewOrderRepository(s.db)
}

func (s *OrderRepoTestSuite) SetupTest() {
	database.WipeDB(s.db)
}

func (s *OrderRepoTestSuite) TestCreate_ValidOrder() {
	s.SetupTest()

	meal1 := models.Meal{
		Name:        "Test Meal 1",
		Description: "Description 1",
		Price:       decimal.NewFromFloat(9.99),
		Category:    "Main",
	}
	meal2 := models.Meal{
		Name:        "Test Meal 2",
		Description: "Description 2",
		Price:       decimal.NewFromFloat(5.99),
		Category:    "Side",
	}

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
