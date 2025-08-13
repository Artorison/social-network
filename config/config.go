package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTMLTemplatePath string
	AppPort          string
	MysqlDSN         string
	MongoAddr        string
	MongoDBName      string
}

func MustLoad() *Config {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	return &Config{
		HTMLTemplatePath: os.Getenv("HTML_TEMPLATE"),
		AppPort:          os.Getenv("APP_PORT"),
		MysqlDSN:         os.Getenv("MYSQL_DSN"),
		MongoAddr:        os.Getenv("MONGO_ADDR"),
		MongoDBName:      os.Getenv("MONGO_DB_NAME"),
	}
}
