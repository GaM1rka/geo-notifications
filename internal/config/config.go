package config

import (
	"os"
)

func GetDBURL() string {
	dbURL := os.Getenv("DATABASE_URL")
	return dbURL
}
