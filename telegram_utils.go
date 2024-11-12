package main

import (
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"log"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func GetBot(token string, isDebug bool) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bot.Debug = isDebug
	log.Printf("Authorized on account %s", bot.Self.UserName)
	return bot, nil
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func GetUserAvatar(bot *tgbotapi.BotAPI, userId int64) (*image.NRGBA, error) {
	photos, err := bot.GetUserProfilePhotos(tgbotapi.UserProfilePhotosConfig{
		UserID: userId,
		Limit:  1,
	})

	if err != nil {
		return nil, fmt.Errorf("error geting avatar: %s", err)
	}

	if len(photos.Photos) == 0 {
		return nil, errors.New("you don't have avatars")
	}

	largestPhoto := photos.Photos[0][len(photos.Photos[0])-1]
	fileConfig, err := bot.GetFileDirectURL(largestPhoto.FileID)
	if err != nil {
		return nil, fmt.Errorf("error loading avatar: %s", err)
	}

	resp, err := httpClient.Get(fileConfig)
	if err != nil {
		return nil, fmt.Errorf("network error loading avatar: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("network error loading avatar, status: %d, body: %s", resp.StatusCode, string(body))
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		println(err)
		return nil, fmt.Errorf("error decoding avatar: %s", err)
	}

	return ImageToNRGBA(img), nil
}
