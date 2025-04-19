package repositories

import (
	"github.com/Ruclo/MyMeals/internal/config"
	"github.com/Ruclo/MyMeals/internal/database"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
	"time"
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
	config.InitConfig()
	s.db = database.CreateConnection()
	s.repo = NewOrderRepository(s.db)
}

func (s *OrderRepoTestSuite) SetupTest() {
	database.WipeDB(s.db)
}

func (s *OrderRepoTestSuite) TestCreate_ValidOrder() {
	s.SetupTest()

	meal := CreateValidMeal()
	meal2 := CreateValidMeal()

	err := s.db.Create(meal).Error
	require.NoError(s.T(), err)

	err = s.db.Create(meal2).Error
	require.NoError(s.T(), err)

	order := &models.Order{
		TableNo: 5,
		Name:    "Customer Name",
		Notes:   "Some notes",
		OrderMeals: []models.OrderMeal{
			{
				MealID:   meal.ID,
				Quantity: 2,
			},
			{
				MealID:   meal2.ID,
				Quantity: 1,
			},
		},
	}

	err = s.repo.Create(order)
	assert.NoError(s.T(), err)

	foundOrder := models.Order{}

	err = s.db.Model(&models.Order{}).Preload("OrderMeals.Meal").First(&foundOrder, order.ID).Error

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), order.TableNo, foundOrder.TableNo)
	assert.Equal(s.T(), order.Name, foundOrder.Name)
	assert.Equal(s.T(), order.Notes, foundOrder.Notes)
	assert.Equal(s.T(), len(order.OrderMeals), len(foundOrder.OrderMeals))

}

func (s *OrderRepoTestSuite) TestCreate_InvalidOrders() {
	s.SetupTest()

	s.Run("Invalid table number", func() {
		meal := CreateValidMeal()
		err := s.db.Create(meal).Error
		require.NoError(s.T(), err)

		// Create order with invalid table number (0)
		order := &models.Order{
			TableNo: 0, // Invalid: table number should be >= 1
			Name:    "Customer Name",
			Notes:   "Some notes",
			OrderMeals: []models.OrderMeal{
				{
					MealID:   meal.ID,
					Quantity: 2,
				},
			},
		}

		// Attempt to create the order
		err = s.repo.Create(order)
		assert.Error(s.T(), err, "Should reject order with invalid table number")

	})

	s.Run("CreatedAtSpecified", func() {
		meal := CreateValidMeal()
		err := s.db.Create(meal).Error
		require.NoError(s.T(), err)

		time1 := time.Now().Add(time.Hour * -24)

		order := &models.Order{
			TableNo: 1,
			Name:    "Customer Name",
			Notes:   "Some notes",
			OrderMeals: []models.OrderMeal{
				{
					MealID:   meal.ID,
					Quantity: 2,
				},
			},
			CreatedAt: time1,
		}

		// Attempt to create the order
		err = s.repo.Create(order)
		assert.NoError(s.T(), err, "Should reject order with invalid table number")
		assert.NotEqual(s.T(), order.CreatedAt, time1)

	})

	s.Run("Empty meals", func() {
		order := &models.Order{
			TableNo:    5,
			Name:       "Customer Name",
			Notes:      "Some notes",
			OrderMeals: []models.OrderMeal{}, // Empty meals array
		}

		// Attempt to create the order
		err := s.repo.Create(order)
		assert.Error(s.T(), err, "Should reject order with no meals")

	})
	s.Run("1 Valid Meal, 1 Invalid Meal", func() {
		meal := CreateValidMeal()
		err := s.db.Create(meal).Error
		require.NoError(s.T(), err)

		// Create order with one valid and one invalid meal (invalid quantity)
		order := &models.Order{
			TableNo: 5,
			Name:    "Customer Name",
			Notes:   "Some notes",
			OrderMeals: []models.OrderMeal{
				{
					MealID:   meal.ID,
					Quantity: 2, // Valid
				},
				{
					MealID:   9999,
					Quantity: 0, // Invalid: quantity should be >= 1
				},
			},
		}

		// Attempt to create the order
		err = s.repo.Create(order)
		assert.Error(s.T(), err, "Should reject order with invalid meal quantity")

	})
	s.Run("Completed set to done", func() {
		meal := CreateValidMeal()
		err := s.db.Create(meal).Error
		require.NoError(s.T(), err)

		// Create order with status explicitly set to Done
		order := &models.Order{
			TableNo: 5,
			Name:    "Customer Name",
			Notes:   "Some notes",
			OrderMeals: []models.OrderMeal{
				{
					MealID:    meal.ID,
					Quantity:  2,
					Completed: 4, // Explicitly set to Done (should be overridden to Pending)
				},
			},
		}

		// Create the order
		err = s.repo.Create(order)
		assert.NoError(s.T(), err, "Should accept order even with Completed set to Done")

		// Check that the status was changed to Pending
		foundOrder := models.Order{}
		err = s.db.Model(&models.Order{}).Preload("OrderMeals.Meal").First(&foundOrder, order.ID).Error
		assert.NoError(s.T(), err)

		// Verify status was changed to Pending
		assert.Equal(s.T(), uint(0), foundOrder.OrderMeals[0].Completed,
			"Completed should have been changed to Pending")

	})
}

