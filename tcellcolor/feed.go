package tcellcolor

import (
	"math"
)

const (
	brightnessLowerLimit = 100
	chromaUpperLimit     = 200
	chromaLowerLimit     = 50
)

var ComfortableColorCode = narrowDownColors(brightnessLowerLimit, chromaUpperLimit, chromaLowerLimit)
var ValidColorCode = narrowDownColors(0, 255, 0)
var ColorCodes = getColorCodes()

func getColorCodes() []int32 {
	colorCodes := []int32{}
	for _, c := range TcellColors {
		colorCodes = append(colorCodes, c.Hex())
	}
	return colorCodes
}

func getBrightness(r, g, b int32) int {
	return int(math.Round(float64(r)*0.299 + float64(g)*0.587 + float64(b)*0.114))
}

func getChroma(r, g, b int32) int {
	max := math.Max(math.Max(float64(r), float64(g)), float64(b))
	min := math.Min(math.Min(float64(r), float64(g)), float64(b))
	return int(math.Round((max - min) * 255 / max))
}

func RGB(v int32) (int32, int32, int32) {
	if v < 0 {
		return -1, -1, -1
	}
	return (v >> 16) & 0xff, (v >> 8) & 0xff, v & 0xff
}

func narrowDownColors(brightnessLowerLimit, chromaUpperLimit, chromaLowerLimit int) []int {

	result := []int{}

	for i, c := range ColorCodes {
		a := getBrightness(RGB(c))
		b := getChroma(RGB(c))
		if a >= brightnessLowerLimit && b >= chromaLowerLimit && b <= chromaUpperLimit {
			result = append(result, i)
		}
	}
	return result
}
