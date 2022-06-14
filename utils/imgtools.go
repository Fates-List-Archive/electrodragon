// https://goplay.space/#_RTH0rWA7Ae
package utils

import (
	"image"
	"image/color"
	"image/draw"
)

type circle struct {
	centerPoint  image.Point
	radius       int
	transparents []image.Point
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

func Circle(src image.Image) image.Image {
	// create a pure black dst img
	dst := image.NewRGBA(src.Bounds())
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.RGBA{0, 0, 0, 255}), image.Point{}, draw.Src)

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
