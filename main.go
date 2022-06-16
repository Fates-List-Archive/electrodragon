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
	"github.com/alexedwards/argon2id"
	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/h2non/bimg"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp/totp"
	"github.com/valyala/fastjson"

	jsoniter "github.com/json-iterator/go"

	"github.com/happeens/xkcdpass"
)

const (
	notFoundPage  = "Not Found"
	internalError = "Something went wrong"
	invalidMethod = "Invalid method"
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

	os.Setenv("LIST_ID", "5800d395-beb3-4d79-90b9-93e1ca674b40")

	metroKey, err := v.Get("token_main").StringBytes()

	if err != nil {
		panic(err)
	}

	discordJson, err := os.ReadFile(os.Getenv("HOME") + "/FatesList/config/data/discord.json")

	if err != nil {
		panic(err)
	}

	v, err = p.Parse(string(discordJson))

	if err != nil {
		panic(err)
	}

	var servers = v.GetObject("servers")

	mainServer := string(servers.Get("main").GetStringBytes())
	staffServer := string(servers.Get("staff").GetStringBytes())

	metro, err = discordgo.New("Bot " + string(metroKey))

	if err != nil {
		panic(err)
	}

	if err := metro.Open(); err != nil {
		panic(err)
	}

	pool, err := pgxpool.Connect(ctx, "")

	if err != nil {
		panic(err)
	}

	redisPool := redis.NewClient(&redis.Options{
		Addr:     "localhost:1001",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	fmt.Println(pool.Ping(ctx))

	if devMode {
		api = "https://api.fateslist.xyz"
	}

	// Get required variables

	r := mux.NewRouter()

	adp := DummyAdapter{}

	r.HandleFunc("/widgets/{id}", func(w http.ResponseWriter, r *http.Request) {
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
	})

	r.HandleFunc("/doctree", utils.CorsWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.Write([]byte(invalidMethod))
			return
		}

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

			if strings.HasSuffix(path, ".js") {
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
		if r.Method != "GET" {
			w.Write([]byte(invalidMethod))
			return
		}

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
		if r.Method != "GET" {
			w.Write([]byte(invalidMethod))
			return
		}

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
		if r.Method != "GET" {
			w.Write([]byte(invalidMethod))
			return
		}

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

	r.HandleFunc("/ap/newcat", utils.CorsWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Write([]byte(invalidMethod))
			return
		}

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

		if auth.Verified {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("You are already verified"))
			return
		}

		if auth.Perms.Perm < 2 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("You do not have permission to do this"))
			return
		}

		// Check code sent in request body
		defer r.Body.Close()

		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}

		var verified bool

		if devMode {
			verified = utils.CheckCodeDev(r.URL.Query().Get("user_id"), string(body))
		} else {
			verified = utils.CheckCodeSecure(r.URL.Query().Get("user_id"), string(body))
		}

		if !verified {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Invalid code"))
			return
		}

		newTotp, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "Fates List Electrodragon Auth",
			AccountName: r.URL.Query().Get("user_id"),
		})

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}

		newPass := xkcdpass.GenerateWithLength(6)

		newPassHashed, err := argon2id.CreateHash(newPass, argon2id.DefaultParams)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}

		_, err = pool.Exec(ctx, "UPDATE users SET staff_verify_code = $1, staff_password = $2 WHERE user_id = $3", body, newPassHashed, r.URL.Query().Get("user_id"))

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}

		metro.GuildMemberRoleAdd(mainServer, r.URL.Query().Get("user_id"), auth.Perms.ID)
		metro.GuildMemberRoleAdd(staffServer, r.URL.Query().Get("user_id"), auth.Perms.StaffID)

		imgHash := utils.RandString(512)

		imgQr, err := newTotp.Image(150, 150)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}

		// Convert imgQr to webp
		buf := new(bytes.Buffer)

		png.Encode(buf, imgQr)

		// Insert newTotp.Image() into redis
		redisPool.Set(ctx, imgHash, buf.Bytes(), 5*time.Minute)

		imageUrl := "https://corona.fateslist.xyz/qr/" + imgHash

		if devMode {
			imageUrl = "https://localhost:1800/qr/" + imgHash
		}

		data := types.NewStaff{
			Pass:      newPass,
			SharedKey: newTotp.Secret(),
			Image:     imageUrl,
		}

		json.NewEncoder(w).Encode(data)
	}))

	r.HandleFunc("/qr/{hash}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			w.Write([]byte(invalidMethod))
			return
		}

		hash := mux.Vars(r)["hash"]

		// Get image from redis
		img, err := redisPool.Get(ctx, hash).Result()

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}

		// Write image to response
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte(img))
	})

	r.HandleFunc("/ap/pouncecat", utils.CorsWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Write([]byte(invalidMethod))
			return
		}

		auth, err := utils.AuthorizeUser(utils.AuthRequest{
			UserID:   r.URL.Query().Get("user_id"),
			Token:    r.Header.Get("Authorization"),
			TOTP:     r.Header.Get("Frostpaw-MFA"),
			Password: r.Header.Get("Frostpaw-Pass"),
			DevMode:  devMode,
			Context:  ctx,
			DB:       pool,
		})

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		if !auth.PasswordLogin {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Password incorrect. Retry staff verification if you have not done it before"))
			return
		}

		if !auth.MFA {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("MFA incorrect. Retry staff verification if you have not done it before"))
			return
		}

		if auth.Perms.Perm < 2 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("You do not have permission to do this"))
			return
		}

		session := utils.RandString(512)

		authSession := map[string]string{
			"user_id": r.URL.Query().Get("user_id"),
			"token":   r.Header.Get("Authorization"),
		}

		bytes, err := json.Marshal(authSession)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}

		redisPool.Set(ctx, session, bytes, time.Hour*2)

		w.Write([]byte(session))
	}))

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(notFoundPage))
	})

	integrase.StartServer(adp, r)

}
