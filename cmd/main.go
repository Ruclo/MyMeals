package main

import (
	"github.com/Ruclo/MyMeals/internal/config"
	"github.com/Ruclo/MyMeals/internal/database"
)

func main() {
	config.InitConfig()

	_ = database.InitDB()

}
