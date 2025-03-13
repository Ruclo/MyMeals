package database

import (
	"fmt"
	"github.com/Ruclo/MyMeals/internal/models"
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

func createConnection() *gorm.DB {
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
	err := db.SetupJoinTable(&models.Order{}, "Meals", &models.OrderMeal{})
	if err != nil {
		log.Fatal("Setting up the join table for orders and meals failed: ", err)
	}
	err = db.AutoMigrate(&models.Meal{}, &models.Order{}, &models.StaffMember{}, &models.Review{})
	if err != nil {
		log.Fatal("Schema migration failed: ", err)
	}
}

func wipeDB(db *gorm.DB) {
	err := db.Migrator().DropTable(&models.OrderMeal{}, &models.Review{}, &models.Order{}, &models.Meal{}, &models.StaffMember{})
	if err != nil {
		log.Fatal("Failed to wipe the DB: ", err)
	}
}

func InitDB() *gorm.DB {
	log.Println("Starting DB Initialization")

	db := createConnection()
	migrateSchema(db)

	return db
}

func InitTestDB() *gorm.DB {
	log.Println("Starting DB Initialization")
	log.Println("Wiping DB")

	db := createConnection()
	wipeDB(db)
	migrateSchema(db)
	return db
}
