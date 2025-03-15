package config

import (
	"log"
	"os"
)

var ConfigInstance Config

type Config struct {
	dbHost     string
	dbUser     string
	dbPassword string
	dbName     string
	dbPort     string
	jwtSecret  []byte
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

func InitConfig() {
	ConfigInstance = Config{}
	ConfigInstance.dbHost = getEnvOrExit("DB_HOST")
	ConfigInstance.dbUser = getEnvOrExit("DB_USER")
	ConfigInstance.dbPassword = getEnvOrExit("DB_PASSWORD")
	ConfigInstance.dbName = getEnvOrExit("DB_NAME")
	ConfigInstance.dbPort = getEnvOrExit("DB_PORT")
	ConfigInstance.jwtSecret = []byte(getEnvOrExit("JWT_SECRET_KEY"))

}

func getEnvOrExit(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatal(key + " env variable not found")
	}
	return value
}