func (s *OrderRepoTestSuite) TestGetAllPendingOrders() {
	s.SetupTest()

	// Create meals to use in our orders
	meal1 := CreateValidMeal()
	meal2 := CreateValidMeal()
	meal3 := CreateValidMeal()

	err := s.db.Create(meal1).Error
	require.NoError(s.T(), err)
	err = s.db.Create(meal2).Error
	require.NoError(s.T(), err)
	err = s.db.Create(meal3).Error
	require.NoError(s.T(), err)

	// Create multiple orders - all should be pending by default
	orders := []*models.Order{
		{
			TableNo: 5,
			Name:    "Customer One",
			Notes:   "First customer",
			OrderMeals: []models.OrderMeal{
				{
					MealID:   meal1.ID,
					Quantity: 2,
				},
			},
		},
		{
			TableNo: 3,
			Name:    "Customer Two",
			Notes:   "Second customer",
			OrderMeals: []models.OrderMeal{
				{
					MealID:   meal2.ID,
					Quantity: 1,
				},
				{
					MealID:   meal3.ID,
					Quantity: 3,
				},
			},
		},
		{
			TableNo: 8,
			Name:    "Customer Three",
			Notes:   "Third customer",
			OrderMeals: []models.OrderMeal{
				{
					MealID:   meal1.ID,
					Quantity: 1,
				},
				{
					MealID:   meal2.ID,
					Quantity: 1,
				},
			},
		},
	}

	// Create all orders
	for _, order := range orders {
		err := s.repo.Create(order)
		require.NoError(s.T(), err)
	}

	// Get all pending orders
	pendingOrders, err := s.repo.GetAllPendingOrders()

	// Assertions
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), len(orders), len(pendingOrders), "Should retrieve all pending orders")

	// Create a map of expected order IDs for easier verification
	expectedOrderIDs := make(map[uint]bool)
	for _, order := range orders {
		expectedOrderIDs[order.ID] = true
	}

	// Verify that all orders were returned
	for _, order := range pendingOrders {
		assert.True(s.T(), expectedOrderIDs[order.ID], "Returned order should be in expected set")

		// Verify that OrderMeals were properly loaded
		assert.True(s.T(), len(order.OrderMeals) > 0, "Order meals should be loaded")

		// Verify each meal is in pending status
		for _, meal := range order.OrderMeals {
			assert.Equal(s.T(), uint(0), meal.Completed, "All order meals should have pending status")
		}
	}

	// Additional check for data integrity - all fields should be populated
	for _, order := range pendingOrders {
		assert.NotZero(s.T(), order.ID)
		assert.NotZero(s.T(), order.TableNo)
		assert.NotEmpty(s.T(), order.Name)
		assert.NotEmpty(s.T(), order.CreatedAt)
		assert.NotZero(s.T(), order.OrderMeals[0].OrderID)
	}

}

