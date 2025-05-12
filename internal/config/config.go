package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

// ConfigInstance is the global config instance.
// It is initialized in InitConfig() and should be used everywhere else.
var ConfigInstance Config

// Config represents the configuration of the application.
// A single global instance of Config is used throughout the application and is initialized in InitConfig().
type Config struct {
	dbHost        string
	dbUser        string
	dbPassword    string
	dbName        string
	dbPort        string
	jwtSecret     []byte
	cloudinaryUrl string
}

// DBHost returns the host of the database.
func (c *Config) DBHost() string {
	return c.dbHost
}

// DBUser returns the username of the database user.
func (c *Config) DBUser() string {
	return c.dbUser
}

// DBPassword returns the password of the database user.
func (c *Config) DBPassword() string {
	return c.dbPassword
}

// DBName returns the name of the database.
func (c *Config) DBName() string {
	return c.dbName
}

// DBPort returns the port of the database.
func (c *Config) DBPort() string {
	return c.dbPort
}

// JWTSecret returns the secret used to sign and verify JWTs.
func (c *Config) JWTSecret() []byte {
	return c.jwtSecret
}

// CloudinaryUrl returns the url of the cloudinary account.
// This is used to upload images to cloudinary and retrieve the url of the uploaded image.
func (c *Config) CloudinaryUrl() string {
	return c.cloudinaryUrl
}

// InitConfig initializes the config instance with values from the .env file.
// It exits the program if the .env file is not found or if any of the required
// environment variables are not set.
//
// The .env file should be located in the root directory of the project.
//
// The required environment variables are listed in .env.example
func InitConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ConfigInstance = Config{}
	ConfigInstance.dbHost = getEnvOrExit("DB_HOST")
	ConfigInstance.dbUser = getEnvOrExit("DB_USER")
	ConfigInstance.dbPassword = getEnvOrExit("DB_PASSWORD")
	ConfigInstance.dbName = getEnvOrExit("DB_NAME")
	ConfigInstance.dbPort = getEnvOrExit("DB_PORT")
	ConfigInstance.jwtSecret = []byte(getEnvOrExit("JWT_SECRET"))
	ConfigInstance.cloudinaryUrl = getEnvOrExit("CLOUDINARY_URL")

}

// getEnvOrExit returns the value of the environment variable with the given key,
// or exits the program if it is not set.
func getEnvOrExit(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatal(key + " env variable not found")
	}
	return value
}
