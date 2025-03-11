package database

import (
	"gorm.io/driver/postgres"
    "gorm.io/gorm"
	"log"
	"os"
	"github.com/Ruclo/MyMeals/internal/models"
	"fmt"
)


func GetEnvOrExit(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatal(key + " env variable not found")
	}
	return value
}

func InitDB() (*gorm.DB) {
	log.Println("Starting DB Initialization")

	host := GetEnvOrExit("DB_HOST")
	user := GetEnvOrExit("DB_USER")
	password := GetEnvOrExit("DB_PASSWORD")
	name := GetEnvOrExit("DB_NAME")
	port := GetEnvOrExit("DB_PORT")

	db_string := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", host, user, password, name, port)
	db, err := gorm.Open(postgres.Open(db_string), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the DB: ", err)
	}
	
	err = db.SetupJoinTable(&models.Order{}, "Meals", &models.OrderMeal{})
	if err != nil {
		log.Fatal("Setting up the join table for orders and meals failed: ", err)
	}
	err = db.AutoMigrate(&models.Meal{}, &models.Order{}, &models.StaffMember{}, &models.Review{})
	if err != nil {
		log.Fatal("Schema migration failed: ", err)
	}
	

	return db
}