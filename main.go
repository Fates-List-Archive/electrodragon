package main

import (
	"bytes"
	"flag"
	"fmt"
	"image/png"
	"io/ioutil"
	"net/http"
	"time"
	"wv2/types"
	"wv2/widgets"

	integrase "github.com/MetroReviews/metro-integrase/lib"
	"github.com/gorilla/mux"
	"github.com/h2non/bimg"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	notFoundPage  = "{\"message\":\"NotFound\"}"
	internalError = "{\"message\":\"InternalError\"}"
)

var (
	devMode bool
	api     string = "http://localhost:3010"
)

func main() {
	flag.BoolVar(&devMode, "dev", false, "Enable development mode")

	flag.Parse()

	if devMode {
		api = "https://api.fateslist.xyz"
	}

	r := mux.NewRouter()

	adp := DummyAdapter{}

	r.HandleFunc("/widgets/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		id := vars["id"]

		// Fetch bot from api-v3 blazefire
		req, err := http.NewRequest("GET", api+"/blazefire/"+id, nil)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}

		client := http.Client{Timeout: 10 * time.Second}

		resp, err := client.Do(req)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Println(resp.StatusCode)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Invalid status code from main site: " + resp.Status))
			return
		}

		// Read the user info from the response
		var user types.User

		defer resp.Body.Close()

		bytesD, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		err = json.Unmarshal(bytesD, &user)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		widgetData := types.WidgetUser{
			ID:       user.ID,
			Username: user.Username,
			Avatar:   user.Avatar,
			Disc:     user.Disc,
			Bot:      user.Bot,
		}

		err = widgetData.ParseData()

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		img := widgets.DrawWidget(widgetData)

		w.Header().Set("Cache-Control", "public, max-age=28800")
		w.Header().Set("Expires", time.Now().Add(time.Hour*8).Format(http.TimeFormat))

		tmpBuf := bytes.NewBuffer([]byte{})

		err = png.Encode(tmpBuf, img)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		format := r.URL.Query().Get("format")

		if format == "png" {
			w.Header().Set("Content-Type", "image/png")
			w.Write(tmpBuf.Bytes())
		} else {
			w.Header().Set("Content-Type", "image/webp")

			bimgImg, err := bimg.NewImage(tmpBuf.Bytes()).Convert(bimg.WEBP)

			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}

			w.Write(bimgImg)
		}
	})

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(notFoundPage))
	})

	integrase.StartServer(adp, r)

}
