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

func CreateConnection() *gorm.DB {
	conf := config.ConfigInstance

	dbString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", conf.DBHost(), conf.DBUser(),
		conf.DBPassword(), conf.DBName(), conf.DBPort())
	db, err := gorm.Open(postgres.Open(dbString), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Set log level to Info
	})
	if err != nil {
		log.Fatal("Failed to connect to the DB: ", err)
	}

	return db
}

func migrateSchema(db *gorm.DB) {

	err := db.AutoMigrate(&models.Meal{}, &models.Order{}, &models.StaffMember{}, &models.Review{}, &models.OrderMeal{})
	if err != nil {
		log.Fatal("Schema migration failed: ", err)
	}
}

func WipeDB(db *gorm.DB) {
	if err := db.Exec("TRUNCATE TABLE order_meals, reviews, orders, meals, staff_members RESTART IDENTITY CASCADE").Error; err != nil {
		log.Fatal("Failed to truncate tables: ", err)
	}

}

func InitDB() *gorm.DB {
	log.Println("Starting DB Initialization")
	db := CreateConnection()
	migrateSchema(db)

	return db
}
