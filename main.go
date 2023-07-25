package main

import (
	"embed"
	"guessasketch/server"
	"log"
	"net/http"
	"strings"
)

func getWordBank() []string {
	//go:embed words.txt
	var f embed.FS
	data, _ := f.ReadFile("hello.txt")
	gameWordBank := strings.Split(string(data), "\n")
	log.Printf("Word bank size %d", len(gameWordBank))
	return gameWordBank
}

func main() {
	log.Printf("Started the server...")

	gameWordBank := getWordBank()
	wsController := server.NewWsController(gameWordBank)

	http.HandleFunc("/rooms/create", wsController.CreateRoom)
	http.HandleFunc("/rooms/join", wsController.JoinRoom)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