func (s *OrderRepoTestSuite) TestAddReview() {
	s.SetupTest()

	// Create a meal to use in our test order
	meal := CreateValidMeal()
	err := s.db.Create(meal).Error
	require.NoError(s.T(), err)

	// Create a valid order to review
	validOrder := &models.Order{
		TableNo: 5,
		Name:    "Customer Name",
		Notes:   "Some notes",
		OrderMeals: []models.OrderMeal{
			{
				MealID:   meal.ID,
				Quantity: 2,
			},
		},
	}

	err = s.repo.Create(validOrder)
	require.NoError(s.T(), err)

	comment := "Great service and food!"
	s.Run("Valid Review", func() {
		// Create a valid review
		review := &models.Review{
			OrderID:   validOrder.ID,
			Rating:    4,
			Comment:   &comment,
			PhotoURLs: pq.StringArray{"url1", "url2", "url3"},
		}

		// Add review
		err = s.repo.CreateReview(review)
		assert.NoError(s.T(), err)

		// Verify review was added
		var foundOrder models.Order
		err = s.db.Preload("Review").First(&foundOrder, validOrder.ID).Error
		assert.NoError(s.T(), err)

		// Assertions
		assert.NotNil(s.T(), foundOrder.Review)
		assert.Equal(s.T(), review.Rating, foundOrder.Review.Rating)
		assert.Equal(s.T(), *review.Comment, *foundOrder.Review.Comment)
		assert.Equal(s.T(), validOrder.ID, foundOrder.Review.OrderID)
		assert.Equal(s.T(), review.PhotoURLs, foundOrder.Review.PhotoURLs)
	})

	s.Run("Empty Comment", func() {
		// Create a new order since we can't review the same order twice
		emptyCommentOrder := &models.Order{
			TableNo: 6,
			Name:    "Another Customer",
			Notes:   "Some notes",
			OrderMeals: []models.OrderMeal{
				{
					MealID:   meal.ID,
					Quantity: 1,
				},
			},
		}
		err = s.repo.Create(emptyCommentOrder)
		require.NoError(s.T(), err)

		// Create review with empty comment
		comment := ""
		review := &models.Review{
			OrderID: emptyCommentOrder.ID,
			Rating:  5,
			Comment: &comment, // Empty comment
		}

		// Add review
		err = s.repo.CreateReview(review)
		assert.NoError(s.T(), err, "Empty comments should be allowed")

		// Verify review was added
		var foundOrder models.Order
		err = s.db.Preload("Review").First(&foundOrder, emptyCommentOrder.ID).Error
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), (*string)(nil), foundOrder.Review.Comment)
	})

	s.Run("Invalid Rating (Too Low)", func() {
		// Create review with rating too low
		comment := "Good food"
		review := &models.Review{
			OrderID: validOrder.ID,
			Rating:  0, // Invalid: should be 1-5
			Comment: &comment,
		}

		// Add review
		err = s.repo.CreateReview(review)
		assert.Error(s.T(), err, "Should reject review with rating lower than 1")
		assert.Contains(s.T(), err.Error(), "rating", "Error should mention invalid rating")
	})

	s.Run("Invalid Rating (Too High)", func() {
		// Create review with rating too high
		comment := "Excellent food"
		review := &models.Review{
			OrderID: validOrder.ID,
			Rating:  6, // Invalid: should be 1-5
			Comment: &comment,
		}

		// Add review
		err = s.repo.CreateReview(review)
		assert.Error(s.T(), err, "Should reject review with rating higher than 5")
		assert.Contains(s.T(), err.Error(), "rating", "Error should mention invalid rating")
	})

	s.Run("Invalid Order ID", func() {
		// Create review with non-existent order ID
		invalidOrderID := uint(9999) // Assuming this order doesn't exist
		comment := "Great food"
		review := &models.Review{
			OrderID: invalidOrderID,
			Rating:  4,
			Comment: &comment,
		}

		// Add review
		err = s.repo.CreateReview(review)
		assert.Error(s.T(), err, "Should reject review for non-existent order")
		assert.Contains(s.T(), err.Error(), "order", "Error should mention the order not found")
	})

	s.Run("Review Already Exists", func() {
		// Create a new order
		newOrder := &models.Order{
			TableNo: 7,
			Name:    "Yet Another Customer",
			Notes:   "Some notes",
			OrderMeals: []models.OrderMeal{
				{
					MealID:   meal.ID,
					Quantity: 1,
				},
			},
		}
		err = s.repo.Create(newOrder)
		require.NoError(s.T(), err)

		// Add first review
		comment := "Good food"
		firstReview := &models.Review{
			OrderID: newOrder.ID,
			Rating:  4,
			Comment: &comment,
		}
		err = s.repo.CreateReview(firstReview)
		require.NoError(s.T(), err)

		// Try to add second review for same order
		comment2 := "Changed my mind, excellent food"
		secondReview := &models.Review{
			OrderID: newOrder.ID,
			Rating:  5,
			Comment: &comment2,
		}
		err = s.repo.CreateReview(secondReview)
		assert.Error(s.T(), err, "Should not allow multiple reviews for the same order")

		// Verify first review remains unchanged
		var foundOrder models.Order
		err = s.db.Preload("Review").First(&foundOrder, newOrder.ID).Error
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), firstReview.Rating, foundOrder.Review.Rating)
		assert.Equal(s.T(), *firstReview.Comment, *foundOrder.Review.Comment)
	})
}

