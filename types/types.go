package types

import (
	"bytes"
	"fmt"
	"image"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"wv2/utils"

	"github.com/h2non/bimg"
	"golang.org/x/image/webp"
)

type WidgetUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Disc     string `json:"disc"`

	// This must be created by API call
	AvatarBytes image.Image `json:"avatar_bytes"`

	Bot bool `json:"bot"`
}

func (w *WidgetUser) ParseData() error {
	// Download avatar using net/http
	req, err := http.NewRequest("GET", strings.Replace(w.Avatar, ".gif", ".webp", 1), nil)

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

	r := bytes.NewReader(imgBytes)

	imgType, err := utils.GuessImageFormat(r)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Guessed type:", imgType)

	// Convert to webp using bimg
	imgBytes, err = bimg.NewImage(imgBytes).Convert(bimg.WEBP)

	if err != nil {
		return err
	}

	r = bytes.NewReader(imgBytes)

	avatar, err = webp.Decode(r)

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
