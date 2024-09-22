package main

import (
	"hash/fnv"
	"image"
	"image/color"
	"image/draw"
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

func CreateGradient(width, height int, startColor, endColor color.NRGBA) *image.NRGBA {
	gradientImg := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		alpha := float64(y) / float64(height)
		colorRow := BlendColors(startColor, endColor, alpha)

		for x := 0; x < width; x++ {
			gradientImg.SetNRGBA(x, y, colorRow)
		}
	}

	return gradientImg
}

func getGradientKey(width, height int, startColor, endColor color.NRGBA) uint32 {
	hasher := fnv.New32a()

	buf := []byte{
		byte(width), byte(width >> 8),
		byte(height), byte(height >> 8),
		startColor.R, startColor.G, startColor.B, startColor.A,
		endColor.R, endColor.G, endColor.B, endColor.A,
	}
	hasher.Write(buf)
	return hasher.Sum32()
}

var gradientCache = map[uint32]*image.NRGBA{}

func CachedCreateGradient(width, height int, startColor, endColor color.NRGBA) *image.NRGBA {
	key := getGradientKey(width, height, startColor, endColor)
	if cacheValue, cacheExists := gradientCache[key]; cacheExists {
		return cacheValue
	}

	value := CreateGradient(width, height, startColor, endColor)
	gradientCache[key] = value

	return value
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
	return color.NRGBA{
		R: uint8(float64(colorA.R)*(1-alpha) + float64(colorB.R)*alpha),
		G: uint8(float64(colorA.G)*(1-alpha) + float64(colorB.G)*alpha),
		B: uint8(float64(colorA.B)*(1-alpha) + float64(colorB.B)*alpha),
		A: uint8(float64(colorA.A)*(1-alpha) + float64(colorB.A)*alpha),
	}
}
