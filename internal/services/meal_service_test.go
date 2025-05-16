package services_test

import (
	"context"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"testing"

	"github.com/Ruclo/MyMeals/internal/apperrors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/services"
	"github.com/Ruclo/MyMeals/internal/storage"
	"github.com/Ruclo/MyMeals/internal/testing/mocks"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"mime/multipart"
)

// MealServiceTestSuite defines the test suite for MealService
type MealServiceTestSuite struct {
	suite.Suite
	mealService      services.MealService
	mockRepo         *MockMealRepository
	mockImageStorage *mocks.MockImageStorage
	ginContext       *gin.Context
}

func (s *MealServiceTestSuite) SetupTest() {
	// Create fresh mocks for each test
	s.mockRepo = new(MockMealRepository)
	s.mockImageStorage = new(mocks.MockImageStorage)

	s.mealService = services.NewMealService(s.mockRepo, s.mockImageStorage)

	// Create a Gin context for testing
	s.ginContext = &gin.Context{}
}

// TearDownTest runs after each test
func (s *MealServiceTestSuite) TearDownTest() {
	// Verify all mock expectations were met
	s.mockRepo.AssertExpectations(s.T())
	s.mockImageStorage.AssertExpectations(s.T())
}

// TestCreate tests the Create method
func (s *MealServiceTestSuite) TestCreate() {
	price1599, _ := decimal.NewFromString("15.99")

	testCases := []struct {
		name           string
		meal           *models.Meal
		setupMock      func()
		expectedError  bool
		errorPredicate func(error) bool
		checkMeal      func(*models.Meal)
	}{
		{
			name: "Success",
			meal: &models.Meal{
				Name:        "Test Meal",
				Category:    models.MainCourses,
				Description: "Test Description",
				Price:       price1599,
			},
			setupMock: func() {
				// Mock successful upload
				uploadResult := &storage.ImageResult{
					URL:      "https://cloudinary.com/test-image.jpg",
					PublicID: "test-image",
				}
				s.mockImageStorage.On("UploadCropped",
					s.ginContext,
					mock.AnythingOfType("*multipart.FileHeader"),
					1000, 1000,
				).Return(uploadResult, nil)

				// Mock successful meal creation
				s.mockRepo.On("Create", mock.MatchedBy(func(meal *models.Meal) bool {
					return meal.Name == "Test Meal" &&
						meal.ImageURL == "https://cloudinary.com/test-image.jpg"
				})).Return(nil)
			},
			expectedError: false,
			checkMeal: func(meal *models.Meal) {
				s.Equal("Test Meal", meal.Name)
				s.Equal(models.MainCourses, meal.Category)
				s.Equal("Test Description", meal.Description)
				s.True(price1599.Equal(meal.Price))
				s.Equal("https://cloudinary.com/test-image.jpg", meal.ImageURL)
			},
		},
		{
			name: "UploadError",
			meal: &models.Meal{
				Name:        "Test Meal",
				Category:    models.MainCourses,
				Description: "Test Description",
				Price:       price1599,
			},
			setupMock: func() {
				// Mock upload error
				uploadErr := apperrors.NewInternalServerErr("Upload failed", nil)
				s.mockImageStorage.On("UploadCropped",
					s.ginContext,
					mock.AnythingOfType("*multipart.FileHeader"),
					1000, 1000,
				).Return(nil, uploadErr)

				// Repository should not be called
			},
			expectedError: true,
		},
		{
			name: "DatabaseError",
			meal: &models.Meal{
				Name:        "Test Meal",
				Category:    models.MainCourses,
				Description: "Test Description",
				Price:       price1599,
			},
			setupMock: func() {
				// Mock successful upload
				uploadResult := &storage.ImageResult{
					URL:      "https://cloudinary.com/test-image.jpg",
					PublicID: "test-image",
				}
				s.mockImageStorage.On("UploadCropped",
					s.ginContext,
					mock.AnythingOfType("*multipart.FileHeader"),
					1000, 1000,
				).Return(uploadResult, nil)

				// Mock database error
				dbErr := apperrors.NewInternalServerErr("Database error", nil)
				s.mockRepo.On("Create", mock.AnythingOfType("*models.Meal")).Return(dbErr)

				// Mock delete call due to rollback
				s.mockImageStorage.On("Delete",
					s.ginContext,
					"test-image",
				).Return(nil)
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup fresh mocks
			s.SetupTest()

			// Setup mock expectations
			tc.setupMock()

			// Create a dummy file header for testing
			dummyFileHeader := &multipart.FileHeader{
				Filename: "test.jpg",
				Size:     1024,
			}

			// Create a copy to avoid modifications between tests
			mealCopy := &models.Meal{
				Name:        tc.meal.Name,
				Category:    tc.meal.Category,
				Description: tc.meal.Description,
				Price:       tc.meal.Price,
			}

			// Act
			err := s.mealService.Create(s.ginContext, mealCopy, dummyFileHeader)

			// Assert
			if tc.expectedError {
				s.Error(err)
				if tc.errorPredicate != nil {
					s.True(tc.errorPredicate(err))
				}
			} else {
				s.NoError(err)
				if tc.checkMeal != nil {
					tc.checkMeal(mealCopy) // Test on the modified input model
				}
			}
		})
	}
}

