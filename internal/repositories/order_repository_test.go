package repositories_test

import (
	"github.com/Ruclo/MyMeals/internal/errors"
	testinghelpers "github.com/Ruclo/MyMeals/internal/testing"
	"gorm.io/gorm"
	"slices"
	"time"

	"github.com/stretchr/testify/require"

	"testing"

	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func assertEqualOrders(t *testing.T, expected, actual *models.Order) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.TableNo, actual.TableNo)
	assert.Equal(t, expected.Notes, actual.Notes)
	assert.Equal(t, len(expected.OrderMeals), len(actual.OrderMeals))
	for i, expectedOrderMeal := range expected.OrderMeals {
		actualOrderMeal := actual.OrderMeals[i]
		assert.Equal(t, expectedOrderMeal.MealID, actualOrderMeal.MealID)
		assert.Equal(t, expectedOrderMeal.Quantity, actualOrderMeal.Quantity)
		assert.Equal(t, expectedOrderMeal.Completed, actualOrderMeal.Completed)
	}
	assert.Equal(t, expected.Review, actual.Review)
}

func TestOrderRepository_GetByID(t *testing.T) {
	db := testinghelpers.NewTestDB(t)
	defer testinghelpers.CleanupTestDB(t, db)
	repo := repositories.NewOrderRepository(db)

	_, err := repo.GetByID(999)
	assert.Error(t, err)
	assert.True(t, errors.IsNotFoundErr(err))

	meal := getTestMeal()
	require.NoError(t, db.Create(&meal).Error)

	order := &models.Order{
		TableNo: 5,
		Notes:   "Test notes",
		OrderMeals: []models.OrderMeal{
			{
				MealID:   meal.ID,
				Quantity: 3,
			},
		},
	}

	require.NoError(t, db.Create(order).Error)

	foundOrder, err := repo.GetByID(order.ID)
	assertEqualOrders(t, order, foundOrder)

	comment := "comment"
	review := &models.Review{
		OrderID:   order.ID,
		Comment:   &comment,
		Rating:    5,
		PhotoURLs: []string{"http://example.com/photo1.jpg"},
	}
	require.NoError(t, db.Create(review).Error)
	order.Review = review
	foundOrder, err = repo.GetByID(order.ID)
	assert.NoError(t, err)
	assertEqualOrders(t, order, foundOrder)
	assert.NotNil(t, foundOrder.Review)
	assert.Equal(t, review.Comment, foundOrder.Review.Comment)
	assert.Equal(t, review.Rating, foundOrder.Review.Rating)
	assert.NotNil(t, foundOrder.OrderMeals)
	assert.NotZero(t, len(foundOrder.OrderMeals))
	assert.NotNil(t, foundOrder.OrderMeals[0].Meal)
}

func TestOrderRepository_Create(t *testing.T) {
	db := testinghelpers.NewTestDB(t)
	defer testinghelpers.CleanupTestDB(t, db)
	repo := repositories.NewOrderRepository(db)

	meal := getTestMeal()
	require.NoError(t, db.Create(&meal).Error)
	order := &models.Order{
		TableNo: 5,
		Notes:   "Test notes",
		OrderMeals: []models.OrderMeal{
			{
				MealID:   meal.ID,
				Quantity: 3,
			},
		},
	}

	assert.NoError(t, repo.Create(order))
	assert.NotZero(t, order.ID)

	foundOrder, err := repo.GetByID(order.ID)
	assert.NoError(t, err)
	assertEqualOrders(t, order, foundOrder)

	order2 := &models.Order{
		TableNo: 5,
		Notes:   "Test notes",
	}

	assert.Error(t, repo.Create(order2))
}

func TestOrderRepository_GetOrderMeal(t *testing.T) {
	db := testinghelpers.NewTestDB(t)
	defer testinghelpers.CleanupTestDB(t, db)
	repo := repositories.NewOrderRepository(db)

	meal := getTestMeal()
	require.NoError(t, db.Create(&meal).Error)

	om := models.OrderMeal{
		MealID:   meal.ID,
		Quantity: 3,
	}

	order := &models.Order{
		TableNo: 5,
		Notes:   "Test notes",
		OrderMeals: []models.OrderMeal{
			om,
		},
	}
	require.NoError(t, db.Create(order).Error)

	_, err := repo.GetOrderMeal(order.ID, 999)
	assert.Error(t, err)

	_, err = repo.GetOrderMeal(999, meal.ID)
	assert.Error(t, err)

	orderMeal, err := repo.GetOrderMeal(order.ID, meal.ID)
	assert.NoError(t, err)

	assert.Equal(t, om.MealID, orderMeal.MealID)
	assert.Equal(t, om.Quantity, orderMeal.Quantity)
	assert.Zero(t, orderMeal.Completed)
	assert.Equal(t, order.ID, orderMeal.OrderID)

}

