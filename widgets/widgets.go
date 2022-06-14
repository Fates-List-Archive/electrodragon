package widgets

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"os"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"

	"wv2/types"
)

var (
	optionsE *encoder.Options
	listicon draw.Image
	mainImg  draw.Image
	fontD    *truetype.Font
)

const (
	imgScaleFactor = 5    // Scale factor for listicon
	titleSize      = 25   // Size of title
	textIndent     = 10   // How much to indent text to the left of the screen
	extraTopIndent = 8    // How much to indent text to the top of the screen (extra indent)
	dpi            = 72   // DPI for freetype (screen resolution in Dots Per Inch)
	spacing        = 1.25 // Line spacing (e.g. size*spacing = line height)
)

func addLabel(img *image.RGBA, size float64, x, y int, label []string) (ptY int) {
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
	pt := freetype.Pt(x, y+int(c.PointToFixed(size)>>6))
	for _, s := range label {
		_, err := c.DrawString(s, pt)
		if err != nil {
			fmt.Println(err)
			return 0
		}
		pt.Y += c.PointToFixed(size * spacing)
	}

	return pt.Y.Ceil()
}

func resizeImage(img image.Image) draw.Image {
	// Read the image from the file
	dst := image.NewRGBA(image.Rect(0, 0, img.Bounds().Max.X/imgScaleFactor, img.Bounds().Max.Y/imgScaleFactor))

	draw.ApproxBiLinear.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)

	return dst
}

func copyImage(x, y int, img draw.Image) {
	dp := image.Point{X: x, Y: y}

	// Carve out rectangle for the image
	draw.Draw(mainImg, image.Rectangle{Min: dp, Max: dp.Add(img.Bounds().Size())}, img, image.ZP, draw.Src)
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

	optionsE, err = encoder.NewLosslessEncoderOptions(encoder.PresetDefault, 1)

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

	listicon = resizeImage(listiconUnparsed)
}

func DrawWidget(bot types.WidgetUser) {
	// Draw a 640x480 black rectangle first
	draw.Draw(mainImg, mainImg.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)

	// - textIndent is the amount of space to leave on the left of the screen. Negative because positive means reverse direction.
	copyImage(listicon.Bounds().Min.X+textIndent, listicon.Bounds().Min.Y+textIndent+extraTopIndent, listicon)

	labelY := addLabel(mainImg.(*image.RGBA), titleSize, textIndent, listicon.Bounds().Dy()+textIndent+extraTopIndent, []string{"Fates List"})

	// Resive avatar
	avatarImg := resizeImage(bot.AvatarBytes)

	// Now insert the avatar image into the main image
	copyImage(0, labelY, avatarImg)

	f, _ := os.Create(bot.OutFile)

	defer f.Close()

	// Write the image to the file

	if err := webp.Encode(f, mainImg, optionsE); err != nil {
		panic(err)
	}
}
