package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"net/http"
	"os"

	"github.com/disintegration/imaging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func getToken() (string, error) {
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

func getBot(token string, isDebug bool) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bot.Debug = isDebug
	log.Printf("Authorized on account %s", bot.Self.UserName)
	return bot, nil
}

func getUserAvatar(bot *tgbotapi.BotAPI, userId int64) (*image.NRGBA, error) {
	photos, err := bot.GetUserProfilePhotos(tgbotapi.UserProfilePhotosConfig{
		UserID: userId,
		Limit:  1,
	})

	if len(photos.Photos) == 0 {
		return nil, errors.New("you don't have avatars")
	}

	if err != nil {
		return nil, fmt.Errorf("error geting avatar: %s", err)
	}

	largestPhoto := photos.Photos[0][len(photos.Photos[0])-1]
	fileConfig, err := bot.GetFileDirectURL(largestPhoto.FileID)
	if err != nil {
		return nil, fmt.Errorf("error loading avatar: %s", err)
	}

	resp, err := http.Get(fileConfig)
	if err != nil {
		return nil, fmt.Errorf("network error loading avatar: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("network error loading avatar, status: %d", resp.StatusCode)
	}

	img, err := imaging.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error decoding avatar: %s", err)
	}

	return ImageToNRGBA(img), nil
}

func main() {
	token, err := getToken()
	if err != nil {
		panic(err)
	}

	bot, err := getBot(token, true)

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
				userAvatar, err := getUserAvatar(bot, update.Message.From.ID)
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

func ImageToNRGBA(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	imgNRGBA := image.NewNRGBA(bounds)
	draw.Draw(imgNRGBA, bounds, img, bounds.Min, draw.Src)
	return imgNRGBA
}

func CreateGradient(width, height int, startColor, endColor color.NRGBA) *image.NRGBA {
	gradientImg := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		ratio := float64(y) / float64(height)

		r := uint8(float64(startColor.R)*(1-ratio) + float64(endColor.R)*ratio)
		g := uint8(float64(startColor.G)*(1-ratio) + float64(endColor.G)*ratio)
		b := uint8(float64(startColor.B)*(1-ratio) + float64(endColor.B)*ratio)
		a := uint8(float64(startColor.A)*(1-ratio) + float64(endColor.A)*ratio)

		gradColor := color.NRGBA{R: r, G: g, B: b, A: a}

		for x := 0; x < width; x++ {
			gradientImg.Set(x, y, gradColor)
		}
	}

	return gradientImg
}

func OverlayImage(imageA, imageB *image.NRGBA, alpha float64) *image.NRGBA {
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}

	result := image.NewNRGBA(imageA.Bounds())

	draw.Draw(result, imageA.Bounds(), imageA, image.Point{}, draw.Src)

	for y := 0; y < imageA.Bounds().Dy(); y++ {
		for x := 0; x < imageA.Bounds().Dx(); x++ {
			originalPixel := result.NRGBAAt(x, y)
			overlayPixel := imageB.NRGBAAt(x, y)

			r := uint8(float64(originalPixel.R)*(1-alpha) + float64(overlayPixel.R)*alpha)
			g := uint8(float64(originalPixel.G)*(1-alpha) + float64(overlayPixel.G)*alpha)
			b := uint8(float64(originalPixel.B)*(1-alpha) + float64(overlayPixel.B)*alpha)
			a := uint8(float64(originalPixel.A)*(1-alpha) + float64(overlayPixel.A)*alpha)

			result.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: a})
		}
	}

	return result
}