func TestOrderRepository_CreateOrderMeal(t *testing.T) {

	db := testinghelpers.NewTestDB(t)
	defer testinghelpers.CleanupTestDB(t, db)
	repo := repositories.NewOrderRepository(db)

	meal := getTestMeal()
	require.NoError(t, db.Create(&meal).Error)

	order := &models.Order{
		TableNo: 5,
		Notes:   "Test notes",
		OrderMeals: []models.OrderMeal{
			{
				MealID:   meal.ID,
				Quantity: 3,
			},
		},
	}

	require.NoError(t, db.Create(order).Error)

	orderMeal := models.OrderMeal{
		OrderID:  order.ID,
		MealID:   meal.ID,
		Quantity: 2,
	}

	err := repo.CreateOrderMeal(&orderMeal)
	assert.Error(t, err)

	orderMeal.MealID = 999
	err = repo.CreateOrderMeal(&orderMeal)
	assert.Error(t, err)

	meal2 := getTestMeal()
	require.NoError(t, db.Create(&meal2).Error)

	orderMeal.MealID = meal2.ID
	err = repo.CreateOrderMeal(&orderMeal)
	assert.NoError(t, err)

	var foundOrderMeal models.OrderMeal

	assert.NoError(t, db.Model(&models.OrderMeal{}).Where("order_id = ? AND meal_id = ?", order.ID, meal2.ID).First(&foundOrderMeal).Error)

	assert.Equal(t, orderMeal.MealID, foundOrderMeal.MealID)
	assert.Equal(t, orderMeal.Quantity, foundOrderMeal.Quantity)
	assert.Zero(t, foundOrderMeal.Completed)
	assert.Equal(t, order.ID, foundOrderMeal.OrderID)

}

func TestOrderRepository_UpdateOrderMeal(t *testing.T) {

	db := testinghelpers.NewTestDB(t)
	defer testinghelpers.CleanupTestDB(t, db)
	repo := repositories.NewOrderRepository(db)

	meal := getTestMeal()
	require.NoError(t, db.Create(&meal).Error)

	om := models.OrderMeal{
		OrderID:  999,
		MealID:   meal.ID,
		Quantity: 3,
	}

	assert.Error(t, repo.UpdateOrderMeal(&om))

	order := &models.Order{
		TableNo: 5,
		Notes:   "Test notes",
		OrderMeals: []models.OrderMeal{
			om,
		},
	}

	require.NoError(t, db.Create(order).Error)

	om.OrderID = order.ID
	om.Completed = 3
	err := repo.UpdateOrderMeal(&om)
	assert.NoError(t, err)

	var foundOrderMeal models.OrderMeal
	assert.NoError(t, db.Model(&models.OrderMeal{}).Where("order_id = ? AND meal_id = ?", order.ID, meal.ID).First(&foundOrderMeal).Error)
	assert.Equal(t, om.Completed, foundOrderMeal.Completed)

}

func TestOrderRepository_CreateReview(t *testing.T) {
	db := testinghelpers.NewTestDB(t)
	defer testinghelpers.CleanupTestDB(t, db)
	repo := repositories.NewOrderRepository(db)

	meal := getTestMeal()
	require.NoError(t, db.Create(&meal).Error)

	om := models.OrderMeal{
		MealID:   meal.ID,
		Quantity: 3,
	}

	order := &models.Order{
		TableNo: 5,
		Notes:   "Test notes",
		OrderMeals: []models.OrderMeal{
			om,
		},
	}

	comm := "comment"
	review := &models.Review{
		Rating:  5,
		Comment: &comm,
	}
	require.NoError(t, db.Create(order).Error)

	assert.Error(t, repo.CreateReview(review))
	review.OrderID = order.ID
	assert.NoError(t, repo.CreateReview(review))

	var foundReview models.Review
	assert.NoError(t, db.Model(&models.Review{}).Where("order_id = ?", order.ID).First(&foundReview).Error)
	assert.Equal(t, review.Rating, foundReview.Rating)
	assert.Equal(t, review.Comment, foundReview.Comment)
}

