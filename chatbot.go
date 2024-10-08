package main

import (
	"bytes"
	"errors"
	"image/color"
	"log"
	"os"

	"github.com/disintegration/imaging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type VacatoBot struct {
	bot *tgbotapi.BotAPI
}

func (vb *VacatoBot) handleGradient(userId, chatId int64, text string) error {
	userAvatar, err := GetUserAvatar(vb.bot, userId)
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
	_, err = vb.bot.Send(photo)
	return err
}

func (vb *VacatoBot) handleStart(chatId int64) {
	vb.sendMessage(chatId, "Hi! ✨\nUse the /gradient command to apply a stunning gradient and custom text to your avatar!")
}

func (vb *VacatoBot) sendMessage(chatId int64, text string) {
	vb.bot.Send(tgbotapi.NewMessage(chatId, text))
}

func (vb *VacatoBot) Init() {
	command := tgbotapi.BotCommand{
		Command:     "gradient",
		Description: "Applies a gradient and custom text to your avatar. Text limited to two lines.",
	}

	_, err := vb.bot.Request(tgbotapi.NewSetMyCommands(command))
	if err != nil {
		log.Panic(err)
	}
}

func (vb *VacatoBot) Start() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := vb.bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {
			userId := update.Message.From.ID
			chatId := update.Message.Chat.ID

			switch update.Message.Command() {
			case "start":
				vb.handleStart(chatId)
			case "gradient":
				text := update.Message.CommandArguments()
				if text == "" {
					msg := tgbotapi.NewMessage(chatId, "Please enter text after the /gradient command.\nFor example:\n\n/gradient day-off\ntoday")
					vb.bot.Send(msg)
					continue
				}

				go func() {
					err := vb.handleGradient(userId, chatId, text)
					if err != nil {
						vb.sendMessage(chatId, err.Error())
					}
				}()

				continue
			default:
				vb.sendMessage(chatId, "Unknown command")
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command")
				vb.bot.Send(msg)
			}
		}
	}
}

func NewVacatoBot() VacatoBot {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		panic("failed to retrieve the Telegram token from the environment")
	}

	bot, err := GetBot(token, true)

	if err != nil {
		panic(err)
	}

	return VacatoBot{bot}
}
