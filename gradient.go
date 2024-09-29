package main

import (
	"hash/fnv"
	"image"
	"image/color"
)

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