func TestOrderRepository_GetOrders(t *testing.T) {
	// Set up the test database
	db := testinghelpers.NewTestDB(t)
	repo := repositories.NewOrderRepository(db)
	// Create test meals
	meals := []*models.Meal{
		getTestMeal(),
		getTestMeal(),
	}

	for _, meal := range meals {
		require.NoError(t, db.Create(meal).Error)
	}

	// Store creation time for reference
	var orderedOrders []*models.Order

	// Create test orders with automatic timestamps
	// Order 1: Fully completed order
	order1 := &models.Order{
		TableNo: 1,
		Notes:   "Test notes",
		OrderMeals: []models.OrderMeal{
			{
				MealID:    meals[0].ID,
				Quantity:  2,
				Completed: 2,
			},
			{
				MealID:    meals[1].ID,
				Quantity:  1,
				Completed: 1,
			},
		},
	}

	require.NoError(t, db.Create(order1).Error)
	assert.NotZero(t, order1.CreatedAt)
	orderedOrders = append(orderedOrders, order1)

	// Order 2: Pending order (partially completed)
	order2 := &models.Order{
		TableNo: 2,
		Notes:   "Test notes",
		OrderMeals: []models.OrderMeal{
			{
				MealID:    meals[0].ID,
				Quantity:  2,
				Completed: 2,
			},
			{
				MealID:   meals[1].ID,
				Quantity: 1,
			},
		},
	}
	require.NoError(t, db.Create(order2).Error)
	assert.NotZero(t, order2.CreatedAt)
	orderedOrders = append(orderedOrders, order2)

	// Order 3: Pending order (not completed at all)
	order3 := &models.Order{
		TableNo: 2,
		Notes:   "Test notes",
		OrderMeals: []models.OrderMeal{
			{
				MealID:    meals[0].ID,
				Quantity:  2,
				Completed: 1,
			},
			{
				MealID:   meals[1].ID,
				Quantity: 1,
			},
		},
	}
	require.NoError(t, db.Create(order3).Error)
	assert.NotZero(t, order3.CreatedAt)
	orderedOrders = append(orderedOrders, order3)

	// Order 4: Fully completed order
	order4 := &models.Order{
		TableNo: 2,
		Notes:   "Test notes",
		OrderMeals: []models.OrderMeal{
			{
				MealID:    meals[0].ID,
				Quantity:  2,
				Completed: 2,
			},
			{
				MealID:    meals[1].ID,
				Quantity:  1,
				Completed: 1,
			},
		},
	}
	require.NoError(t, db.Create(order4).Error)
	assert.NotZero(t, order4.CreatedAt)
	orderedOrders = append(orderedOrders, order4)

	// Order 5: Fully completed order with review
	order5 := &models.Order{
		TableNo: 5,
		Notes:   "Test notes",
		OrderMeals: []models.OrderMeal{
			{
				MealID:    meals[0].ID,
				Quantity:  2,
				Completed: 2,
			},
			{
				MealID:    meals[1].ID,
				Quantity:  1,
				Completed: 1,
			},
		},
	}
	require.NoError(t, db.Create(order5).Error)
	assert.NotZero(t, order5.CreatedAt)
	order5.Review = createReview(t, db, order5.ID, 5, "Excellent service!")

	orderedOrders = append(orderedOrders, order5)
	slices.Reverse(orderedOrders)
	// Test cases
	testCases := []struct {
		name           string
		params         repositories.OrderQueryParams
		expectedOrders []*models.Order
	}{
		{
			name: "Get all orders without filtering",
			params: repositories.OrderQueryParams{
				OlderThan:   time.Time{}, // No time filtering
				PageSize:    0,           // No pagination limit
				OnlyPending: false,
			},
			expectedOrders: orderedOrders,
		},
		{
			name: "Get only pending orders",
			params: repositories.OrderQueryParams{
				OnlyPending: true,
			},
			expectedOrders: []*models.Order{order3, order2}, // Newest first
		},
		{
			name: "Get orders with pagination limit",
			params: repositories.OrderQueryParams{
				PageSize: 2, // Limit to 2 orders
			},
			expectedOrders: []*models.Order{order5, order4},
		},
		{
			name: "Get orders older than order3",
			params: repositories.OrderQueryParams{
				OlderThan: order3.CreatedAt,
			}, // Only order1 and order2 are older
			expectedOrders: []*models.Order{order2, order1}, // Most recent first
		},
		{
			name: "Get pending orders with pagination",
			params: repositories.OrderQueryParams{
				OnlyPending: true,
				PageSize:    1,
			},
			expectedOrders: []*models.Order{order3},
		},
		{
			name: "Zero results when filtering for too old",
			params: repositories.OrderQueryParams{
				OlderThan: order1.CreatedAt.Add(-1 * time.Hour), // Older than the oldest order
			},
			expectedOrders: []*models.Order{},
		},
	}

	// Run test cases
	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			// Call the method
			orders, err := repo.GetOrders(tc.params)

			// Check for errors
			assert.NoError(t, err)

			// Check number of returned orders
			assert.Len(t, orders, len(tc.expectedOrders))

			// If no results expected, skip ID checks
			if len(tc.expectedOrders) == 0 {
				return
			}
			// Verify order IDs match expected
			for i := range orders {
				assert.Equal(t, tc.expectedOrders[i].ID, orders[i].ID)
			}

			// Verify orders have correct relationships loaded
			for _, order := range orders {
				// Check order meals are loaded
				assert.NotEmpty(t, order.OrderMeals)

				// Check review is loaded (may be nil if no review)
				if order.ID == order5.ID {
					assert.NotNil(t, order.Review)
					assert.Equal(t, 5, order.Review.Rating)
				}
			}

			// For pending orders test case, verify they are actually pending
			if tc.params.OnlyPending {
				for _, order := range orders {
					hasPendingMeal := false
					for _, meal := range order.OrderMeals {
						if meal.Completed < meal.Quantity {
							hasPendingMeal = true
							break
						}
					}
					assert.True(t, hasPendingMeal, "Order %d should have at least one pending meal", order.ID)
				}
			}
		})
	}
}