// TestGetAll tests the GetAll method
func (s *MealServiceTestSuite) TestGetAll() {
	price999, _ := decimal.NewFromString("9.99")
	price1299, _ := decimal.NewFromString("12.99")

	testCases := []struct {
		name           string
		setupMock      func()
		expectedMeals  []models.Meal
		expectedError  bool
		errorPredicate func(error) bool
	}{
		{
			name: "Success",
			setupMock: func() {
				meals := []models.Meal{
					{
						ID:          1,
						Name:        "Meal 1",
						Category:    "Category 1",
						Description: "Description 1",
						Price:       price999,
						ImageURL:    "image1.jpg",
					},
					{
						ID:          2,
						Name:        "Meal 2",
						Category:    "Category 2",
						Description: "Description 2",
						Price:       price1299,
						ImageURL:    "image2.jpg",
					},
				}
				s.mockRepo.On("GetAll").Return(meals, nil)
			},
			expectedMeals: []models.Meal{
				{
					ID:          1,
					Name:        "Meal 1",
					Category:    "Category 1",
					Description: "Description 1",
					Price:       price999,
					ImageURL:    "image1.jpg",
				},
				{
					ID:          2,
					Name:        "Meal 2",
					Category:    "Category 2",
					Description: "Description 2",
					Price:       price1299,
					ImageURL:    "image2.jpg",
				},
			},
			expectedError: false,
		},
		{
			name: "EmptyList",
			setupMock: func() {
				s.mockRepo.On("GetAll").Return([]models.Meal{}, nil)
			},
			expectedMeals: []models.Meal{},
			expectedError: false,
		},
		{
			name: "DatabaseError",
			setupMock: func() {
				dbErr := apperrors.NewInternalServerErr("Database error", nil)
				s.mockRepo.On("GetAll").Return(nil, dbErr)
			},
			expectedMeals: nil,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup fresh mocks
			s.SetupTest()

			// Setup mock expectations
			tc.setupMock()

			// Act
			meals, err := s.mealService.GetAll()

			// Assert
			if tc.expectedError {
				s.Error(err)
				if tc.errorPredicate != nil {
					s.True(tc.errorPredicate(err))
				}
				s.Nil(meals)
			} else {
				s.NoError(err)
				s.Equal(len(tc.expectedMeals), len(meals))

				if len(meals) > 0 {
					s.Equal(tc.expectedMeals[0].ID, meals[0].ID)
					s.Equal(tc.expectedMeals[0].Name, meals[0].Name)
					s.Equal(tc.expectedMeals[0].Category, meals[0].Category)
					s.True(tc.expectedMeals[0].Price.Equal(meals[0].Price))
				}
			}
		})
	}
}

