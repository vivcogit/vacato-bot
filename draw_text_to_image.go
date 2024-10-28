package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

var fontCache = map[string]*FontCache{}

type FontCache struct {
	font          *sfnt.Font
	defaultFace   font.Face
	signatureFace font.Face
}

func CachedLoadFont(fontPath string) (*FontCache, error) {
	if cacheValue, cacheExists := fontCache[fontPath]; cacheExists {
		return cacheValue, nil
	}

	ttf, err := LoadFont(fontPath)

	if err != nil {
		return nil, err
	}

	defaultFace, err := opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    48,
		DPI:     72,
		Hinting: font.HintingNone,
	})

	if err != nil {
		return nil, err
	}

	signatureFace, _ := opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    16,
		DPI:     72,
		Hinting: font.HintingNone,
	})

	fontCache[fontPath] = &FontCache{
		font:          ttf,
		defaultFace:   defaultFace,
		signatureFace: signatureFace,
	}

	return fontCache[fontPath], nil
}

func LoadFont(fontPath string) (*sfnt.Font, error) {
	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("error reading font: %v", err)
	}

	ttfFont, err := opentype.Parse(fontBytes)

	if err != nil {
		return nil, fmt.Errorf("error parsing font: %v", err)
	}

	return ttfFont, nil
}

func DrawTextToImage(img *image.NRGBA, text string) error {
	bounds := img.Bounds()

	ttfFont, err := CachedLoadFont("./assets/Roboto-Regular.ttf")
	if err != nil {
		return err
	}

	const fontSize = 48
	const horizontalPaddingPercent = 10.0
	const verticalPaddingPercent = 10.0

	horizontalPadding := float64(bounds.Dx()) * horizontalPaddingPercent / 100.0
	verticalPadding := float64(bounds.Dy()) * verticalPaddingPercent / 100.0

	lines := strings.Split(text, "\n")
	if len(lines) > 2 {
		lines = lines[:2]
	}

	textWidth, textHeight := measureMultilineTextSize(ttfFont.defaultFace, lines)
	scaleFactor := calculateScaleFactor(textWidth, textHeight, float64(bounds.Dx()), float64(bounds.Dy()), horizontalPadding, verticalPadding)

	scaledFace, err := opentype.NewFace(ttfFont.font, &opentype.FaceOptions{
		Size:    fontSize * scaleFactor,
		DPI:     72,
		Hinting: font.HintingNone,
	})

	if err != nil {
		return fmt.Errorf("error creating scaled font face: %v", err)
	}

	scaledDrawer := &font.Drawer{
		Face: scaledFace,
	}
	defer scaledFace.Close()

	textHeight = measureMultilineTextHeight(scaledFace, len(lines))

	startY := verticalPadding + ((float64(bounds.Dy()) - 2*verticalPadding - textHeight) / 2) + float64(scaledFace.Metrics().Ascent.Ceil())

	for i, line := range lines {
		lineWidth := measureLineWidth(scaledDrawer, line)
		lineStartX := horizontalPadding + (float64(bounds.Dx())-2*horizontalPadding-lineWidth)/2
		y := startY + float64(i)*float64(scaledFace.Metrics().Height.Ceil())
		drawer := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(color.White),
			Face: scaledFace,
			Dot:  fixed.Point26_6{X: fixed.Int26_6(lineStartX * 64), Y: fixed.Int26_6(y * 64)},
		}
		drawer.DrawString(line)
	}

	return nil
}

const signature = "@VacatoBot"

func DrawSignature(img *image.NRGBA) error {
	ttfFont, err := CachedLoadFont("./assets/Roboto-Regular.ttf")
	if err != nil {
		return err
	}

	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.White),
		Face: ttfFont.signatureFace,
		Dot: fixed.Point26_6{
			X: fixed.Int26_6(img.Bounds().Dx()*64 - 6000),
			Y: fixed.Int26_6(img.Bounds().Dy()*64 - 500),
		},
	}
	drawer.DrawString(signature)

	return nil
}

func measureMultilineTextSize(face font.Face, lines []string) (float64, float64) {
	maxWidth := 0.0
	drawer := &font.Drawer{
		Face: face,
	}

	for _, line := range lines {
		lineWidth := measureLineWidth(drawer, line)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}

	return maxWidth, measureMultilineTextHeight(face, len(lines))
}

func measureMultilineTextHeight(face font.Face, linesCount int) float64 {
	return float64(face.Metrics().Height.Ceil() * linesCount)
}

func measureLineWidth(drawer *font.Drawer, text string) float64 {
	width := drawer.MeasureString(text)
	return float64(width) / 64
}

func calculateScaleFactor(textWidth, textHeight, imageWidth, imageHeight, horizontalPadding, verticalPadding float64) float64 {
	maxWidth := imageWidth - 2*horizontalPadding
	maxHeight := imageHeight - 2*verticalPadding

	scaleWidth := maxWidth / textWidth
	scaleHeight := maxHeight / textHeight

	return math.Min(scaleWidth, scaleHeight)
}
