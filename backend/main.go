package main

import (
	"embed"
	"fmt"
	"guessasketch/api"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

//go:embed words.txt
var f embed.FS

func main() {
	err := godotenv.Load()
	if err != nil {
	  log.Fatal("Error loading .env file")
	}

	rand.Seed(time.Now().UnixNano())
	log.Printf("Started the server...")

	data, err := f.ReadFile("words.txt")
	if err != nil {
        fmt.Println("Error reading embedded file:", err)
        return
    }
	gameWordBank := strings.Split(string(data), "\n")
	// log.Printf("Word bank size %s", string(data))

	wsServer := api.NewWsServer(gameWordBank)
	playerServer := api.NewPlayerServer(nil)

	http.HandleFunc("/rooms/create", wsServer.CreateRoom)
	http.HandleFunc("/rooms/join", wsServer.JoinRoom)
	http.HandleFunc("/players/stats", playerServer.Get)
	http.HandleFunc("/players/leaderboard", playerServer.Leaderboard)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
