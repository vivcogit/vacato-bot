package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

func ImageToNRGBA(img image.Image) *image.NRGBA {
	if nrgba, ok := img.(*image.NRGBA); ok {
		return nrgba
	}

	bounds := img.Bounds()
	imgNRGBA := image.NewNRGBA(bounds)
	draw.Draw(imgNRGBA, bounds, img, bounds.Min, draw.Src)

	return imgNRGBA
}

func OverlayImage(imageA, imageB *image.NRGBA, alpha float64) {
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}

	for y := 0; y < imageA.Bounds().Dy(); y++ {
		for x := 0; x < imageA.Bounds().Dx(); x++ {
			originalPixel := imageA.NRGBAAt(x, y)
			overlayPixel := imageB.NRGBAAt(x, y)

			imageA.SetNRGBA(x, y, BlendColors(originalPixel, overlayPixel, alpha))
		}
	}
}

func BlendColors(colorA, colorB color.NRGBA, alpha float64) color.NRGBA {
	return color.NRGBA{
		R: uint8(float64(colorA.R)*(1-alpha) + float64(colorB.R)*alpha),
		G: uint8(float64(colorA.G)*(1-alpha) + float64(colorB.G)*alpha),
		B: uint8(float64(colorA.B)*(1-alpha) + float64(colorB.B)*alpha),
		A: uint8(float64(colorA.A)*(1-alpha) + float64(colorB.A)*alpha),
	}
}

func LoadImage(path string) (*image.NRGBA, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening image: %v", err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %v", err)
	}

	bounds := img.Bounds()
	nrgbaImg := image.NewNRGBA(bounds)
	draw.Draw(nrgbaImg, bounds, img, bounds.Min, draw.Src)

	return nrgbaImg, nil
}

func CompareImages(img1, img2 *image.NRGBA) bool {
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	if !bounds1.Eq(bounds2) {
		return false
	}

	for y := bounds1.Min.Y; y < bounds1.Max.Y; y++ {
		for x := bounds1.Min.X; x < bounds1.Max.X; x++ {
			idx := img1.PixOffset(x, y)
			r1, g1, b1, a1 := img1.Pix[idx], img1.Pix[idx+1], img1.Pix[idx+2], img1.Pix[idx+3]
			r2, g2, b2, a2 := img2.Pix[idx], img2.Pix[idx+1], img2.Pix[idx+2], img2.Pix[idx+3]

			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				return false
			}
		}
	}
	return true
}
