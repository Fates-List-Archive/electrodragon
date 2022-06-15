// To be rewritten in https://github.com/h2non/bimg once design is finalized

package widgets

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"os"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/draw"

	"wv2/imgtools"
	"wv2/types"
)

var (
	listicon draw.Image
	mainImg  draw.Image
	fontD    *truetype.Font
)

const (
	imgScaleFactor    = 5    // Scale factor for listicon
	avatarScaleFactor = 10   // Scale factor for avatar
	growImageFactor   = 2    // Factor to grow image
	titleSize         = 25   // Size of title
	textIndent        = 10   // How much to indent text to the left of the screen
	extraTopIndent    = 8    // How much to indent text to the top of the screen (extra indent)
	dpi               = 72   // DPI for freetype (screen resolution in Dots Per Inch)
	spacing           = 1.25 // Line spacing (e.g. size*spacing = line height)
)

func getImageCenter(img image.Image) image.Point {
	return image.Point{X: img.Bounds().Dx() / 2, Y: img.Bounds().Dy() / 2}
}

// Returns the X coordinate at which the image would be cenetred
func centeredImage(img image.Image) int {
	return getImageCenter(mainImg).X - getImageCenter(img).X
}

// Returns the X coordinate at which the image would be cenetred
func centeredImageY(img image.Image) int {
	return getImageCenter(mainImg).Y - getImageCenter(img).Y
}

func init() {
	// Load the font first
	fontBytes, err := ioutil.ReadFile("assets/font.ttf")
	if err != nil {
		fmt.Println(err)
		return
	}
	fontD, err = freetype.ParseFont(fontBytes)
	if err != nil {
		fmt.Println(err)
		return
	}

	mainImg = image.NewRGBA(image.Rect(0, 0, 640, 480))

	if err != nil {
		panic(err)
	}

	f, err := os.Open("assets/listicon.png")

	if err != nil {
		panic(err)
	}

	defer f.Close()

	// Read the image from the file
	var listiconUnparsed image.Image
	listiconUnparsed, err = png.Decode(f)

	if err != nil {
		panic(err)
	}

	// Resize avatar to 512x512
	var w = bytes.NewBuffer([]byte{})
	err = png.Encode(w, listiconUnparsed)

	if err == nil {
		listiconDraw, err := imgtools.ScaleImage(w.Bytes(), 24, 24)

		if err != nil {
			panic(err)
		}

		listicon = imgtools.ResizeImage(listiconDraw, 1)
	} else {
		panic(err)
	}
}

func DrawWidget(bot types.WidgetUser) image.Image {
	// Draw a 640x480 black rectangle first
	fmt.Println("Starting draw")

	draw.Draw(mainImg, mainImg.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	// textIndent is the amount of space to leave on the left of the screen. Negative because positive means reverse direction.
	imgtools.CopyImage(textIndent, mainImg.Bounds().Dy()-listicon.Bounds().Dy()-textIndent-extraTopIndent, listicon, mainImg)

	imgtools.AddLabel(mainImg.(*image.RGBA), types.Label{
		Size: titleSize,
		X:    textIndent + listicon.Bounds().Dx() + textIndent,
		// The Y coordinate is calculated as the main image height minus the height of the avatar minus text indents minus the amount of space to center it with the avatar (1/8 of avatar height).
		Y:        mainImg.Bounds().Dy() - listicon.Bounds().Dy() - textIndent - extraTopIndent - listicon.Bounds().Dy()/8,
		Labels:   []string{"Fates List"},
		FontData: fontD,
		DPI:      dpi,
		Spacing:  spacing,
	})

	var avatarImg image.Image
	avatarImg = imgtools.ResizeImage(bot.AvatarBytes, 1)

	fmt.Println(bot.AvatarBytes.Bounds().Dx())

	// Convert draw.Image to RGBA
	fmt.Println("Trying to resize")
	avatarImgD := avatarImg.(*image.RGBA)

	// Resize avatar to 512x512
	var w = bytes.NewBuffer([]byte{})
	err := png.Encode(w, avatarImgD)

	if err == nil {
		avatarImg, err = imgtools.ScaleImage(w.Bytes(), 128, 128)

		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println(err)
	}

	/* Now insert the avatar image into the main image.
	To get the point we insert at, we first find center of main image and subtract X of that from X of avatar image */
	imgtools.CopyImage(centeredImage(avatarImg), centeredImageY(avatarImg), imgtools.Circle(avatarImg), mainImg)

	// centeredImageY(avatarImg)+(getImageCenter(avatarImg).Y*2) means we add the center of the avatar image * 2 (to get diameter) to the center of the main image
	imgtools.AddLabel(mainImg.(*image.RGBA), types.Label{
		Size:     titleSize,
		X:        centeredImage(avatarImg),
		Y:        centeredImageY(avatarImg) + (getImageCenter(avatarImg).Y * 2),
		Labels:   []string{bot.Username},
		FontData: fontD,
		DPI:      dpi,
		Spacing:  spacing,
	})

	fmt.Println("Ending draw")

	return mainImg
}
