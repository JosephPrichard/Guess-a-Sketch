/*
 * Copyright (c) Joseph Prichard 2023
 */

package server

import (
	crand "crypto/rand"
	"encoding/hex"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"guessthesketch/database"
	"guessthesketch/game"
	"log"
	"net/http"
	"strconv"
)

type RoomsServer struct {
	upgrade       websocket.Upgrader
	rooms         game.RoomsStore
	authenticator Authenticator
	worker        game.RoomWorker
	gameWordBank  []string
}

func NewRoomsServer(rooms game.RoomsStore, authenticator Authenticator, worker game.RoomWorker, gameWordBank []string) *RoomsServer {
	return &RoomsServer{
		upgrade:       CreateUpgrade(),
		rooms:         rooms,
		authenticator: authenticator,
		worker:        worker,
		gameWordBank:  gameWordBank,
	}
}

func (server *RoomsServer) Rooms(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)

	query := r.URL.Query()
	offsetStr := query.Get("offsetStr")

	offset := 0
	if offsetStr != "" {
		parsedOffset, err := strconv.ParseInt(offsetStr, 10, 32)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "Offset parameters must be a 32-bit integer")
			return
		}
		offset = int(parsedOffset)
	}

	rooms := server.rooms.Codes(offset, 20)
	w.WriteHeader(http.StatusOK)
	WriteJson(w, rooms)
}

func HexCode(len int) (string, error) {
	b := make([]byte, len/2)
	_, err := crand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (server *RoomsServer) CreateRoom(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)

	// generate a code, create a room, start it, then store it in the map
	code, err := HexCode(8)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to generate a valid room code")
		return
	}

	var settings game.RoomSettings
	err = ReadJson(r, &settings)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	game.SettingsWithDefaults(&settings)
	settings.SharedWordBank = server.gameWordBank

	err = game.IsSettingsValid(settings)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	initialState := game.NewGameState(code, settings)
	room := game.NewGameRoom(initialState, server.worker)
	log.Printf("Starting a room for code %s", code)
	go room.Start()
	server.rooms.Store(code, room)

	type RoomCodeResp struct {
		Code     string            `json:"code"`
		Settings game.RoomSettings `json:"settings"`
	}
	roomCode := RoomCodeResp{Code: code, Settings: settings}
	w.WriteHeader(http.StatusOK)
	WriteJson(w, roomCode)
}

func (server *RoomsServer) JoinRoom(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)

	query := r.URL.Query()
	code := query.Get("code")
	token := query.Get("token")

	player := server.authenticator.GetPlayer(token)

	room := server.rooms.Load(code)
	if room == nil {
		WriteError(w, http.StatusNotFound, "Cannot find room for provided code")
		return
	}

	ws, err := server.upgrade.Upgrade(w, r, nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to upgrade to websocket")
		return
	}

	// create a new subscription channel and join the room with it
	subscriber := make(game.Subscriber)
	room.Join(game.SubscriberMsg{Subscriber: subscriber, Player: player})

	log.Printf("Joined room %s with name %s and id %s", code, player.Name, player.ID)

	go server.subscriberListener(ws, subscriber)
	go server.socketListener(ws, room, subscriber)
}

// reads messages from socket and sends them to room
func (server *RoomsServer) socketListener(ws *websocket.Conn, room game.Room, subscriber game.Subscriber) {
	defer func() {
		// unsubscribes from the room when the websocket is closed
		room.Leave(subscriber)
		_ = ws.Close()
		log.Printf("Socket listener close function called")
		if panicInfo := recover(); panicInfo != nil {
			log.Println(panicInfo)
		}
	}()
	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Client closed connection with err %s", err.Error())
			return
		}
		// read any message from the socket and broadcast it to the room
		log.Println("Receiving message", string(p))
		room.SendMessage(game.SentMsg{Message: p, Sender: subscriber})
	}
}

// reads messages from a subscribed channel and sends them to socket
func (server *RoomsServer) subscriberListener(ws *websocket.Conn, subscriber game.Subscriber) {
	defer func() {
		// closes the websocket connection when the subscriber is informed no more messages will be sent
		log.Println("Subscriber channel was closed")
		_ = ws.Close()
		if panicInfo := recover(); panicInfo != nil {
			log.Println(panicInfo)
		}
	}()
	for resp := range subscriber {
		// read values from channel and write back to socket
		log.Println("Sending message", string(resp))
		err := ws.WriteMessage(websocket.TextMessage, resp)
		if err != nil {
			log.Printf("Error writing message %s", err)
			return
		}
	}
}

type RoomServer struct {
	db *sqlx.DB
}

func NewRoomServer(db *sqlx.DB) *RoomServer {
	return &RoomServer{db}
}

func (server RoomServer) DoShutdown(results []game.GameResult) {
	// perform the batch update stats on a separate goroutine
	go database.UpdateStats(server.db, results)
}

func (server RoomServer) DoCapture(drawing game.Snapshot) {
	// perform a capture operation on a separate goroutine
	go database.SaveDrawing(server.db, drawing)
}
