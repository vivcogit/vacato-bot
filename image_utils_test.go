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
				startColor: color.NRGBA{R: 255, G: 0, B: 0, A: 255}, // Красный
				endColor:   color.NRGBA{R: 0, G: 0, B: 255, A: 255}, // Синий
			},
			params2: GradientParams{
				width:      100,
				height:     200,
				startColor: color.NRGBA{R: 0, G: 255, B: 0, A: 255}, // Зеленый
				endColor:   color.NRGBA{R: 0, G: 0, B: 255, A: 255}, // Синий
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
				startColor: color.NRGBA{R: 255, G: 255, B: 0, A: 255}, // Желтый
				endColor:   color.NRGBA{R: 0, G: 255, B: 255, A: 255}, // Бирюзовый
			},
			params2: GradientParams{
				width:      100,
				height:     200,
				startColor: color.NRGBA{R: 255, G: 255, B: 0, A: 255}, // Желтый
				endColor:   color.NRGBA{R: 255, G: 0, B: 255, A: 255}, // Пурпурный
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
