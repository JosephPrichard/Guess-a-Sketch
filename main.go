package main

import (
	"embed"
	"fmt"
	"guessasketch/server"
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

	wsController := server.NewWsController(gameWordBank)

	http.HandleFunc("/rooms/create", wsController.CreateRoom)
	http.HandleFunc("/rooms/join", wsController.JoinRoom)
	http.HandleFunc("/rooms/random", wsController.GetRandomCode)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
