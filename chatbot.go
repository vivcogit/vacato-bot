package main

import (
	"bytes"
	"errors"
	"image/color"
	"os"

	"github.com/disintegration/imaging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type VacatoBot struct {
	bot    *tgbotapi.BotAPI
	logger *logrus.Logger
}

func getUpdateChatId(update tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	} else if update.CallbackQuery != nil {
		return update.CallbackQuery.Message.Chat.ID
	}
	return 0
}

func getUpdateUserFrom(update tgbotapi.Update) *tgbotapi.User {
	if update.Message != nil {
		return update.Message.From
	} else if update.CallbackQuery != nil {
		return update.CallbackQuery.From
	}
	return nil
}

func (vb *VacatoBot) getUpdateLogger(update tgbotapi.Update) *logrus.Entry {
	user := getUpdateUserFrom(update)
	chatId := getUpdateChatId(update)

	if user != nil {
		return vb.logger.WithFields(logrus.Fields{
			"user_id":   user.ID,
			"user_name": user.UserName,
			"chat_id":   chatId,
		})
	}

	return vb.logger.WithField("chat_id", chatId)
}

func (vb *VacatoBot) sendMessage(update tgbotapi.Update, text string) {
	chatId := getUpdateChatId(update)
	logger := vb.getUpdateLogger(update)

	logger.WithField("text", text).Info("Sending message")

	_, err := vb.bot.Send(tgbotapi.NewMessage(chatId, text))
	if err != nil {
		logger.WithError(err).Error("Failed to send message")
	}
}

func (vb *VacatoBot) handleGradient(update tgbotapi.Update) error {
	logger := vb.getUpdateLogger(update)
	text := update.Message.Text

	logger.WithField("text", text).Info("Handling gradient")

	userAvatar, err := GetUserAvatar(vb.bot, update.Message.From.ID)
	if err != nil {
		logger.WithError(err).Error("Failed to get user avatar")
		return err
	}

	gradient := CachedCreateGradient(
		userAvatar.Bounds().Dx(), userAvatar.Bounds().Dy(),
		color.NRGBA{R: 0, G: 0, B: 255, A: 255},
		color.NRGBA{R: 255, G: 0, B: 255, A: 255},
	)

	OverlayImage(userAvatar, gradient, 0.5)
	DrawTextToImage(userAvatar, text)
	DrawSignature(userAvatar)

	var buf bytes.Buffer
	err = imaging.Encode(&buf, userAvatar, imaging.PNG)
	if err != nil {
		logger.WithError(err).Error("Failed to encode image")
		return errors.New("error during overlaying")
	}

	photo := tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FileBytes{
		Name:  "avatar_with_gradient.png",
		Bytes: buf.Bytes(),
	})
	_, err = vb.bot.Send(photo)
	if err != nil {
		logger.WithError(err).Error("Failed to send photo")
	}
	return err
}

func (vb *VacatoBot) handleMenu(update tgbotapi.Update) {
	logger := vb.getUpdateLogger(update)
	logger.Info("Displaying menu")

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Tap the button below and tell me what text you'd like on your avatar.")
	button := tgbotapi.NewInlineKeyboardButtonData("Add text to my avatar", "request_text")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(button),
	)
	msg.ReplyMarkup = keyboard

	_, err := vb.bot.Send(msg)
	if err != nil {
		logger.WithError(err).Error("Failed to send menu message")
	}
}

func (vb *VacatoBot) handleCommand(update tgbotapi.Update) {
	command := update.Message.Command()
	logger := vb.getUpdateLogger(update)
	logger.WithField("command", command).Info("Received command")

	switch command {
	case "start":
		vb.sendMessage(update, "Hey there! Want to add a fun message to your avatar? Use /menu to get started!")
		vb.handleMenu(update)

	case "menu":
		vb.handleMenu(update)

	case "avatar":
		vb.sendMessage(update, requestTextMsg)

	default:
		logger.Errorf("Unknown command %s", command)
		vb.sendMessage(update, "Oops! I don't recognize that command. Try something else!")
	}
}

func (vb *VacatoBot) handleCallback(update tgbotapi.Update) {
	logger := vb.getUpdateLogger(update)
	logger.WithField("callback_data", update.CallbackQuery.Data).Info("Received callback query")

	if update.CallbackQuery.Data == "request_text" {
		vb.sendMessage(update, requestTextMsg)
	}
}

func (vb *VacatoBot) handleReply(update tgbotapi.Update) {
	text := update.Message.Text
	logger := vb.getUpdateLogger(update)
	logger.WithField("text", text).Info("Handling text reply")

	if update.Message.ReplyToMessage.Text == requestTextMsg {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.WithField("error", r).Error("Panic in handleGradient")
				}
			}()

			err := vb.handleGradient(update)
			if err != nil {
				vb.sendMessage(update, "Oh no! Something went wrong. Try again, please!\n\n"+err.Error())
			}
		}()
	}
}

func (vb *VacatoBot) Start() {
	vb.logger.Info("Starting bot")
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := vb.bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil && update.Message.IsCommand() {
			vb.handleCommand(update)
		} else if update.CallbackQuery != nil {
			vb.handleCallback(update)
		} else if update.Message != nil && update.Message.ReplyToMessage != nil {
			vb.handleReply(update)
		}
	}
}

func NewVacatoBot() VacatoBot {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		logrus.Fatal("Failed to retrieve the Telegram token from the environment")
	}

	isDebug := os.Getenv("DEBUG") == "1"

	logger := logrus.New()
	if isDebug {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{})
		logger.SetLevel(logrus.InfoLevel)
	}

	bot, err := GetBot(token, isDebug)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize bot")
	}

	return VacatoBot{bot: bot, logger: logger}
}
