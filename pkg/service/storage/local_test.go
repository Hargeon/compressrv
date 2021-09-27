package storage

import (
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"log"
	"os"
)

func init() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalln("Can't initialize zap package in local_test", err)
	}

	err = godotenv.Load("../../../.env")
	if err != nil {
		logger.Fatal("Can't read .env file", zap.String("Error", err.Error()))
	}
	root := os.Getenv("ROOT")
	if root == "" {
		logger.Fatal("Invalid ROOT ENV variable")
	}
}
