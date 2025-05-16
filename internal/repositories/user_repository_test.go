package repositories_test

import (
	"github.com/Ruclo/MyMeals/internal/apperrors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	testinghelpers "github.com/Ruclo/MyMeals/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestUserRepository_GetByUsername(t *testing.T) {
	t.Run("existing user", func(t *testing.T) {
		// Setup
		db := testinghelpers.NewTestDB(t)
		defer testinghelpers.CleanupTestDB(t, db)

		// Create a test user directly in the database
		testUser := &models.User{
			Username: "existinguser",
			Password: "password123",
		}
		require.NoError(t, db.Create(testUser).Error)

		repo := repositories.NewUserRepository(db)

		// Execute
		user, err := repo.GetByUsername("existinguser")

		// Verify
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "existinguser", user.Username)
		assert.Equal(t, "password123", user.Password)
		assert.Equal(t, models.RegularStaffRole, user.Role)
	})

	t.Run("non-existing user", func(t *testing.T) {
		// Setup
		db := testinghelpers.NewTestDB(t)
		defer testinghelpers.CleanupTestDB(t, db)

		repo := repositories.NewUserRepository(db)

		// Execute
		user, err := repo.GetByUsername("nonexistinguser")

		// Verify
		assert.Nil(t, user)
		assert.Error(t, err)

		// Check that it's the right type of error
		var appErr *apperrors.AppError
		assert.ErrorAs(t, err, &appErr)
		assert.Equal(t, http.StatusNotFound, appErr.StatusCode)
	})
}

func TestUserRepository_Create(t *testing.T) {
	t.Run("create new user", func(t *testing.T) {
		// Setup
		db := testinghelpers.NewTestDB(t)
		defer testinghelpers.CleanupTestDB(t, db)

		repo := repositories.NewUserRepository(db)

		newUser := &models.User{
			Username: "newuser",
			Password: "password123",
		}

		// Execute
		err := repo.Create(newUser)

		// Verify
		require.NoError(t, err)

		// Verify user was created in the database
		var found models.User
		result := db.Where("username = ?", "newuser").First(&found)
		require.NoError(t, result.Error)
		assert.Equal(t, "newuser", found.Username)
		assert.Equal(t, "password123", found.Password)
		assert.Equal(t, models.RegularStaffRole, found.Role)
	})

	t.Run("create existing user", func(t *testing.T) {
		// Setup
		db := testinghelpers.NewTestDB(t)
		defer testinghelpers.CleanupTestDB(t, db)

		// Create a test user directly in the database
		existingUser := &models.User{
			Username: "existinguser",
			Password: "password123",
		}
		require.NoError(t, db.Create(existingUser).Error)

		repo := repositories.NewUserRepository(db)

		duplicateUser := &models.User{
			Username: "existinguser", // Same username as existing user
			Password: "newpassword",
			Role:     models.AdminRole,
		}

		// Execute
		err := repo.Create(duplicateUser)

		// Verify
		assert.Error(t, err)

	})
}

func TestUserRepository_Update(t *testing.T) {
	t.Run("update existing user", func(t *testing.T) {
		// Setup
		db := testinghelpers.NewTestDB(t)
		defer testinghelpers.CleanupTestDB(t, db)

		// Create a test user directly in the database
		existingUser := &models.User{
			Username: "existinguser",
			Password: "password123",
			Role:     models.AdminRole,
		}
		require.NoError(t, db.Create(existingUser).Error)

		repo := repositories.NewUserRepository(db)

		updatedUser := &models.User{
			Username: "existinguser", // Same username to find the record
			Password: "newpassword",
		}

		// Execute
		err := repo.Update(updatedUser)

		// Verify
		require.NoError(t, err)

		var found models.User
		result := db.Where("username = ?", "existinguser").First(&found)
		require.NoError(t, result.Error)
		assert.Equal(t, "existinguser", found.Username)
		assert.Equal(t, "newpassword", found.Password)
		assert.Equal(t, models.AdminRole, found.Role)
	})

	t.Run("update non-existing user", func(t *testing.T) {
		// Setup
		db := testinghelpers.NewTestDB(t)
		defer testinghelpers.CleanupTestDB(t, db)

		repo := repositories.NewUserRepository(db)

		nonExistingUser := &models.User{
			Username: "nonexistinguser",
			Password: "newpassword",
		}

		// Execute
		err := repo.Update(nonExistingUser)

		// Verify
		assert.Error(t, err)

		// Check that it's the right type of error
		var appErr *apperrors.AppError
		assert.ErrorAs(t, err, &appErr)
		assert.Equal(t, http.StatusNotFound, appErr.StatusCode)
	})
}

func TestUserRepository_Integration(t *testing.T) {
	// Setup
	db := testinghelpers.NewTestDB(t)
	defer testinghelpers.CleanupTestDB(t, db)

	repo := repositories.NewUserRepository(db)

	// 1. Create a new user
	newUser := &models.User{
		Username: "testintegration",
		Password: "initialpassword",
	}

	err := repo.Create(newUser)
	require.NoError(t, err, "Should create user successfully")

	// 2. Try to create the same user again
	duplicateUser := &models.User{
		Username: "testintegration",
		Password: "anotherpassword",
		Role:     models.AdminRole,
	}

	err = repo.Create(duplicateUser)
	require.Error(t, err, "Should fail creating duplicate user")

	// 3. Get the user by username
	foundUser, err := repo.GetByUsername("testintegration")
	require.NoError(t, err, "Should find the user")
	assert.Equal(t, "testintegration", foundUser.Username)
	assert.Equal(t, "initialpassword", foundUser.Password)
	assert.Equal(t, models.RegularStaffRole, foundUser.Role)

	// 4. Replace the user
	foundUser.Password = "updatedpassword"
	foundUser.Role = models.AdminRole

	err = repo.Update(foundUser)
	require.NoError(t, err, "Should update the user")

	// 5. Get the updated user
	updatedUser, err := repo.GetByUsername("testintegration")
	require.NoError(t, err, "Should find the updated user")
	assert.Equal(t, "testintegration", updatedUser.Username)
	assert.Equal(t, "updatedpassword", updatedUser.Password)
	assert.Equal(t, models.AdminRole, updatedUser.Role)

	// 6. Try to get a non-existing user
	nonExistingUser, err := repo.GetByUsername("nonexistinguser")
	require.Error(t, err, "Should error on non-existing user")
	assert.Nil(t, nonExistingUser)

	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.StatusCode)

	// 7. Try to update a non-existing user
	badUser := &models.User{
		Username: "nonexistinguser",
		Password: "somepassword",
		Role:     models.RegularStaffRole,
	}

	err = repo.Update(badUser)
	require.Error(t, err, "Should error on updating non-existing user")
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.StatusCode)
}