// Helper functions

func createReview(t *testing.T, db *gorm.DB, orderID uint, rating int, comment string) *models.Review {
	review := &models.Review{
		OrderID:   orderID,
		Rating:    rating,
		Comment:   &comment,
		PhotoURLs: []string{"http://example.com/review.jpg"},
	}

	err := db.Create(review).Error
	require.NoError(t, err)

	return review
}

func TestOrderRepository_WithTransaction(t *testing.T) {
	// Setup test database
	db := testinghelpers.NewTestDB(t)
	repo := repositories.NewOrderRepository(db)

	// Create a test meal
	meal := getTestMeal()
	require.NoError(t, db.Create(meal).Error)

	meal2 := getTestMeal()
	require.NoError(t, db.Create(meal2).Error)

	// Create a test order
	order := &models.Order{
		TableNo: 7,
		Notes:   "Complex transaction test order",
		OrderMeals: []models.OrderMeal{
			{
				MealID:   meal.ID,
				Quantity: 3,
			},
		},
	}
	require.NoError(t, db.Create(order).Error)

	// Test that multiple operations in the same transaction succeed together
	t.Run("Multiple Operations Success", func(t *testing.T) {
		err := repo.WithTransaction(func(tx repositories.OrderRepository) error {
			// 1. Create order meal
			orderMeal := &models.OrderMeal{
				OrderID:   order.ID,
				MealID:    meal2.ID,
				Quantity:  3,
				Completed: 0,
			}

			err := tx.CreateOrderMeal(orderMeal)
			if err != nil {
				return err
			}

			// 2. Update the same order meal
			orderMeal.Quantity = 5
			return tx.UpdateOrderMeal(orderMeal)
		})

		// Assert transaction completed without error
		assert.NoError(t, err)

		// Verify both operations were applied
		var orderMeals []models.OrderMeal
		result := db.Where("order_id = ?", order.ID).Find(&orderMeals)
		assert.NoError(t, result.Error)
		assert.Len(t, orderMeals, 2)
		if orderMeals[0].MealID == meal.ID {
			assert.Equal(t, uint(3), orderMeals[0].Quantity)
			assert.Equal(t, uint(5), orderMeals[1].Quantity)
		} else {
			assert.Equal(t, uint(5), orderMeals[0].Quantity)
			assert.Equal(t, uint(3), orderMeals[1].Quantity)

		}
	})

	t.Run("Second Operation Fails", func(t *testing.T) {
		// First delete the existing order meal to start fresh
		err := repo.WithTransaction(func(tx repositories.OrderRepository) error {
			// 1. Create order meal successfully

			order.ID = 0
			err := tx.Create(order)
			require.NoError(t, err)

			// 2. Try to update a non-existent order meal (should fail)
			nonExistentOrderMeal := &models.OrderMeal{
				OrderID:   999, // Non-existent order ID
				MealID:    meal.ID,
				Quantity:  7,
				Completed: 0,
			}

			// This should fail and cause rollback
			return tx.UpdateOrderMeal(nonExistentOrderMeal)
		})

		// Assert transaction failed
		assert.Error(t, err)

		// Verify first operation was rolled back
		var count int64
		db.Model(&models.Order{}).Count(&count)
		assert.Equal(t, int64(1), count, "Transaction should have rolled back, no order meal should exist")
	})
}
