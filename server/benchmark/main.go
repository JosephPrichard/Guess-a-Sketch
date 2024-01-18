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
	"sync"
	"time"
)

const HOST = "localhost:8080"

func createRoom() servers.RoomCodeResp {
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
		log.Fatalf("Failed to unmarshal json %v", err)
	}
	return roomResp
}

func runRoomClient(playerCount int, roomsWg *sync.WaitGroup) {
	if playerCount < 1 {
		log.Fatalf("Cannot start a room with less than 1 player")
	}

	roomResp := createRoom()
	wss := joinPlayersToRoom(roomResp.Code, playerCount)

	// wait on all player clients to finish
	var playersWg sync.WaitGroup
	playersWg.Add(playerCount)

	// start a host and 9 players
	go runPlayerClient(wss[0], true, &playersWg)
	for i := 1; i < playerCount; i++ {
		go runPlayerClient(wss[i], false, &playersWg)
	}

	playersWg.Wait()
	roomsWg.Done()
}

func joinPlayersToRoom(code string, count int) []*websocket.Conn {
	// create connections and join a room for each player client
	wss := make([]*websocket.Conn, 0)
	for i := 0; i < count; i++ {
		u := fmt.Sprintf("ws://%s/api/rooms/join?code=%s", HOST, code)
		ws, _, err := websocket.DefaultDialer.Dial(u, nil)
		if err != nil {
			log.Fatalf("%v", err)
		}
		wss = append(wss, ws)
	}
	return wss
}

func runPlayerClient(ws *websocket.Conn, isHost bool, wg *sync.WaitGroup) {
	// listen to messages from connection, send a stop signal when we're finished
	go func(ws *websocket.Conn) {
		for {
			_, buf, err := ws.ReadMessage()
			if err != nil {
				log.Printf("Server closed connection with err %s", err.Error())
				return
			}
			log.Printf("Received a message from server %s", string(buf))
		}
	}(ws)

	if isHost {
		sendStartMessage(ws)
	}
	// send a draw message 24 times per second
	for range time.NewTicker(time.Second / 24).C {
		sendDrawMessage(ws)
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
	roomCount := 1
	var roomsWg sync.WaitGroup

	for i := 0; i < roomCount; i++ {
		roomsWg.Add(1)
		go runRoomClient(10, &roomsWg)
	}

	roomsWg.Wait()
}
