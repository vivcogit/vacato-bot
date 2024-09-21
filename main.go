package main

import (
	"bytes"
	"image/color"
	"log"

	"github.com/disintegration/imaging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
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
				userAvatar, err := GetUserAvatar(bot, update.Message.From.ID)
				if err != nil {
					log.Printf("[ERR] error geting user avatar: %v", err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
					bot.Send(msg)
					continue
				}

				gradient := CreateGradient(userAvatar.Bounds().Dx(), userAvatar.Bounds().Dy(), color.NRGBA{R: 0, G: 0, B: 255, A: 255}, color.NRGBA{R: 255, G: 0, B: 255, A: 255})

				imgWithGradient := OverlayImage(userAvatar, gradient, 0.5)

				var buf bytes.Buffer
				err = imaging.Encode(&buf, imgWithGradient, imaging.PNG)
				if err != nil {
					log.Printf("[ERR] error saving imgWithGradient: %v", err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error due overlaying.")
					bot.Send(msg)
					continue
				}

				photo := tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FileBytes{
					Name:  "avatar_with_gradient.png",
					Bytes: buf.Bytes(),
				})
				bot.Send(photo)

			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command")
				bot.Send(msg)
			}
		}
	}
}
