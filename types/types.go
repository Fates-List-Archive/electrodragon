package types

import (
	"bytes"
	"fmt"
	"image"
	"io/ioutil"
	"net/http"
	"time"
	"wv2/utils"

	"golang.org/x/image/webp"
)

type WidgetUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`

	// This must be created by API call
	AvatarBytes image.Image `json:"avatar_bytes"`

	// Temporary
	OutFile string
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

	r := bytes.NewReader(imgBytes)

	imgType, err := utils.GuessImageFormat(r)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Guessed type:", imgType)

	r = bytes.NewReader(imgBytes)

	avatar, err = webp.Decode(r)

	if err != nil {
		return err
	}

	w.AvatarBytes = avatar

	// Parse avatar into image.Image
	return nil
}
