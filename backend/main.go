/*
 * Copyright (c) Joseph Prichard 2023
 */

package main

import (
	_ "embed"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"guessasketch/game"
	"guessasketch/server"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

//go:embed words.txt
var words string

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

	roomServer := server.NewRoomServer(db)
	authServer := server.NewAuthServer(jwtSecretKey)
	playerServer := server.NewPlayerServer(db, authServer)

	roomsStore := game.NewRoomsMap(time.Minute)
	roomsServer := server.NewRoomsServer(roomsStore, authServer, roomServer, gameWordBank)
	metaServer := server.NewMetaServer()

	http.HandleFunc("/rooms/create", roomsServer.CreateRoom)
	http.HandleFunc("/rooms/join", roomsServer.JoinRoom)
	http.HandleFunc("/rooms", roomsServer.Rooms)
	http.HandleFunc("/players/stats", playerServer.Get)
	http.HandleFunc("/players/leaderboard", playerServer.Leaderboard)
	http.HandleFunc("/login", authServer.Login)
	http.HandleFunc("/logout", authServer.Logout)
	http.HandleFunc("/meta", metaServer.Subscribe)

	log.Println("Starting the server...")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
