package main

import (
	"embed"
	"fmt"
	"guessasketch/server"
	"log"
	"net/http"
	"strings"
)

//go:embed words.txt
var f embed.FS

func main() {
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
	log.Fatal(http.ListenAndServe(":8080", nil))
}
