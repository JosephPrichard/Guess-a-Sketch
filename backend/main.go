package main

import (
	_ "embed"
	"guessasketch/server"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

//go:embed words.txt
var words string

func main() {
	var envVars map[string]string
	envVars, err := godotenv.Read()
	if err != nil {
		panic(err)
	}
	log.Println("Env vars", envVars)

	rand.Seed(time.Now().UnixNano())
	log.Println("Started the server...")

	gameWordBank := strings.Split(words, "\n")

	authServer := server.NewAuthServer(envVars["JWT_SECRET_KEY"])
	playerServer := server.NewPlayerServer(nil)
	wsServerConfig := server.RoomsServerConfig{
		GameWordBank: gameWordBank,
		AuthServer:   authServer,
		PlayerServer: playerServer,
	}
	roomsServer := server.NewRoomsServer(wsServerConfig)

	http.HandleFunc("/rooms/create", roomsServer.CreateRoom)
	http.HandleFunc("/rooms/join", roomsServer.JoinRoom)
	http.HandleFunc("/rooms", roomsServer.Rooms)
	http.HandleFunc("/players/stats", playerServer.Get)
	http.HandleFunc("/players/leaderboard", playerServer.Leaderboard)
	http.HandleFunc("/login", authServer.Login)
	http.HandleFunc("/logout", authServer.Logout)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
