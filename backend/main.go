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
)

//go:embed words.txt
var f embed.FS

func main() {
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
	playerServer := api.NewPlayerServer()

	http.HandleFunc("/rooms/create", wsServer.CreateRoom)
	http.HandleFunc("/rooms/join", wsServer.JoinRoom)
	http.HandleFunc("/players/stats", playerServer.Get)
	http.HandleFunc("/players/leaderboard", playerServer.Leaderboard)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
