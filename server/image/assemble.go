package image

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"sync"

	svg "github.com/ajstarks/svgo"
	"github.com/paw-digital/Pawnimals/server/color"
	"github.com/golang/glog"
	minify "github.com/tdewolff/minify/v2"
	minifysvg "github.com/tdewolff/minify/v2/svg"
)

const DefaultSize = 1080            // Default SVG width/height attribute
const lodBwReplacement = "#9CA2AF" // Replace white with this color on bw assets

type SVG struct {
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
	Doc    string `xml:",innerxml"`
}

func CombineSVG(accessories Accessories) ([]byte, error) {
	var (
		face         SVG
		hair         SVG
		mouth        SVG
		eye          SVG
		backHair     SVG
		bodyOutline  SVG
		hairOutline  SVG
		mouthOutline SVG
		badgeAsset   SVG
	)
	// Parse all SVG assets
	if err := xml.Unmarshal(accessories.FaceAsset.SVGContents, &face); err != nil {
		glog.Fatalf("Unable to parse face SVG %v", err)
		return nil, err
	}
	if accessories.HairAsset != nil {
		if err := xml.Unmarshal(accessories.HairAsset.SVGContents, &hair); err != nil {
			glog.Fatalf("Unable to parse hair SVG %v", err)
			return nil, err
		}
	}
	if accessories.MouthAsset != nil {
		if err := xml.Unmarshal(accessories.MouthAsset.SVGContents, &mouth); err != nil {
			glog.Fatalf("Unable to parse mouth SVG %v", err)
			return nil, err
		}
	}
	if accessories.EyeAsset != nil {
		if err := xml.Unmarshal(accessories.EyeAsset.SVGContents, &eye); err != nil {
			glog.Fatalf("Unable to parse eye SVG %v", err)
			return nil, err
		}
	}
	if accessories.BackHairAsset != nil {
		if err := xml.Unmarshal(accessories.BackHairAsset.SVGContents, &backHair); err != nil {
			glog.Fatalf("Unable to parse back hair SVG %v", err)
			return nil, err
		}
	}
	if accessories.BodyOutlineAsset != nil {
		if err := xml.Unmarshal(accessories.BodyOutlineAsset.SVGContents, &bodyOutline); err != nil {
			glog.Fatalf("Unable to parse body outline SVG %v", err)
			return nil, err
		}
	}
	if accessories.HairOutlineAsset != nil {
		if err := xml.Unmarshal(accessories.HairOutlineAsset.SVGContents, &hairOutline); err != nil {
			glog.Fatalf("Unable to parse hair outline SVG %v", err)
			return nil, err
		}
	}
	if accessories.MouthOutlineAsset != nil {
		if err := xml.Unmarshal(accessories.MouthOutlineAsset.SVGContents, &mouthOutline); err != nil {
			glog.Fatalf("Unable to parse mouth outline SVG %v", err)
			return nil, err
		}
	}
	if accessories.BadgeAsset != nil {
		if err := xml.Unmarshal(accessories.BadgeAsset.SVGContents, &badgeAsset); err != nil {
			glog.Fatalf("Unable to parse badge SVG %v", err)
			return nil, err
		}
	}
	// Perceved brightness of body used for some manipulations
	perceivedBrightness := int(accessories.BodyColor.PerceivedBrightness())
	// Create new SVG writer
	var b bytes.Buffer
	canvas := svg.New(&b)
	canvas.Startraw(fmt.Sprintf("viewBox=\"0 0 %d %d\"", DefaultSize, DefaultSize))
	// Add body outline
	if accessories.BodyOutlineAsset != nil {
		canvas.Gid("bodyOutline")
		bodyOutline.Doc = strings.ReplaceAll(bodyOutline.Doc, "black", accessories.OutlineColor.ToHTML(true))
		io.WriteString(canvas.Writer, bodyOutline.Doc)
		canvas.Gend()
	}
	// Add mouth outline
	if accessories.MouthOutlineAsset != nil {
		canvas.Gid("mouthOutline")
		mouthOutline.Doc = strings.ReplaceAll(mouthOutline.Doc, "black", accessories.OutlineColor.ToHTML(true))
		io.WriteString(canvas.Writer, mouthOutline.Doc)
		canvas.Gend()
	}
	// Add hair outline
	if accessories.HairOutlineAsset != nil {
		canvas.Gid("hairOutline")
		hairOutline.Doc = strings.ReplaceAll(hairOutline.Doc, "black", accessories.OutlineColor.ToHTML(true))
		io.WriteString(canvas.Writer, hairOutline.Doc)
		canvas.Gend()
	}
	// Add back hair
	if accessories.BackHairAsset != nil {
		canvas.Gid("backhair")
		if accessories.HairAsset.HairColored {
			backHair.Doc = strings.ReplaceAll(backHair.Doc, "#FF0000", accessories.HairColor.ToHTML(true))
			backHair.Doc = strings.ReplaceAll(backHair.Doc, "fill-opacity=\"0.15\"", fmt.Sprintf("fill-opacity=\"%f\"", GetTargetOpacity(accessories.HairColor)))
		}
		io.WriteString(canvas.Writer, backHair.Doc)
		canvas.Gend()
	}
	// Face group
	canvas.Gid("face")
	if accessories.FaceAsset.BodyColored {
		//face.Doc = strings.ReplaceAll(face.Doc, "#00FFFF", accessories.BodyColor.ToHTML(true))
		//face.Doc = strings.ReplaceAll(face.Doc, "fill-opacity=\"0.15\"", fmt.Sprintf("fill-opacity=\"%f\"", GetTargetOpacity(accessories.BodyColor)))
		face.Doc = strings.ReplaceAll(face.Doc, "#06c2b5", accessories.BodyColor.ToHTML(true))
		face.Doc = strings.ReplaceAll(face.Doc, "fill-opacity=\"0.15\"", fmt.Sprintf("fill-opacity=\"%f\"", GetTargetOpacity(accessories.BodyColor)))
		face.Doc = strings.ReplaceAll(face.Doc, "#6dfff9", accessories.HairColor.ToHTML(true))
		face.Doc = strings.ReplaceAll(face.Doc, "fill-opacity=\"0.65\"", fmt.Sprintf("fill-opacity=\"%f\"", GetTargetOpacity(accessories.HairColor)))
		//face.Doc = strings.ReplaceAll(face.Doc, "#6dfff9", accessories.BodyColor.ToHTML(true))
		//face.Doc = strings.ReplaceAll(face.Doc, "fill-opacity=\"0.15\"", fmt.Sprintf("fill-opacity=\"%f\"", GetTargetOpacity(accessories.BodyColor)))
		
	}
	io.WriteString(canvas.Writer, face.Doc)
	canvas.Gend()
	// Hair Group
	if accessories.HairAsset != nil {
		canvas.Gid("hair")
		if accessories.HairAsset.HairColored {
			hair.Doc = strings.ReplaceAll(hair.Doc, "#FF0000", accessories.HairColor.ToHTML(true))
			hair.Doc = strings.ReplaceAll(hair.Doc, "fill-opacity=\"0.15\"", fmt.Sprintf("fill-opacity=\"%f\"", GetTargetOpacity(accessories.HairColor)))
		}
		io.WriteString(canvas.Writer, hair.Doc)
		canvas.Gend()
	}
	// Mouth group
	if accessories.MouthAsset != nil {
		canvas.Gid("mouth")
		if accessories.HairAsset.HairColored {
			mouth.Doc = strings.ReplaceAll(mouth.Doc, "#FFFF00", accessories.HairColor.ToHTML(true))
			mouth.Doc = strings.ReplaceAll(mouth.Doc, "fill-opacity=\"0.15\"", fmt.Sprintf("fill-opacity=\"%f\"", GetTargetOpacity(accessories.HairColor)))
		}
		if LightToDarkSwitchPoint > perceivedBrightness && accessories.MouthAsset.DarkBWColored {
			mouth.Doc = strings.ReplaceAll(mouth.Doc, "white", lodBwReplacement)
		}
		if LightToDarkSwitchPoint > perceivedBrightness && accessories.MouthAsset.DarkColored {
			mouth.Doc = strings.ReplaceAll(mouth.Doc, "black", "white")
		} else if perceivedBrightness >= LightToDarkSwitchPoint && accessories.MouthAsset.BLK299 {
			mouth.Doc = strings.ReplaceAll(mouth.Doc, "fill-opacity=\"0.299\"", fmt.Sprintf("fill-opacity=\"%f\"", GetBlk299Opacity(accessories.BodyColor)))
		}
		io.WriteString(canvas.Writer, mouth.Doc)
		canvas.Gend()
	}
	// Eye group
	if accessories.EyeAsset != nil {
		canvas.Gid("eye")
		if LightToDarkSwitchPoint > perceivedBrightness && accessories.EyeAsset.DarkBWColored {
			eye.Doc = strings.ReplaceAll(eye.Doc, "white", lodBwReplacement)
		}
		if LightToDarkSwitchPoint > perceivedBrightness && accessories.EyeAsset.DarkColored {
			eye.Doc = strings.ReplaceAll(eye.Doc, "black", "white")
		} else if perceivedBrightness >= LightToDarkSwitchPoint && accessories.EyeAsset.BLK299 {
			eye.Doc = strings.ReplaceAll(eye.Doc, "fill-opacity=\"0.299\"", fmt.Sprintf("fill-opacity=\"%f\"", GetBlk299Opacity(accessories.BodyColor)))
		}
		io.WriteString(canvas.Writer, eye.Doc)
		canvas.Gend()
	}
	// Badge group
	if accessories.BadgeAsset != nil {
		canvas.Gid("badge")
		// Change color based on outline
		if accessories.BodyOutlineAsset != nil {
			badgeAsset.Doc = strings.ReplaceAll(badgeAsset.Doc, "white", accessories.OutlineColor.ToHTML(true))
		}
		io.WriteString(canvas.Writer, badgeAsset.Doc)
		canvas.Gend()
	}
	// End document
	canvas.End()

	// Minify
	var ret []byte
	ret, _ = getMinifier().minifier.Bytes("image/svg+xml", b.Bytes())

	return ret, nil
}

func GetTargetOpacity(color color.RGB) float64 {
	return MinShadowOpacity + (1-color.PerceivedBrightness()/100)*(MaxShadowOpacity-MinShadowOpacity)
}

func GetBlk299Opacity(color color.RGB) float64 {
	return MinBlk29AccessoryOpacity + (1-color.PerceivedBrightness()/100)*(MaxBlk29AccessoryOpacity-MinBlk29AccessoryOpacity)
}

// Singleton to get minifier
type minifySingleton struct {
	minifier *minify.M
}

var mSingleton *minifySingleton
var onceM sync.Once

func getMinifier() *minifySingleton {
	onceM.Do(func() {
		minifier := minify.New()
		minifier.AddFunc("image/svg+xml", minifysvg.Minify)
		mSingleton = &minifySingleton{
			minifier: minifier,
		}
	})
	return mSingleton
}
