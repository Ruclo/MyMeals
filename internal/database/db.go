package database

import (
	"fmt"
	"github.com/Ruclo/MyMeals/internal/config"
	"github.com/Ruclo/MyMeals/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
)

// CreateConnection creates a new database connection using the configuration from the config package
// It exits the program if the connection fails.
func CreateConnection() *gorm.DB {
	conf := config.ConfigInstance

	dbString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		conf.DBHost(),
		conf.DBUser(),
		conf.DBPassword(),
		conf.DBName(),
		conf.DBPort())

	db, err := gorm.Open(postgres.Open(dbString),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	if err != nil {
		log.Fatal("Failed to connect to the DB: ", err)
	}

	return db
}

// migrateSchema migrates the database schema to the latest version.
// It exits the program if the migration fails.
// Make sure to add new models to the migration function.
func migrateSchema(db *gorm.DB) {

	err := db.AutoMigrate(&models.Meal{}, &models.Order{}, &models.User{}, &models.Review{}, &models.OrderMeal{})
	if err != nil {
		log.Fatal("Schema migration failed: ", err)
	}
}

// WipeDB deletes all data from the database.
// It exits the program if the wipe fails.
// Make sure to add new models to the wipe function.
// Used in testing only.
func WipeDB(db *gorm.DB) {
	if err := db.Exec("TRUNCATE TABLE order_meals, reviews, orders, meals, staff_members RESTART IDENTITY CASCADE").Error; err != nil {
		log.Fatal("Failed to truncate tables: ", err)
	}

}

// InitDB creates a new database connection and migrates the schema and returns the database connection.
// Exits the program on failure.
func InitDB() *gorm.DB {
	log.Println("Starting DB Initialization")
	db := CreateConnection()
	migrateSchema(db)

	return db
}
