// https://goplay.space/#_RTH0rWA7Ae
package imgtools

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"wv2/types"

	"github.com/golang/freetype"
	"github.com/h2non/bimg"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
)

type circle struct {
	centerPoint image.Point
	radius      int
}

func (c *circle) ColorModel() color.Model {
	return color.AlphaModel
}

func (c *circle) Bounds() image.Rectangle {
	return image.Rect(
		c.centerPoint.X-c.radius,
		c.centerPoint.Y-c.radius,
		c.centerPoint.X+c.radius,
		c.centerPoint.Y+c.radius,
	)
}

func (c *circle) At(x, y int) color.Color {
	xpos := float64(x-c.centerPoint.X) + 0.5
	ypos := float64(y-c.centerPoint.Y) + 0.5
	radiusSquared := float64(c.radius * c.radius)
	if xpos*xpos+ypos*ypos < radiusSquared {
		return color.RGBA{255, 255, 255, 255}
	}

	return color.RGBA{0, 0, 0, 0}
}

func Circle(src image.Image, colorEdge color.Color) image.Image {
	// create a pure black dst img
	dst := image.NewRGBA(src.Bounds())
	draw.Draw(dst, dst.Bounds(), image.NewUniform(colorEdge), image.Point{}, draw.Src)

	r := src.Bounds().Dx() / 2
	p := image.Point{
		X: src.Bounds().Dx() / 2,
		Y: src.Bounds().Dy() / 2,
	}

	c1 := &circle{
		centerPoint: p,
		radius:      r,
	}

	draw.DrawMask(dst, dst.Bounds(), src, image.Point{}, c1, image.Point{}, draw.Over)

	return dst
}

func ScaleImage(imgBuf []byte, width, height int) (image.Image, error) {
	// Read the image from the bytes
	newImage := bimg.NewImage(imgBuf)
	resized, err := newImage.ResizeAndCrop(width, height)

	if err != nil {
		return nil, err
	}

	// Re-encode to PNG
	buf := bytes.NewBuffer(resized)
	pngData, err := png.Decode(buf)

	if err != nil {
		return nil, err
	}

	// Convert to draw.Image

	return pngData, nil
}

func WatermarkImage(img image.Image, text string) image.Image {
	watermark := bimg.Watermark{
		Text:       text,
		Opacity:    1,
		Width:      200,
		DPI:        100,
		Margin:     150,
		Font:       "assets/font.ttf",
		Background: bimg.Color{R: 255, G: 255, B: 255},
	}

	bytesBuf := new(bytes.Buffer)

	err := png.Encode(bytesBuf, img)

	if err != nil {
		fmt.Println(err)
		return img
	}

	watermarked, err := bimg.NewImage(bytesBuf.Bytes()).Watermark(watermark)

	if err != nil {
		fmt.Println(err)
		return img
	}

	// Re-encode to PNG
	buf := bytes.NewBuffer(watermarked)
	pngData, err := png.Decode(buf)

	if err != nil {
		fmt.Println(err)
		return img
	}

	return pngData
}

func AddLabel(img *image.RGBA, label types.Label) (ptX, ptY int) {
	c := freetype.NewContext()
	c.SetDPI(label.DPI)
	c.SetFont(label.FontData)
	c.SetFontSize(label.Size)
	c.SetHinting(font.HintingNone)

	// Set source (https://github.com/golang/freetype/blob/master/example/freetype/main.go)

	c.SetClip(img.Bounds())
	c.SetSrc(image.NewUniform(label.Color))
	c.SetDst(img)

	// Draw the text.
	lastLineLen := 0

	pt := freetype.Pt(label.X, label.Y+int(c.PointToFixed(label.Size)>>6))
	for _, s := range label.Labels {
		_, err := c.DrawString(s, pt)
		if err != nil {
			fmt.Println(err)
			return 0, 0
		}
		pt.Y += c.PointToFixed(label.Size * label.Spacing)
		lastLineLen = len(s) + 1
	}

	return lastLineLen, pt.Y.Ceil()
}

func ResizeImage(img image.Image, factor int) draw.Image {
	// Read the image from the file
	dst := image.NewRGBA(image.Rect(0, 0, img.Bounds().Max.X/factor, img.Bounds().Max.Y/factor))

	draw.BiLinear.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)

	return dst
}

func CopyImage(x, y int, img image.Image, dst draw.Image) {
	dp := image.Point{X: x, Y: y}

	// Carve out rectangle for the image
	draw.Draw(dst, image.Rectangle{Min: dp, Max: dp.Add(img.Bounds().Size())}, img, image.Point{}, draw.Src)
}

func GetColorSimilarity(c1 color.Color, c2 color.Color) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	return math.Abs(float64(r1-r2)) + math.Abs(float64(g1-g2)) + math.Abs(float64(b1-b2))
}

func ReplaceImageColor(img draw.Image, c1 color.Color, c2 color.Color) *image.RGBA {
	// Copy image
	dst := image.NewRGBA(img.Bounds())
	draw.Draw(dst, dst.Bounds(), img, image.Point{}, draw.Src)

	for x := dst.Bounds().Min.X; x < dst.Bounds().Max.X; x++ {
		for y := dst.Bounds().Min.Y; y < dst.Bounds().Max.Y; y++ {
			if GetColorSimilarity(dst.At(x, y), c1) < 100 {
				dst.Set(x, y, c2)
			}
		}
	}

	return dst
}
