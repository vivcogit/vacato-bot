package main

import (
	"image"
	"testing"
)

func TestDrawTextToImage(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		text   string
	}{
		{
			name:   "Simple text rendering",
			width:  200,
			height: 100,
			text:   "Hello, world!",
		},
		{
			name:   "Multiline text rendering",
			width:  300,
			height: 150,
			text:   "First line\nSecond line",
		},
		{
			name:   "Long text",
			width:  400,
			height: 200,
			text:   "This is a very long text that might need scaling",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := image.NewNRGBA(image.Rect(0, 0, tt.width, tt.height))
			err := DrawTextToImage(img, tt.text)
			if err != nil {
				t.Fatalf("DrawTextToImage failed: %v", err)
			}

			if img.Bounds().Dx() != tt.width || img.Bounds().Dy() != tt.height {
				t.Fatalf("Expected image size %dx%d, got %dx%d",
					tt.width, tt.height, img.Bounds().Dx(), img.Bounds().Dy())
			}
		})
	}
}

func TestDrawTextToImageWithReference(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 400, 200))
	err := DrawTextToImage(img, "Hello World!")
	if err != nil {
		t.Fatalf("DrawTextToImage failed: %v", err)
	}

	referenceImg, err := LoadImage("./test_assets/text_image.png")
	if err != nil {
		t.Fatalf("Failed to load reference image: %v", err)
	}

	if !CompareImages(img, referenceImg) {
		t.Error("Generated image does not match the reference image")
	}
}

func BenchmarkDrawTextToImage(b *testing.B) {
	img := image.NewNRGBA(image.Rect(0, 0, 800, 400))
	text := "This is a test text.\nWe will render this text multiple times for benchmarking."

	for i := 0; i < b.N; i++ {
		err := DrawTextToImage(img, text)
		if err != nil {
			b.Fatalf("DrawTextToImage failed: %v", err)
		}
	}
}