func (s *MealServiceTestSuite) TestUpdate() {
	price1999, _ := decimal.NewFromString("19.99")
	ctx := context.Background()

	testCases := []struct {
		name           string
		meal           *models.Meal
		photo          *multipart.FileHeader
		setupMock      func()
		expectedError  bool
		errorPredicate func(error) bool
		checkMeal      func(*models.Meal)
	}{
		{
			name: "Success with photo update",
			meal: &models.Meal{
				ID:          1,
				Name:        "Updated Meal",
				Category:    "Updated Category",
				Description: "Updated Description",
				Price:       price1999,
			},
			photo: &multipart.FileHeader{Filename: "new-photo.jpg"},
			setupMock: func() {
				existingMeal := &models.Meal{
					ID:          1,
					Name:        "Original Meal",
					Category:    "Original Category",
					Description: "Original Description",
					Price:       decimal.NewFromFloat(9.99),
					ImageURL:    "old-image.jpg",
				}
				s.mockRepo.On("GetByID", uint(1)).Return(existingMeal, nil)

				s.mockImageStorage.On("UploadCropped", mock.Anything, mock.AnythingOfType("*multipart.FileHeader"), 1000, 1000).
					Return(&storage.ImageResult{URL: "new-image.jpg", PublicID: "new-image"}, nil)

				s.mockImageStorage.On("Delete", mock.Anything, "old-image.jpg").Return(nil)

				// Mock transaction
				s.mockRepo.On("WithTransaction", mock.AnythingOfType("func(repositories.MealRepository) error")).
					Run(func(args mock.Arguments) {
						// Execute the transaction function with our mock repo
						txFunc := args.Get(0).(func(repositories.MealRepository) error)
						txFunc(s.mockRepo)
					}).Return(nil)

				// Mock creating new meal
				s.mockRepo.On("Create", mock.MatchedBy(func(meal *models.Meal) bool {
					return meal.Name == "Updated Meal" &&
						meal.Category == "Updated Category" &&
						meal.Description == "Updated Description" &&
						meal.ImageURL == "new-image.jpg" &&
						meal.Price.Equal(price1999)
				})).
					Run(func(args mock.Arguments) {
						// Set an ID on the new meal
						newMeal := args.Get(0).(*models.Meal)
						newMeal.ID = 2 // Simulate auto-increment ID
					}).
					Return(nil)

				// Mock deleting the old meal
				s.mockRepo.On("Delete", existingMeal).Return(nil)
			},
			expectedError: false,
			checkMeal: func(meal *models.Meal) {
				// Verify the input meal was updated with the new meal's properties
				s.Equal("Updated Meal", meal.Name)
				s.Equal(models.MealCategory("Updated Category"), meal.Category)
				s.Equal("Updated Description", meal.Description)
				s.Equal("new-image.jpg", meal.ImageURL)
				s.True(price1999.Equal(meal.Price))
				s.Equal(uint(2), meal.ID) // Should have new ID
			},
		},
		{
			name: "Success without photo update",
			meal: &models.Meal{
				ID:          2,
				Name:        "Updated Meal No Photo",
				Category:    "Updated Category",
				Description: "Updated Description",
				Price:       price1999,
				ImageURL:    "existing-image.jpg",
			},
			photo: nil,
			setupMock: func() {
				// Mock getting the existing meal
				existingMeal := &models.Meal{
					ID:          2,
					Name:        "Original Meal",
					Category:    "Original Category",
					Description: "Original Description",
					Price:       decimal.NewFromFloat(9.99),
					ImageURL:    "existing-image.jpg",
				}
				s.mockRepo.On("GetByID", uint(2)).Return(existingMeal, nil)

				// Mock transaction
				s.mockRepo.On("WithTransaction", mock.AnythingOfType("func(repositories.MealRepository) error")).
					Run(func(args mock.Arguments) {
						// Execute the transaction function with our mock repo
						txFunc := args.Get(0).(func(repositories.MealRepository) error)
						txFunc(s.mockRepo)
					}).Return(nil)

				// Mock creating new meal
				s.mockRepo.On("Create", mock.MatchedBy(func(meal *models.Meal) bool {
					return meal.Name == "Updated Meal No Photo" &&
						meal.Category == "Updated Category" &&
						meal.Description == "Updated Description" &&
						meal.Price.Equal(price1999) &&
						meal.ImageURL == "existing-image.jpg"
				})).
					Run(func(args mock.Arguments) {
						newMeal := args.Get(0).(*models.Meal)
						newMeal.ID = 3
					}).
					Return(nil)

				s.mockRepo.On("Delete", existingMeal).Return(nil)
			},
			expectedError: false,
			checkMeal: func(meal *models.Meal) {
				s.Equal("Updated Meal No Photo", meal.Name)
				s.Equal(models.MealCategory("Updated Category"), meal.Category)
				s.Equal("Updated Description", meal.Description)
				s.Equal(uint(3), meal.ID) // Should have new ID
			},
		},
		{
			name: "Meal not found",
			meal: &models.Meal{
				ID:          99,
				Name:        "Nonexistent Meal",
				Category:    "Category",
				Description: "Description",
				Price:       price1999,
			},
			photo: nil,
			setupMock: func() {
				// Mock meal not found
				notFoundErr := apperrors.NewNotFoundErr("Meal not found", nil)
				s.mockRepo.On("GetByID", uint(99)).Return(nil, notFoundErr)
			},
			expectedError: true,
			errorPredicate: func(err error) bool {
				return apperrors.IsNotFoundErr(err)
			},
		},
		{
			name: "Transaction error",
			meal: &models.Meal{
				ID:          3,
				Name:        "Transaction Error Meal",
				Category:    "Category",
				Description: "Description",
				Price:       price1999,
				ImageURL:    "image.jpg",
			},
			photo: nil,
			setupMock: func() {
				// Mock getting the existing meal
				existingMeal := &models.Meal{
					ID:          3,
					Name:        "Original Meal",
					Description: "Original Description",
					Price:       decimal.NewFromFloat(9.99),
					ImageURL:    "image.jpg",
				}
				s.mockRepo.On("GetByID", uint(3)).Return(existingMeal, nil)

				// Mock transaction error
				s.mockRepo.On("WithTransaction", mock.AnythingOfType("func(repositories.MealRepository) error")).
					Return(apperrors.NewInternalServerErr("Transaction failed", nil))
			},
			expectedError: true,
			errorPredicate: func(err error) bool {
				return apperrors.IsInternalServerErr(err)
			},
		},
		{
			name: "Failed to create new meal",
			meal: &models.Meal{
				ID:          4,
				Name:        "Create Error Meal",
				Category:    "Category",
				Description: "Description",
				Price:       price1999,
				ImageURL:    "image.jpg",
			},
			photo: nil,
			setupMock: func() {
				// Mock getting the existing meal
				existingMeal := &models.Meal{
					ID:          4,
					Name:        "Original Meal",
					Description: "Original Description",
					Price:       decimal.NewFromFloat(9.99),
					ImageURL:    "image.jpg",
				}
				s.mockRepo.On("GetByID", uint(4)).Return(existingMeal, nil)

				// Mock transaction
				s.mockRepo.On("WithTransaction", mock.AnythingOfType("func(repositories.MealRepository) error")).
					Run(func(args mock.Arguments) {
						// Execute the transaction function with our mock repo
						txFunc := args.Get(0).(func(repositories.MealRepository) error)

						// Inside transaction, Create returns an error
						s.mockRepo.On("Create", mock.AnythingOfType("*models.Meal")).
							Return(apperrors.NewInternalServerErr("Failed to create meal", nil))

						// Call the function which should now return an error
						txFunc(s.mockRepo)
					}).Return(apperrors.NewInternalServerErr("Failed to create meal", nil))
			},
			expectedError: true,
			errorPredicate: func(err error) bool {
				return apperrors.IsInternalServerErr(err) && err.Error() == "Failed to create meal"
			},
		},
		{
			name: "Failed to upload new photo",
			meal: &models.Meal{
				ID:          5,
				Name:        "Photo Upload Error Meal",
				Category:    "Category",
				Description: "Description",
				Price:       price1999,
				ImageURL:    "old-image.jpg",
			},
			photo: &multipart.FileHeader{Filename: "new-photo.jpg"},
			setupMock: func() {
				// Mock getting the existing meal
				existingMeal := &models.Meal{
					ID:          5,
					Name:        "Original Meal",
					Description: "Original Description",
					Price:       decimal.NewFromFloat(9.99),
					ImageURL:    "old-image.jpg",
				}
				s.mockRepo.On("GetByID", uint(5)).Return(existingMeal, nil)

				// Mock transaction
				s.mockRepo.On("WithTransaction", mock.AnythingOfType("func(repositories.MealRepository) error")).
					Run(func(args mock.Arguments) {
						// Execute the transaction function with our mock repo
						txFunc := args.Get(0).(func(repositories.MealRepository) error)

						// Inside transaction, upload fails
						uploadErr := apperrors.NewInternalServerErr("Failed to upload photo", nil)
						s.mockImageStorage.On("UploadCropped", mock.Anything, mock.AnythingOfType("*multipart.FileHeader"), 1000, 1000).
							Return(nil, uploadErr)

						// Call the function which should return the upload error
						txFunc(s.mockRepo)
					}).Return(apperrors.NewInternalServerErr("Failed to upload photo", nil))
			},
			expectedError: true,
			errorPredicate: func(err error) bool {
				return apperrors.IsInternalServerErr(err) && err.Error() == "Failed to upload photo"
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup fresh mocks
			s.SetupTest()

			// Setup mock expectations
			tc.setupMock()

			// Create a copy to avoid modifications between tests
			mealCopy := &models.Meal{
				ID:          tc.meal.ID,
				Name:        tc.meal.Name,
				Category:    tc.meal.Category,
				Description: tc.meal.Description,
				Price:       tc.meal.Price,
				ImageURL:    tc.meal.ImageURL,
			}

			// Act
			err := s.mealService.Replace(ctx, mealCopy, tc.photo)

			// Assert
			if tc.expectedError {
				s.Error(err)
				if tc.errorPredicate != nil {
					s.True(tc.errorPredicate(err))
				}
			} else {
				s.NoError(err)
				if tc.checkMeal != nil {
					tc.checkMeal(mealCopy)
				}
			}
		})
	}
}

