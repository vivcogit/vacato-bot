package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func GetToken() (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", fmt.Errorf("error loading .env file: %v", err)
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		return "", errors.New("failed to retrieve the Telegram token from the .env file")
	}

	return botToken, nil
}
