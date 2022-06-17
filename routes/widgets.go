package routes

import (
	"bytes"
	"fmt"
	"html/template"
	"image/png"
	"io/ioutil"
	"net/http"
	"time"
	"wv2/types"
	"wv2/widgets"

	"github.com/gorilla/mux"
	"github.com/h2non/bimg"
)

var widgetdocs *template.Template

func init() {
	var err error
	widgetdocs, err = template.ParseFiles("templates/widgetdocs.html")
	if err != nil {
		panic(err)
	}
}

func WidgetsCreateWidget(w http.ResponseWriter, r *http.Request, opts types.RouteInfo) {
	if r.Method != "GET" {
		w.Write([]byte(invalidMethod))
		return
	}
	vars := mux.Vars(r)

	id := vars["id"]

	if id == "docs" {
		// Send over html docs here
		widgetdocs.Execute(w, nil)
		return
	}

	// Fetch bot from api-v3 blazefire
	req, err := http.NewRequest("GET", opts.APIUrl+"/blazefire/"+id, nil)

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
		return
	}

	err = json.Unmarshal(bytesD, &user)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
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
		return
	}

	bgcolor := r.URL.Query().Get("bgcolor")

	img, err := widgets.DrawWidget(widgetData, types.WidgetOptions{
		Bgcolor: bgcolor,
	})

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=28800")
	w.Header().Set("Expires", time.Now().Add(time.Hour*8).Format(http.TimeFormat))

	tmpBuf := bytes.NewBuffer([]byte{})

	err = png.Encode(tmpBuf, img)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
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
}
