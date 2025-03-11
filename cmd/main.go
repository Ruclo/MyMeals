package main

import (
	"github.com/Ruclo/MyMeals/internal/database"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}

	_ = database.InitDB()

	for {

	}
}