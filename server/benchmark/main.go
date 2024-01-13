/*
 * Copyright (c) Joseph Prichard 2024
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"guessthesketch/game"
	"guessthesketch/servers"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const HOST = "localhost:8080"

func startRoom(playerCount int) {
	if playerCount < 1 {
		log.Fatalf("Cannot start a room with less than 1 player")
	}

	u := fmt.Sprintf("http://%s/api/rooms/create", HOST)

	roomSettings := game.RoomSettings{
		PlayerLimit:   10,
		TimeLimitSecs: 200,
		TotalRounds:   6,
	}
	jsonBody, err := json.Marshal(roomSettings)
	if err != nil {
		log.Fatalf("Failed to marshal json %v", err)
	}
	resp, err := http.Post(u, "text/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading resp body: %s", err)
	}

	var roomResp servers.RoomCodeResp
	err = json.Unmarshal(body, &roomResp)
	if err != nil {
		log.Fatalf("Failed to marshal json %v", err)
	}
	code := roomResp.Code

	// each player waits for all other players to finish joining
	var wg sync.WaitGroup
	wg.Add(playerCount)

	// start a host and 9 players
	go startPlayer(code, true, &wg)
	for j := 0; j < playerCount-1; j++ {
		go startPlayer(code, false, &wg)
	}
}

func startPlayer(code string, isHost bool, wg *sync.WaitGroup) {
	u := fmt.Sprintf("ws://%s/api/rooms/join?code=%s", HOST, code)
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// when done, wait for all other players to join the room
	wg.Done()
	wg.Wait()

	go startListen(ws)

	if isHost {
		sendStartMessage(ws)
	}
	// send a draw message 24 times per second
	for range time.NewTicker(time.Second / 24).C {
		sendDrawMessage(ws)
	}
}

func startListen(ws *websocket.Conn) {
	for {
		_, buf, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Server closed connection with err %s", err.Error())
			return
		}
		log.Printf("Received a message from server %s", string(buf))
	}
}

func sendStartMessage(ws *websocket.Conn) {
	input := game.InputPayload[struct{}]{Code: game.StartCode}
	b, err := json.Marshal(input)
	if err != nil {
		log.Fatalf("Failed to marshal json %v", err)
	}

	err = ws.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		log.Printf("Failed to send start message %v", err)
		return
	}
}

func sendDrawMessage(ws *websocket.Conn) {
	input := game.InputPayload[game.DrawMsg]{
		Code: game.DrawCode,
		Msg: game.DrawMsg{
			X:      uint16(rand.Intn(game.MaxX - 1)),
			Y:      uint16(rand.Intn(game.MaxY - 1)),
			Radius: 1,
			Color:  1,
		},
	}

	b, err := json.Marshal(input)
	if err != nil {
		log.Fatalf("Failed to marshal json %v", err)
	}

	err = ws.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		log.Printf("Failed to send start message %v", err)
		return
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	strCount := os.Args[1]
	count, err := strconv.Atoi(strCount)
	if err != nil {
		log.Fatalf("Count argument must be a number: %d", count)
	}

	for i := 0; i < count; i++ {
		go startRoom(10)
	}
}
