package database

import (
	"fmt"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

func getEnvOrExit(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatal(key + " env variable not found")
	}
	return value
}

func CreateConnection() *gorm.DB {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("No .env file found")
	}
	host := getEnvOrExit("DB_HOST")
	user := getEnvOrExit("DB_USER")
	password := getEnvOrExit("DB_PASSWORD")
	name := getEnvOrExit("DB_NAME")
	port := getEnvOrExit("DB_PORT")

	dbString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", host, user, password, name, port)
	db, err := gorm.Open(postgres.Open(dbString), &gorm.Config{})
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
