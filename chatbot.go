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
		log.Printf("[ERR] error getting user avatar: %v", err)
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

func (vb *VacatoBot) sendMessage(chatId int64, text string) {
	_, err := vb.bot.Send(tgbotapi.NewMessage(chatId, text))
	if err != nil {
		log.Printf("[ERR] failed to send message to chat %d: %v", chatId, err)
	}
}

func (vb *VacatoBot) handleMenu(chatId int64) {
	msg := tgbotapi.NewMessage(chatId, "Tap the button below and tell me what text you'd like on your avatar.")
	button := tgbotapi.NewInlineKeyboardButtonData("Add text to my avatar", "request_text")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(button),
	)
	msg.ReplyMarkup = keyboard
	vb.bot.Send(msg)
}

func (vb *VacatoBot) Init() {
	commands := []tgbotapi.BotCommand{
		{Command: "menu", Description: "Show menu"},
		{Command: "avatar", Description: "Add some text to your avatar"},
	}

	_, err := vb.bot.Request(tgbotapi.NewSetMyCommands(commands...))
	if err != nil {
		log.Panic(err)
	}
}

const requestTextMsg = "What would you like to add to your avatar?\n" +
	"You can enter up to two lines, like 'On vacation!' or just 'Day off!'\n" +
	"Please reply directly to this message with your text!"

func (vb *VacatoBot) Start() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := vb.bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil && update.Message.IsCommand() {
			chatId := update.Message.Chat.ID

			switch update.Message.Command() {
			case "start":
				vb.sendMessage(chatId, "Hey there! Want to add a fun message to your avatar? Use /menu to get started!")
				vb.handleMenu(chatId)

			case "menu":
				vb.handleMenu(chatId)

			case "avatar":
				vb.sendMessage(chatId, requestTextMsg)

			default:
				vb.sendMessage(chatId, "Oops! I don't recognize that command. Try something else!")
			}
			continue
		}

		if update.CallbackQuery != nil {
			if update.CallbackQuery.Data == "request_text" {
				vb.sendMessage(update.CallbackQuery.Message.Chat.ID, requestTextMsg)
			}

			continue
		}

		if update.Message != nil && update.Message.ReplyToMessage != nil &&
			update.Message.ReplyToMessage.Text == requestTextMsg {

			text := update.Message.Text
			chatId := update.Message.Chat.ID
			userId := update.Message.From.ID

			go func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[ERR] panic in handleGradient: %v", r)
					}
				}()

				err := vb.handleGradient(userId, chatId, text)
				if err != nil {
					vb.sendMessage(chatId, "Oh no! Something went wrong. Try again, please!\n\n"+err.Error())
				}
			}()
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
