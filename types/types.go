package types

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/h2non/bimg"
)

type WidgetOptions struct {
	Bgcolor string
}

type WidgetUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Disc     string `json:"disc"`

	// This must be created by API call
	AvatarBytes image.Image `json:"avatar_bytes"`

	// Whether or not its a bot or not
	Bot bool `json:"bot"`

	// Whether or not its a server or not
	Server bool `json:"server"`
}

func (w *WidgetUser) ParseData() error {
	// Download avatar using net/http
	req, err := http.NewRequest("GET", w.Avatar, nil)

	if err != nil {
		return err
	}

	// Get the response
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	fmt.Println(resp.StatusCode)

	// Read the image from the response
	var avatar image.Image

	defer resp.Body.Close()

	imgBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	fmt.Println("Image bytes:", len(imgBytes))

	// Convert to png using bimg
	imgBytes, err = bimg.NewImage(imgBytes).Convert(bimg.PNG)

	if err != nil {
		return err
	}

	r := bytes.NewReader(imgBytes)

	avatar, err = png.Decode(r)

	if err != nil {
		return err
	}

	w.AvatarBytes = avatar

	// Parse avatar into image.Image
	return nil
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Bot      bool   `json:"bot"`
	Disc     string `json:"discriminator"`
	Status   string `json:"status"`
}

type Label struct {
	DPI      float64
	FontData *truetype.Font
	Size     float64
	Spacing  float64
	Labels   []string
	Color    color.Color
	X        int
	Y        int
}

type Doc struct {
	MD string `json:"data"`
	JS string `json:"js"`
}

type NewStaff struct {
	Pass      string `json:"pass"`
	SharedKey string `json:"totp_shared_key"`

	// TOTP image hex
	Image string `json:"image"`
}
