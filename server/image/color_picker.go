package image

import (
	"math"
	"strconv"

	"github.com/appditto/natricon/server/color"
	"github.com/appditto/natricon/server/rand"
)

// Min and max perceivedBrightness values (between 0 and 100)
var MinPerceivedBrightness = 18.0
var MaxPerceivedBrightness = 95.0

// Min and max perceivedBrightness values (between 0 and 255)
var MinPerceivedBrightness255 = MinPerceivedBrightness / 100 * 255
var MaxPerceivedBrightness255 = MaxPerceivedBrightness / 100 * 255

// Variable for body and hair hue distance
var BodyAndHairHueDistance = 90.0

// Min total saturation (bodySaturation + hairSaturation shouldn't be below this value)
var MinTotalSaturation = 60.0

// Min total brightness
var MinTotalBrightness = 130.0

// Min hair brightness
var MinHairBrightness = 40.0

// Min and max shadow opacity
var MinShadowOpacity = 0.075
var MaxShadowOpacity = 0.4

// Min and max for _blk29 tagged accessory opacity
var MinBlk29AccessoryOpacity = 0.2
var MaxBlk29AccessoryOpacity = 0.5

// Light-Dark switch for Natricon body (depends on perceived brightness of 0-100)
var LightToDarkSwitchPoint = 30

// Limits that will be used on hairBrightness depending on the hairSaturation
var hairBrightnessDynamicMax = 90.0
var hairSaturationDynamicMin = 10.0

// GetBodyColor - Get body color with given entropy
func GetBodyColor(entropy string) (color.RGB, error) {
	// Want to generate hue between 0-360
	// Get detemrinistic RNG
	randSeed, err := strconv.ParseInt(entropy[0:4], 16, 64)
	if err != nil {
		return color.RGB{}, err
	}
	outRGB := color.RGB{}
	// Generate R between 0..255
	r := rand.Init()
	r.Seed(uint32(randSeed))
	outRGB.R = float64(r.Int31n(255*10000)) / 10000
	// Generate G between 0.255
	randSeed, err = strconv.ParseInt(entropy[4:8], 16, 64)
	if err != nil {
		return color.RGB{}, err
	}
	r = rand.Init()
	r.Seed(uint32(randSeed))
	outRGB.G = float64(r.Int31n(255*10000)) / 10000
	// Generate Blue
	randSeed, err = strconv.ParseInt(entropy[8:12], 16, 64)
	if err != nil {
		return color.RGB{}, err
	}
	r = rand.Init()
	r.Seed(uint32(randSeed))
	lowerBound := math.Max(
		math.Sqrt(
			math.Max(
				(MinPerceivedBrightness255*MinPerceivedBrightness255-color.RedPBMultiplier*outRGB.R*outRGB.R-color.GreenPBMultiplier*outRGB.G*outRGB.G)/color.GreenPBMultiplier,
				0.0,
			),
		),
		0.0,
	) * 10000
	upperBound := math.Min(
		math.Sqrt(
			math.Max(
				(MaxPerceivedBrightness255*MaxPerceivedBrightness255-color.RedPBMultiplier*outRGB.R*outRGB.R-color.GreenPBMultiplier*outRGB.G*outRGB.G)/color.GreenPBMultiplier,
				0.0,
			),
		),
		255.0,
	) * 10000
	outRGB.B = (float64(r.Int31n(int32(upperBound)-int32(lowerBound))) + lowerBound) / 10000

	return outRGB, nil
}

// GetHairColor - Get a complementary color with given entropy
func GetHairColor(bodyColor color.RGB, hEntropy string, sEntropy string, bEntropy string) (color.RGB, error) {
	var err error
	// Get as HSB color
	bodyColorHSB := bodyColor.ToHSB()

	var randSeed int64
	// Want to shift the hue between 90-270
	// Get detemrinistic RNG
	randSeed, err = strconv.ParseInt(hEntropy, 16, 64)
	if err != nil {
		return color.RGB{}, err
	}

	// Generate random shift between <minDistance>...270
	r := rand.Init()
	r.Seed(uint32(randSeed))
	lowerBound := bodyColorHSB.H - 180 - BodyAndHairHueDistance
	upperBound := bodyColorHSB.H - 180 + BodyAndHairHueDistance
	H := (float64(r.Int31n(int32(upperBound*10000)-int32(lowerBound*10000))) + lowerBound*10000) / 10000

	// If < 0 normalize
	if H < 0 {
		H += 360
	}

	// Generate saturation
	randSeed, err = strconv.ParseInt(sEntropy, 16, 64)
	if err != nil {
		return color.RGB{}, err
	}
	r = rand.Init()
	r.Seed(uint32(randSeed))
	// When body saturation is high enough, hair saturation can end up being less than 0 here, so we're making sure that hair saturation's minimum value never goes below 0v
	lowerSBound := int32(math.Max(MinTotalSaturation-bodyColorHSB.S*100.0, 0) * 10000)
	S := float64(r.Int31n(100*10000-lowerSBound)+lowerSBound) / (100.0 * 10000.0)

	// Generate random brightess between MinimumBrightness - 100
	randSeed, err = strconv.ParseInt(bEntropy, 16, 64)
	if err != nil {
		return color.RGB{}, err
	}
	r = rand.Init()
	r.Seed(uint32(randSeed))
	// When the perceived brightness of body is low enough, hair brightness can end up being more than 100 here, so we're making sure that hair brightness's minimum value never goes above 100
	upperBBound := hairBrightnessDynamicMax
	if S*100 > hairSaturationDynamicMin {
		upperBBound = 100
	}
	lowerBBound := math.Min(math.Max(MinTotalBrightness-bodyColorHSB.B*100.0, MinHairBrightness), upperBBound)
	// Allow more precision for RNG
	upperBBound *= 10000
	lowerBBound *= 10000
	B := float64(r.Int31n(int32(upperBBound)-int32(lowerBBound))+int32(lowerBBound)) / (100 * 10000)
	return color.HSB{
		H: H,
		S: S,
		B: B,
	}.ToRGB(), nil
}
