package main

import (
	"bytes"
	"errors"
	"image/color"
	"log"

	"github.com/disintegration/imaging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleGradient(bot *tgbotapi.BotAPI, userId, chatId int64, text string) error {
	userAvatar, err := GetUserAvatar(bot, userId)
	if err != nil {
		log.Printf("[ERR] error geting user avatar: %v", err)
		return err
	}

	gradient := CachedCreateGradient(
		userAvatar.Bounds().Dx(), userAvatar.Bounds().Dy(),
		color.NRGBA{R: 0, G: 0, B: 255, A: 255},
		color.NRGBA{R: 255, G: 0, B: 255, A: 255},
	)

	OverlayImage(userAvatar, gradient, 0.5)
	DrawTextToImage(userAvatar, text)

	var buf bytes.Buffer
	err = imaging.Encode(&buf, userAvatar, imaging.PNG)
	if err != nil {
		log.Printf("[ERR] error saving image: %v", err)
		return errors.New("error due overlaying")
	}

	photo := tgbotapi.NewPhoto(chatId, tgbotapi.FileBytes{
		Name:  "avatar_with_gradient.png",
		Bytes: buf.Bytes(),
	})
	_, err = bot.Send(photo)
	return err
}

func Chatbot() {
	token, err := GetToken()
	if err != nil {
		panic(err)
	}

	bot, err := GetBot(token, true)

	if err != nil {
		panic(err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "gradient":
				userId := update.Message.From.ID
				chatId := update.Message.Chat.ID

				text := update.Message.CommandArguments()
				if text == "" {
					msg := tgbotapi.NewMessage(chatId, "Please enter text after the /gradient command")
					bot.Send(msg)
					continue
				}

				err := handleGradient(bot, userId, chatId, text)
				if err != nil {
					msg := tgbotapi.NewMessage(chatId, err.Error())
					bot.Send(msg)
				}
				continue
			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command")
				bot.Send(msg)
			}
		}
	}
}
