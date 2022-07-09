package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"wv2/routes"
	"wv2/types"
	"wv2/utils"

	integrase "github.com/MetroReviews/metro-integrase/lib"
	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valyala/fastjson"
)

var (
	ctx         = context.Background()
	devMode     bool
	api         string = "http://localhost:3010"
	mainServer  string
	staffServer string
	metro       *discordgo.Session
	pool        *pgxpool.Pool
	redisPool   *redis.Client
)

func Route(fn routes.RouteFunc) utils.HFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Frostpaw-ID, Frostpaw-MFA, Authorization, Frostpaw-Pass")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.Write([]byte(""))
			return
		}
		fn(w, r, types.RouteInfo{
			DevMode:     devMode,
			DB:          pool,
			Context:     ctx,
			Redis:       redisPool,
			MainServer:  mainServer,
			StaffServer: staffServer,
			Bot:         metro,
			APIUrl:      api,
		})
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

	mainServer = string(servers.Get("main").GetStringBytes())
	staffServer = string(servers.Get("staff").GetStringBytes())

	metro, err = discordgo.New("Bot " + string(metroKey))

	if err != nil {
		panic(err)
	}

	if err := metro.Open(); err != nil {
		panic(err)
	}

	pool, err = pgxpool.Connect(ctx, "")

	if err != nil {
		panic(err)
	}

	redisPool = redis.NewClient(&redis.Options{
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
	loadRoutes(r)

	adp := DummyAdapter{}

	integrase.StartServer(adp, integrase.MuxWrap{Router: r})

	log := handlers.LoggingHandler(os.Stdout, r)

	http.ListenAndServe(":1800", log)
}
