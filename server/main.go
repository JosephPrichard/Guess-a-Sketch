/*
 * Copyright (c) Joseph Prichard 2023
 */

package main

import (
	"embed"
	_ "embed"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"guessthesketch/game"
	"guessthesketch/servers"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

//go:embed words.txt
var words string

//go:embed dist/*
var dist embed.FS

func main() {
	rand.Seed(time.Now().UnixNano())

	var envVars map[string]string
	envVars, err := godotenv.Read()
	if err != nil {
		panic(err)
	}
	log.Println("Env vars", envVars)

	jwtSecretKey := envVars["JWT_SECRET_KEY"]
	dbUser := envVars["DB_USER"]
	dbName := envVars["DB_NAME"]
	dbHost := envVars["DB_HOST"]
	dbPass := envVars["DB_PASSWORD"]
	dbPort := envVars["DB_PORT"]

	gameWordBank := strings.Split(words, "\n")

	dataSource := fmt.Sprintf("user=%s dbname=%s host=%s password=%s port=%s sslmode=disable",
		dbUser, dbName, dbHost, dbPass, dbPort)

	db, err := sqlx.Connect("postgres", dataSource)
	if err != nil {
		log.Fatalln(err)
		return
	}

	roomServer := servers.NewRoomServer(db)
	authServer := servers.NewAuthServer(jwtSecretKey)
	playerServer := servers.NewPlayerServer(db, authServer)

	roomsStore := game.NewRoomsMap(time.Minute)
	roomsServer := servers.NewRoomsServer(roomsStore, authServer, roomServer, gameWordBank)
	telemetryServer := servers.NewTelemetryServer()

	router := mux.NewRouter()

	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/rooms/create", roomsServer.CreateRoom)
	apiRouter.HandleFunc("/rooms/join", roomsServer.JoinRoom)
	apiRouter.HandleFunc("/rooms", roomsServer.GetRooms)
	apiRouter.HandleFunc("/players/stats", playerServer.Get)
	apiRouter.HandleFunc("/players/leaderboard", playerServer.Leaderboard)
	apiRouter.HandleFunc("/login", authServer.Login)
	apiRouter.HandleFunc("/logout", authServer.Logout)
	apiRouter.HandleFunc("/telemetry/subscribe", telemetryServer.Subscribe)

	addFileServer(router)

	log.Println("Starting the server...")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func addFileServer(router *mux.Router) {
	// handler function to serve all "spa" files in embed fs dist
	var extContentMap = map[string]string{
		"css":  "text/css",
		"html": "text/html",
		"js":   "text/javascript",
	}

	fsys, err := fs.Sub(dist, "dist")
	if err != nil {
		log.Fatal(err)
	}
	fileHandler := http.FileServer(http.FS(fsys))

	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.String()

		// empty route serves index html
		if url == "/" || url == "" {
			w.Header().Add("Content-type", "text/html")
		}

		tokens := strings.Split(url, ".")
		if len(tokens) >= 1 {
			contentType := extContentMap[tokens[len(tokens)-1]]
			w.Header().Add("Content-type", contentType)
		}

		fileHandler.ServeHTTP(w, r)
	})
}
