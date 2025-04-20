package testing

import (
	"fmt"
	"github.com/Ruclo/MyMeals/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"testing"
	"time"
)

// NewTestDB creates an in-memory SQLite database for testing
func NewTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:memdb%d?mode=memory&cache=shared", time.Now().UnixNano())), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	db.Exec("PRAGMA foreign_keys = ON")

	// Migrate the schema
	err = db.AutoMigrate(
		&models.Meal{},
		&models.Order{},
		&models.OrderMeal{},
		&models.Review{},
		&models.StaffMember{},
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
