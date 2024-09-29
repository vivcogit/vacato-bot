package main

import (
	"image/color"
	"testing"
)

func TestGetGradientKey(t *testing.T) {
	type GradientParams struct {
		width      int
		height     int
		startColor color.NRGBA
		endColor   color.NRGBA
	}

	tests := []struct {
		name       string
		params1    GradientParams
		params2    GradientParams
		expectSame bool
	}{
		{
			name: "Identical inputs should return same key",
			params1: GradientParams{
				width:      100,
				height:     200,
				startColor: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
				endColor:   color.NRGBA{R: 0, G: 0, B: 255, A: 255},
			},
			params2: GradientParams{
				width:      100,
				height:     200,
				startColor: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
				endColor:   color.NRGBA{R: 0, G: 0, B: 255, A: 255},
			},
			expectSame: true,
		},
		{
			name: "Different colors should return different keys",
			params1: GradientParams{
				width:      100,
				height:     200,
				startColor: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
				endColor:   color.NRGBA{R: 0, G: 0, B: 255, A: 255},
			},
			params2: GradientParams{
				width:      100,
				height:     200,
				startColor: color.NRGBA{R: 0, G: 255, B: 0, A: 255},
				endColor:   color.NRGBA{R: 0, G: 0, B: 255, A: 255},
			},
			expectSame: false,
		},
		{
			name: "Different sizes should return different keys",
			params1: GradientParams{
				width:      100,
				height:     200,
				startColor: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
				endColor:   color.NRGBA{R: 0, G: 0, B: 255, A: 255},
			},
			params2: GradientParams{
				width:      101, // Изменили ширину
				height:     200,
				startColor: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
				endColor:   color.NRGBA{R: 0, G: 0, B: 255, A: 255},
			},
			expectSame: false,
		},
		{
			name: "Zero dimensions with same colors should return same key",
			params1: GradientParams{
				width:      0,
				height:     0,
				startColor: color.NRGBA{R: 0, G: 0, B: 0, A: 0},
				endColor:   color.NRGBA{R: 255, G: 255, B: 255, A: 255},
			},
			params2: GradientParams{
				width:      0,
				height:     0,
				startColor: color.NRGBA{R: 0, G: 0, B: 0, A: 0},
				endColor:   color.NRGBA{R: 255, G: 255, B: 255, A: 255},
			},
			expectSame: true,
		},
		{
			name: "Different end colors should return different keys",
			params1: GradientParams{
				width:      100,
				height:     200,
				startColor: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
				endColor:   color.NRGBA{R: 0, G: 255, B: 255, A: 255},
			},
			params2: GradientParams{
				width:      100,
				height:     200,
				startColor: color.NRGBA{R: 255, G: 255, B: 0, A: 255},
				endColor:   color.NRGBA{R: 255, G: 0, B: 255, A: 255},
			},
			expectSame: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key1 := getGradientKey(tt.params1.width, tt.params1.height, tt.params1.startColor, tt.params1.endColor)
			key2 := getGradientKey(tt.params2.width, tt.params2.height, tt.params2.startColor, tt.params2.endColor)

			if (key1 == key2) != tt.expectSame {
				t.Errorf("Test %q failed: expected keys to be equal: %v, got key1=%d, key2=%d", tt.name, tt.expectSame, key1, key2)
			}
		})
	}
}

func TestCreateGradient(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		height     int
		startColor color.NRGBA
		endColor   color.NRGBA
	}{
		{
			name:       "Black to White Gradient",
			width:      10,
			height:     10,
			startColor: color.NRGBA{0, 0, 0, 255},
			endColor:   color.NRGBA{255, 255, 255, 255},
		},
		{
			name:       "Red to Blue Gradient",
			width:      20,
			height:     20,
			startColor: color.NRGBA{255, 0, 0, 255},
			endColor:   color.NRGBA{0, 0, 255, 255},
		},
		{
			name:       "Solid Color Gradient",
			width:      5,
			height:     5,
			startColor: color.NRGBA{100, 150, 200, 255},
			endColor:   color.NRGBA{100, 150, 200, 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gradientImg := CreateGradient(tt.width, tt.height, tt.startColor, tt.endColor)

			if gradientImg.Bounds().Dx() != tt.width || gradientImg.Bounds().Dy() != tt.height {
				t.Fatalf("Expected image size %dx%d, got %dx%d",
					tt.width, tt.height, gradientImg.Bounds().Dx(), gradientImg.Bounds().Dy())
			}

			testRows := []int{0, tt.height / 2, tt.height - 1}
			x := tt.width / 2
			for _, y := range testRows {
				alpha := float64(y) / float64(tt.height)
				expectedColor := BlendColors(tt.startColor, tt.endColor, alpha)

				actualColor := gradientImg.NRGBAAt(x, y)
				if actualColor != expectedColor {
					t.Errorf("At position (%d,%d), expected color %v, got %v",
						x, y, expectedColor, actualColor)
				}
			}
		})
	}
}

func TestCreateGradientWithReference(t *testing.T) {
	generatedImg := CreateGradient(400, 600, color.NRGBA{R: 0, G: 0, B: 255, A: 255},
		color.NRGBA{R: 255, G: 0, B: 255, A: 255})

	referenceImg, err := LoadImage("./test_assets/gradient.png")
	if err != nil {
		t.Fatalf("Failed to load reference image: %v", err)
	}

	if !CompareImages(generatedImg, referenceImg) {
		t.Error("Generated image does not match the reference image")
	}
}
