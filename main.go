package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"image/png"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"wv2/types"
	"wv2/utils"
	"wv2/widgets"

	integrase "github.com/MetroReviews/metro-integrase/lib"
	"github.com/gorilla/mux"
	"github.com/h2non/bimg"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valyala/fastjson"

	jsoniter "github.com/json-iterator/go"
)

const (
	notFoundPage  = "Not Found"
	internalError = "Something went wrong"
)

var (
	ctx        = context.Background()
	json       = jsoniter.ConfigCompatibleWithStandardLibrary
	devMode    bool
	api        string = "http://localhost:3010"
	widgetdocs *template.Template
)

func init() {
	var err error
	widgetdocs, err = template.ParseFiles("templates/widgetdocs.html")
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.BoolVar(&devMode, "dev", false, "Enable development mode")

	flag.Parse()

	if !devMode {
		if _, err := os.Stat(os.Getenv("HOME") + "/FatesList/config/data/secrets.json"); errors.Is(err, os.ErrNotExist) {
			panic("secrets.json not found")
		}
		file := os.Getenv("HOME") + "/FatesList/config/data/secrets.json"

		// Read file
		fileBytes, err := os.ReadFile(file)

		if err != nil {
			panic(err)
		}

		// Unmarshal using fastjson
		var p fastjson.Parser

		v, err := p.Parse(string(fileBytes))

		if err != nil {
			panic(err)
		}

		key, err := v.Get("metro_key").StringBytes()

		if err != nil {
			panic(err)
		}

		os.Setenv("SECRET_KEY", string(key))
	} else {
		os.Setenv("SECRET_KEY", "ABC")
	}

	os.Setenv("LIST_ID", "5800d395-beb3-4d79-90b9-93e1ca674b40")

	pool, err := pgxpool.Connect(ctx, "")

	if err != nil {
		panic(err)
	}

	fmt.Println(pool.Ping(ctx))

	if devMode {
		api = "https://api.fateslist.xyz"
	}

	// Get required variables

	r := mux.NewRouter()

	adp := DummyAdapter{}

	r.HandleFunc("/widgets/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		id := vars["id"]

		if id == "docs" {
			// Send over html docs here
			widgetdocs.Execute(w, nil)
			return
		}

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

	r.HandleFunc("/doctree", utils.CorsWrap(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var doctree []any

		// For every file, append its name into a slice, if its a directory, append its name and its children
		filepath.WalkDir("api-docs", func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			splitted := strings.Split(strings.Replace(path, "api-docs/", "", -1), "/")

			doctree = append(doctree, splitted)
			return nil
		})

		// Convert the slice into a json object
		json.NewEncoder(w).Encode(doctree)
	}))

	r.HandleFunc(`/docs/{rest:[a-zA-Z0-9=\-\/]+}`, utils.CorsWrap(func(w http.ResponseWriter, r *http.Request) {
		path := mux.Vars(r)["rest"]

		if strings.HasSuffix(path, ".md") {
			path = strings.Replace(path, ".md", "", 1)
		}

		if path == "" || path == "/docs" {
			path = "/index"
		}

		// Check if the file exists
		fmt.Println("api-docs/" + path + ".md")
		if _, err := os.Stat("api-docs/" + path + ".md"); os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Read the file
		file, err := ioutil.ReadFile("api-docs/" + path + ".md")

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}

		var data = types.Doc{}

		data.MD = string(file)

		// Look for javascript file in same place
		if _, err := os.Stat("api-docs/" + path + ".js"); err == nil {
			file, err := ioutil.ReadFile("api-docs/" + path + ".js")

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(internalError))
				return
			}

			data.JS = string(file)
		}

		json.NewEncoder(w).Encode(data)
	}))

	// Admin panel
	r.HandleFunc("/ap/schema", utils.CorsWrap(func(w http.ResponseWriter, r *http.Request) {
		opts := utils.SchemaFilter{}

		if r.URL.Query().Get("table_name") != "" {
			opts.TableName = r.URL.Query().Get("table_name")
		}

		res, err := utils.GetSchema(ctx, pool, opts)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}

		bytes, err := json.Marshal(res)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}))

	r.HandleFunc("/ap/schema/allowed-tables", utils.CorsWrap(func(w http.ResponseWriter, r *http.Request) {
		auth, err := utils.AuthorizeUser(utils.AuthRequest{
			UserID:  r.URL.Query().Get("user_id"),
			Token:   r.Header.Get("Authorization"),
			DevMode: devMode,
			Context: ctx,
			DB:      pool,
		})

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		bytes, err := json.Marshal(auth.AllowedTables)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}

		w.Write(bytes)
	}))

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(notFoundPage))
	})

	integrase.StartServer(adp, r)

}
