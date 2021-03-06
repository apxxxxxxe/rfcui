package color

import (
	"math"

	"github.com/gdamore/tcell/v2"
)

const (
	brightnessLowerLimit = 150
	chromaUpperLimit     = 250
	chromaLowerLimit     = 70
)

var ComfortableColorCode = narrowDownColors(brightnessLowerLimit, chromaUpperLimit, chromaLowerLimit)
var ValidColorCode = narrowDownColors(0, 255, 0)
var ColorCodes = getColorCodes()
var TcellColors = []tcell.Color{
	tcell.ColorBlack,
	tcell.ColorMaroon,
	tcell.ColorGreen,
	tcell.ColorOlive,
	tcell.ColorNavy,
	tcell.ColorPurple,
	tcell.ColorTeal,
	tcell.ColorSilver,
	tcell.ColorGray,
	tcell.ColorRed,
	tcell.ColorLime,
	tcell.ColorYellow,
	tcell.ColorBlue,
	tcell.ColorFuchsia,
	tcell.ColorAqua,
	tcell.ColorWhite,
	tcell.Color16,
	tcell.Color17,
	tcell.Color18,
	tcell.Color19,
	tcell.Color20,
	tcell.Color21,
	tcell.Color22,
	tcell.Color23,
	tcell.Color24,
	tcell.Color25,
	tcell.Color26,
	tcell.Color27,
	tcell.Color28,
	tcell.Color29,
	tcell.Color30,
	tcell.Color31,
	tcell.Color32,
	tcell.Color33,
	tcell.Color34,
	tcell.Color35,
	tcell.Color36,
	tcell.Color37,
	tcell.Color38,
	tcell.Color39,
	tcell.Color40,
	tcell.Color41,
	tcell.Color42,
	tcell.Color43,
	tcell.Color44,
	tcell.Color45,
	tcell.Color46,
	tcell.Color47,
	tcell.Color48,
	tcell.Color49,
	tcell.Color50,
	tcell.Color51,
	tcell.Color52,
	tcell.Color53,
	tcell.Color54,
	tcell.Color55,
	tcell.Color56,
	tcell.Color57,
	tcell.Color58,
	tcell.Color59,
	tcell.Color60,
	tcell.Color61,
	tcell.Color62,
	tcell.Color63,
	tcell.Color64,
	tcell.Color65,
	tcell.Color66,
	tcell.Color67,
	tcell.Color68,
	tcell.Color69,
	tcell.Color70,
	tcell.Color71,
	tcell.Color72,
	tcell.Color73,
	tcell.Color74,
	tcell.Color75,
	tcell.Color76,
	tcell.Color77,
	tcell.Color78,
	tcell.Color79,
	tcell.Color80,
	tcell.Color81,
	tcell.Color82,
	tcell.Color83,
	tcell.Color84,
	tcell.Color85,
	tcell.Color86,
	tcell.Color87,
	tcell.Color88,
	tcell.Color89,
	tcell.Color90,
	tcell.Color91,
	tcell.Color92,
	tcell.Color93,
	tcell.Color94,
	tcell.Color95,
	tcell.Color96,
	tcell.Color97,
	tcell.Color98,
	tcell.Color99,
	tcell.Color100,
	tcell.Color101,
	tcell.Color102,
	tcell.Color103,
	tcell.Color104,
	tcell.Color105,
	tcell.Color106,
	tcell.Color107,
	tcell.Color108,
	tcell.Color109,
	tcell.Color110,
	tcell.Color111,
	tcell.Color112,
	tcell.Color113,
	tcell.Color114,
	tcell.Color115,
	tcell.Color116,
	tcell.Color117,
	tcell.Color118,
	tcell.Color119,
	tcell.Color120,
	tcell.Color121,
	tcell.Color122,
	tcell.Color123,
	tcell.Color124,
	tcell.Color125,
	tcell.Color126,
	tcell.Color127,
	tcell.Color128,
	tcell.Color129,
	tcell.Color130,
	tcell.Color131,
	tcell.Color132,
	tcell.Color133,
	tcell.Color134,
	tcell.Color135,
	tcell.Color136,
	tcell.Color137,
	tcell.Color138,
	tcell.Color139,
	tcell.Color140,
	tcell.Color141,
	tcell.Color142,
	tcell.Color143,
	tcell.Color144,
	tcell.Color145,
	tcell.Color146,
	tcell.Color147,
	tcell.Color148,
	tcell.Color149,
	tcell.Color150,
	tcell.Color151,
	tcell.Color152,
	tcell.Color153,
	tcell.Color154,
	tcell.Color155,
	tcell.Color156,
	tcell.Color157,
	tcell.Color158,
	tcell.Color159,
	tcell.Color160,
	tcell.Color161,
	tcell.Color162,
	tcell.Color163,
	tcell.Color164,
	tcell.Color165,
	tcell.Color166,
	tcell.Color167,
	tcell.Color168,
	tcell.Color169,
	tcell.Color170,
	tcell.Color171,
	tcell.Color172,
	tcell.Color173,
	tcell.Color174,
	tcell.Color175,
	tcell.Color176,
	tcell.Color177,
	tcell.Color178,
	tcell.Color179,
	tcell.Color180,
	tcell.Color181,
	tcell.Color182,
	tcell.Color183,
	tcell.Color184,
	tcell.Color185,
	tcell.Color186,
	tcell.Color187,
	tcell.Color188,
	tcell.Color189,
	tcell.Color190,
	tcell.Color191,
	tcell.Color192,
	tcell.Color193,
	tcell.Color194,
	tcell.Color195,
	tcell.Color196,
	tcell.Color197,
	tcell.Color198,
	tcell.Color199,
	tcell.Color200,
	tcell.Color201,
	tcell.Color202,
	tcell.Color203,
	tcell.Color204,
	tcell.Color205,
	tcell.Color206,
	tcell.Color207,
	tcell.Color208,
	tcell.Color209,
	tcell.Color210,
	tcell.Color211,
	tcell.Color212,
	tcell.Color213,
	tcell.Color214,
	tcell.Color215,
	tcell.Color216,
	tcell.Color217,
	tcell.Color218,
	tcell.Color219,
	tcell.Color220,
	tcell.Color221,
	tcell.Color222,
	tcell.Color223,
	tcell.Color224,
	tcell.Color225,
	tcell.Color226,
	tcell.Color227,
	tcell.Color228,
	tcell.Color229,
	tcell.Color230,
	tcell.Color231,
	tcell.Color232,
	tcell.Color233,
	tcell.Color234,
	tcell.Color235,
	tcell.Color236,
	tcell.Color237,
	tcell.Color238,
	tcell.Color239,
	tcell.Color240,
	tcell.Color241,
	tcell.Color242,
	tcell.Color243,
	tcell.Color244,
	tcell.Color245,
	tcell.Color246,
	tcell.Color247,
	tcell.Color248,
	tcell.Color249,
	tcell.Color250,
	tcell.Color251,
	tcell.Color252,
	tcell.Color253,
	tcell.Color254,
	tcell.Color255,
}

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
