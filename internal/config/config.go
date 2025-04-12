package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

var ConfigInstance Config

type Config struct {
	dbHost        string
	dbUser        string
	dbPassword    string
	dbName        string
	dbPort        string
	jwtSecret     []byte
	cloudinaryUrl string
}

func (c *Config) DBHost() string {
	return c.dbHost
}

func (c *Config) DBUser() string {
	return c.dbUser
}

func (c *Config) DBPassword() string {
	return c.dbPassword
}

func (c *Config) DBName() string {
	return c.dbName
}

func (c *Config) DBPort() string {
	return c.dbPort
}

func (c *Config) JWTSecret() []byte {
	return c.jwtSecret
}

func (c *Config) CloudinaryUrl() string {
	return c.cloudinaryUrl
}
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

func getEnvOrExit(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatal(key + " env variable not found")
	}
	return value
}
