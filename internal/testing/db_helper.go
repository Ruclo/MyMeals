package testing

import (
	"github.com/Ruclo/MyMeals/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"testing"
)

// NewTestDB creates an in-memory SQLite database for testing
func NewTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(
		&models.Meal{},
		&models.Order{},
		&models.OrderMeal{},
		&models.Review{},
		// Add other models
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// CleanupTestDB cleans up the test database
func CleanupTestDB(t *testing.T, db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Warning: Failed to get database connection: %v", err)
		return
	}
	sqlDB.Close()
}
