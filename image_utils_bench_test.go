package main

import (
	"image/color"
	"testing"
)

func BenchmarkCachedCreateGradient(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CachedCreateGradient(800, 600, color.NRGBA{R: 0, G: 0, B: 255, A: 255}, color.NRGBA{R: 255, G: 0, B: 255, A: 255})
	}
}

func BenchmarkCreateGradient(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CreateGradient(800, 600, color.NRGBA{R: 0, G: 0, B: 255, A: 255}, color.NRGBA{R: 255, G: 0, B: 255, A: 255})
	}
}
