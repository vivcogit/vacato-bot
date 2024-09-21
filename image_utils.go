package main

import (
	"image"
	"image/color"
	"image/draw"
)

func ImageToNRGBA(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	imgNRGBA := image.NewNRGBA(bounds)
	draw.Draw(imgNRGBA, bounds, img, bounds.Min, draw.Src)
	return imgNRGBA
}

func CreateGradient(width, height int, startColor, endColor color.NRGBA) *image.NRGBA {
	gradientImg := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		alpha := float64(y) / float64(height)

		for x := 0; x < width; x++ {
			gradientImg.Set(x, y, BlendColors(startColor, endColor, alpha))
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

			result.SetNRGBA(x, y, BlendColors(originalPixel, overlayPixel, alpha))
		}
	}

	return result
}

func BlendColors(colorA, colorB color.NRGBA, alpha float64) color.NRGBA {
	r := uint8(float64(colorA.R)*(1-alpha) + float64(colorB.R)*alpha)
	g := uint8(float64(colorA.G)*(1-alpha) + float64(colorB.G)*alpha)
	b := uint8(float64(colorA.B)*(1-alpha) + float64(colorB.B)*alpha)
	a := uint8(float64(colorA.A)*(1-alpha) + float64(colorB.A)*alpha)

	return color.NRGBA{R: r, G: g, B: b, A: a}
}
