/*
 * Copyright (c) Joseph Prichard 2023
 */

package servers

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
	brokerage     game.Brokerage
	authenticator Authenticator
	handler       game.EventHandler
	gameWordBank  []string
}

func NewRoomsServer(
	brokerage game.Brokerage, authenticator Authenticator,
	handler game.EventHandler, gameWordBank []string) *RoomsServer {

	return &RoomsServer{
		upgrade:       CreateUpgrade(),
		brokerage:     brokerage,
		authenticator: authenticator,
		handler:       handler,
		gameWordBank:  gameWordBank,
	}
}

func (server *RoomsServer) GetRooms(w http.ResponseWriter, r *http.Request) {
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

	rooms := server.brokerage.Codes(offset, 20)
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

type RoomCodeResp struct {
	Code     string            `json:"code"`
	Settings game.RoomSettings `json:"settings"`
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
	room := game.NewRoom(initialState, settings.IsPublic, server.handler)
	go room.Start()
	server.brokerage.Set(code, room)

	log.Printf("Started room for code %s", code)

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

	room := server.brokerage.Get(code)
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
	subscriber := make(chan []byte)
	room.Join(game.SubscriberMsg{Subscriber: subscriber, Player: player})

	log.Printf("Joined room %s with name %s and id %s", code, player.Name, player.ID)

	go server.subscriberListener(ws, subscriber)
	go server.socketListener(ws, room, subscriber)
}

// reads messages from socket and sends them to room
func (server *RoomsServer) socketListener(ws *websocket.Conn, room game.Broker, subscriber chan []byte) {
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
		_, buf, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Client closed connection with err %s", err.Error())
			return
		}
		// read any message from the socket and broadcast it to the room
		log.Println("Receiving message", string(buf))
		room.SendMessage(game.SentMsg{Message: buf, Sender: subscriber})
	}
}

// reads messages from a subscribed channel and sends them to socket
func (server *RoomsServer) subscriberListener(ws *websocket.Conn, subscriber chan []byte) {
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
	// perform the batch update stats in the background (ignoring the error)
	go func(results []game.GameResult) {
		_ = database.UpdateStats(server.db, results)
	}(results)
}

func (server RoomServer) DoCapture(snap game.Snapshot) {
	// perform a capture operation in the background (ignoring the error)
	go func(snap game.Snapshot) {
		_ = database.SaveSnapshot(server.db, snap)
	}(snap)
}

func (server RoomServer) OnTermination() {}