// TestDelete tests the Delete method
func (s *MealServiceTestSuite) TestDelete() {
	testCases := []struct {
		name           string
		mealID         uint
		setupMock      func()
		expectedError  bool
		errorPredicate func(error) bool
	}{
		{
			name:   "Success",
			mealID: 1,
			setupMock: func() {
				s.mockRepo.On("Delete", mock.MatchedBy(func(meal *models.Meal) bool {
					return meal.ID == 1
				})).Return(nil)
			},
			expectedError: false,
		},
		{
			name:   "DatabaseError",
			mealID: 1,
			setupMock: func() {
				dbErr := apperrors.NewInternalServerErr("Database error", nil)
				s.mockRepo.On("Delete", mock.AnythingOfType("*models.Meal")).Return(dbErr)
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup fresh mocks
			s.SetupTest()

			// Setup mock expectations
			tc.setupMock()

			// Act
			err := s.mealService.Delete(tc.mealID)

			// Assert
			if tc.expectedError {
				s.Error(err)
				if tc.errorPredicate != nil {
					s.True(tc.errorPredicate(err))
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

// Run the test suite
func TestMealServiceSuite(t *testing.T) {
	suite.Run(t, new(MealServiceTestSuite))
}

// MockMealRepository implementation
type MockMealRepository struct {
	mock.Mock
}

func (m *MockMealRepository) GetByID(id uint) (*models.Meal, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Meal), args.Error(1)
}

func (m *MockMealRepository) GetAll() ([]models.Meal, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Meal), args.Error(1)
}

func (m *MockMealRepository) Create(meal *models.Meal) error {
	args := m.Called(meal)
	return args.Error(0)
}

func (m *MockMealRepository) Delete(meal *models.Meal) error {
	args := m.Called(meal)
	return args.Error(0)
}

// WithTransaction implementation for the mock repository
func (m *MockMealRepository) WithTransaction(fn func(txRepo repositories.MealRepository) error) error {
	args := m.Called(fn)

	if args.Get(0) != nil && args.Error(0) != nil {
		return args.Error(0)
	}

	err := fn(m)

	if args.Error(0) != nil {
		return args.Error(0)
	}

	return err
}
