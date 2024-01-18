/*
 * Copyright (c) Joseph Prichard 2023
 */

package main

import (
	"embed"
	_ "embed"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"guessthesketch/game"
	"guessthesketch/servers"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:embed words.txt
var words string

//go:embed dist/*
var dist embed.FS

//go:embed .env
var env string

func main() {
	envVars := parseEnv(env)
	jwtSecretKey := envVars["JWT_SECRET_KEY"]
	dbFile := envVars["DB_FILE"]

	db := createDb(dbFile)
	defer db.Close()

	gameWordBank := strings.Split(words, "\n")

	telemetryServer := servers.NewTelemetryServer()
	roomServer := servers.NewRoomServer(db)
	authServer := servers.NewAuthServer(jwtSecretKey)
	playerServer := servers.NewPlayerServer(db, authServer)
	drawingServer := servers.NewDrawingServer(db)
	brokerStore := game.NewBrokerStore(time.Minute)
	roomsServer := servers.NewRoomsServer(brokerStore, authServer, roomServer, gameWordBank)

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
	apiRouter.HandleFunc("/drawings", drawingServer.GetDrawings)
	addFileServer(router)

	log.Println("Starting the server...")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func createDb(dbFile string) *sqlx.DB {
	// create the db file if it doesn't exist
	_, err := os.Stat(dbFile)
	if os.IsNotExist(err) {
		_, err := os.Create(dbFile)
		if err != nil {
			log.Fatalf("Failed to create database file %s", dbFile)
			return nil
		}
	}

	db, err := sqlx.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	return db
}

func parseEnv(env string) map[string]string {
	envVars := make(map[string]string)

	lines := strings.Split(env, "\n")
	for _, line := range lines {
		index := strings.Index(line, "=")
		if index == -1 {
			log.Fatalf("Invalid env format: must contain an = symbol")
		} else {
			key := line[:index]
			value := line[index+1:]
			envVars[key] = value
		}
	}

	return envVars
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
