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

	"github.com/h2non/bimg"
	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"

	"wv2/types"
	"wv2/utils"
)

var (
	OptionsE *encoder.Options
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

func addLabel(img *image.RGBA, size float64, x, y int, label []string) (ptX, ptY int) {
	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(fontD)
	c.SetFontSize(size)
	c.SetHinting(font.HintingNone)

	// Set source (https://github.com/golang/freetype/blob/master/example/freetype/main.go)

	c.SetClip(img.Bounds())
	c.SetSrc(image.White)
	c.SetDst(img)

	// Draw the text.
	lastLineLen := 0

	pt := freetype.Pt(x, y+int(c.PointToFixed(size)>>6))
	for _, s := range label {
		_, err := c.DrawString(s, pt)
		if err != nil {
			fmt.Println(err)
			return 0, 0
		}
		pt.Y += c.PointToFixed(size * spacing)
		lastLineLen = len(s) + 1
	}

	return lastLineLen, pt.Y.Ceil()
}

func resizeImage(img image.Image, factor int) draw.Image {
	// Read the image from the file
	dst := image.NewRGBA(image.Rect(0, 0, img.Bounds().Max.X/factor, img.Bounds().Max.Y/factor))

	draw.ApproxBiLinear.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)

	return dst
}

func scaleImage(imgBuf []byte, width, height int) (image.Image, error) {
	// Read the image from the bytes
	newImage := bimg.NewImage(imgBuf)
	resized, err := newImage.ResizeAndCrop(width, height)

	if err != nil {
		return nil, err
	}

	// Re-encode to webp
	buf := bytes.NewBuffer(resized)
	webpBuf, err := webp.Decode(buf, &decoder.Options{})

	// Convert to draw.Image

	return webpBuf, nil
}

func copyImage(x, y int, img image.Image) {
	dp := image.Point{X: x, Y: y}

	// Carve out rectangle for the image
	draw.Draw(mainImg, image.Rectangle{Min: dp, Max: dp.Add(img.Bounds().Size())}, img, image.Point{}, draw.Src)
}

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

// Returns a scale factor for the given text
func getCenterScale(s string) int {
	if len(s) < 6 {
		return 0
	} else if len(s) >= 6 && len(s) <= 8 {
		return 1
	} else {
		return 2
	}
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

	OptionsE, err = encoder.NewLosslessEncoderOptions(encoder.PresetDefault, 1)

	if err != nil {
		panic(err)
	}

	// Read listicon.webp from assets
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

	listicon = resizeImage(listiconUnparsed, imgScaleFactor)
}

func DrawWidget(bot types.WidgetUser) image.Image {
	// Draw a 640x480 black rectangle first
	draw.Draw(mainImg, mainImg.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	// textIndent is the amount of space to leave on the left of the screen. Negative because positive means reverse direction.
	copyImage(listicon.Bounds().Min.X+textIndent, listicon.Bounds().Min.Y+textIndent+extraTopIndent, listicon)

	addLabel(mainImg.(*image.RGBA), titleSize, textIndent, listicon.Bounds().Dy()+textIndent+extraTopIndent, []string{"Fates List"})

	// Resive avatar
	var avatarImg image.Image
	avatarImg = resizeImage(bot.AvatarBytes, 1)

	fmt.Println(bot.AvatarBytes.Bounds().Dx())

	// Convert draw.Image to RGBA
	fmt.Println("Trying to resize")
	avatarImgD := avatarImg.(*image.RGBA)

	// Resize avatar to 512x512
	var w = bytes.NewBuffer([]byte{})
	err := webp.Encode(w, avatarImgD, OptionsE)

	if err == nil {
		avatarImg, err = scaleImage(w.Bytes(), 128, 128)

		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println(err)
	}

	/* Now insert the avatar image into the main image.
	To get the point we insert at, we first find center of main image and subtract X of that from X of avatar image */
	copyImage(centeredImage(avatarImg), centeredImageY(avatarImg), utils.Circle(avatarImg))

	// centeredImageY(avatarImg)+(getImageCenter(avatarImg).Y*2) means we add the center of the avatar image * 2 (to get diameter) to the center of the main image
	addLabel(mainImg.(*image.RGBA), titleSize, centeredImage(avatarImg), centeredImageY(avatarImg)+(getImageCenter(avatarImg).Y*2), []string{bot.Username})

	return mainImg
}