func (s *OrderRepoTestSuite) TestUpdateStatus() {
	s.SetupTest()
	meal := CreateValidMeal()
	err := s.db.Create(meal).Error
	require.NoError(s.T(), err)

	order := models.Order{
		TableNo: 5,
		Name:    "Customer Name",
		Notes:   "Some notes",
		OrderMeals: []models.OrderMeal{
			{
				MealID:   meal.ID,
				Quantity: 2,
			},
		},
	}

	err = s.repo.Create(&order)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), uint(0), order.OrderMeals[0].Completed)

	order2, err := s.repo.MarkCompleted(order.ID, order.OrderMeals[0].MealID)

	assert.NoError(s.T(), err)

	assert.Equal(s.T(), order2.OrderMeals[0].Quantity, order2.OrderMeals[0].Completed)

}

func (s *OrderRepoTestSuite) TestGetAllOrders() {
	s.SetupTest()

	s.Run("No orders", func() {
		orders, err := s.repo.GetOrders(time.Time{}, 0)
		assert.NoError(s.T(), err)
		assert.Empty(s.T(), orders)
	})

	meal := CreateValidMeal()
	err := s.db.Create(meal).Error

	s.Require().NoError(err)

	var orders []models.Order
	orderCount := 13

	for i := 0; i < orderCount; i++ {
		order := models.Order{
			TableNo: 5,
			Name:    "Customer Name",
			Notes:   "Some notes",
			OrderMeals: []models.OrderMeal{
				{
					MealID:   meal.ID,
					Quantity: 2,
				},
			},
		}
		err = s.repo.Create(&order)
		s.Require().NoError(err)
		orders = append([]models.Order{order}, orders...)
	}

	s.Run("Page size bigger than total", func() {
		fetchedOrders, err := s.repo.GetOrders(time.Time{}, 20)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), orderCount, len(orders))
		for i := range fetchedOrders {
			assert.Equal(s.T(), orders[i].ID, fetchedOrders[i].ID)
		}
	})

	s.Run("Page size smaller than total", func() {

		fetchedOrderCount := 0
		lastTime := time.Time{}
		pageSize := uint(5)
		for fetchedOrderCount < orderCount {
			fetchedOrders, err := s.repo.GetOrders(lastTime, pageSize)
			assert.NoError(s.T(), err)
			assert.Equal(s.T(), min(pageSize, uint(orderCount-fetchedOrderCount)), uint(len(fetchedOrders)))
			for i := range fetchedOrders {
				assert.Equal(s.T(), orders[fetchedOrderCount+i].ID, fetchedOrders[i].ID)
			}
			fetchedOrderCount += len(fetchedOrders)
			lastTime = fetchedOrders[len(fetchedOrders)-1].CreatedAt
		}
	})

}
