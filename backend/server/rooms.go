package server

import (
	"encoding/json"
	"fmt"
	"guessasketch/game"
	"guessasketch/message"
	"guessasketch/utils"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type RoomCodeResp struct {
	Code     string
	Settings game.RoomSettings
}

type RoomsServerConfig struct {
	GameWordBank []string
	AuthServer   *AuthServer
	PlayerServer *PlayerServer
}

type RoomsServer struct {
	upgrader  websocket.Upgrader
	brokerage *message.Brokerage
	RoomsServerConfig
}

func NewRoomsServer(config RoomsServerConfig) *RoomsServer {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	return &RoomsServer{
		upgrader:          upgrader,
		brokerage:         message.NewBrokerMap(time.Minute),
		RoomsServerConfig: config,
	}
}

func (server *RoomsServer) Rooms(w http.ResponseWriter, _ *http.Request) {
	utils.EnableCors(&w)

	rooms := server.brokerage.Codes()
	w.WriteHeader(http.StatusOK)
	utils.WriteJson(w, rooms)
}

func (server *RoomsServer) CreateRoom(w http.ResponseWriter, r *http.Request) {
	utils.EnableCors(&w)

	// generate a code, create a broker, start it, then store it in the map
	code, err := utils.GenerateCode(8)
	if err != nil {
		resp := utils.ErrorResp{Status: http.StatusInternalServerError, ErrorDesc: "Failed to generate a valid error code"}
		utils.WriteError(w, resp)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		resp := utils.ErrorResp{Status: http.StatusInternalServerError, ErrorDesc: "Failed to read data from request body"}
		utils.WriteError(w, resp)
		return
	}

	var settings game.RoomSettings
	err = json.Unmarshal(body, &settings)
	if err != nil {
		resp := utils.ErrorResp{Status: http.StatusBadRequest, ErrorDesc: "Invalid format for settings"}
		utils.WriteError(w, resp)
		return
	}

	game.SettingsWithDefaults(&settings)
	settings.SharedWordBank = server.GameWordBank

	broker := message.NewBroker(code, settings)
	log.Printf("Starting a broker for code %s", code)
	go broker.Start()
	server.brokerage.Store(code, broker, settings.IsPublic)

	roomCode := RoomCodeResp{Code: code, Settings: settings}
	w.WriteHeader(http.StatusOK)
	utils.WriteJson(w, roomCode)
}

func guestName() string {
	return fmt.Sprintf("Guest %d", 10+rand.Intn(89))
}

func (server *RoomsServer) JoinRoom(w http.ResponseWriter, r *http.Request) {
	utils.EnableCors(&w)

	session, err := server.AuthServer.GetSession(w, r)
	if err != nil {
		resp := utils.ErrorResp{Status: http.StatusBadRequest, ErrorDesc: err.Error()}
		utils.WriteError(w, resp)
		return
	}
	id := session.ID

	query := r.URL.Query()
	code := query.Get("code")
	name := query.Get("name")

	if len(name) == 0 {
		name = guestName()
	}
	if len(name) > 15 {
		resp := utils.ErrorResp{Status: http.StatusBadRequest, ErrorDesc: "Player name must be 15 or less characters"}
		utils.WriteError(w, resp)
		return
	}

	broker := server.brokerage.Load(code)
	if broker == nil {
		resp := utils.ErrorResp{Status: http.StatusNotFound, ErrorDesc: "Cannot find room for provided code"}
		utils.WriteError(w, resp)
		return
	}

	ws, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		resp := utils.ErrorResp{Status: http.StatusInternalServerError, ErrorDesc: "Failed to upgrade to websocket"}
		utils.WriteError(w, resp)
		return
	}

	log.Printf("Joined room %s with name %s", code, name)

	// create a new subscription channel and join the broker with it
	player := message.Player{ID: id, Name: name}
	subscriber := make(message.Subscriber)
	broker.Subscribe <- message.SubscriberMsg{Subscriber: subscriber, Player: player}

	go subscriberListener(ws, subscriber)
	go socketListener(ws, broker, subscriber)
}

// reads messages from socket and sends them to broker
func socketListener(ws *websocket.Conn, broker *message.Broker, subscriber message.Subscriber) {
	defer func() {
		broker.Unsubscribe <- subscriber
		err := ws.Close()
		if err != nil {
			log.Println("Failed to close a websocket conn")
			return
		}
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
		// read any message from the socket and broadcast it to the broker
		broker.SendMessage <- message.SentMsg{Message: p, Sender: subscriber}
	}
}

// reads messages from a subscribed channel and sends them to socket
func subscriberListener(ws *websocket.Conn, subscriber message.Subscriber) {
	defer func() {
		err := ws.Close()
		if err != nil {
			log.Println("Failed to close a websocket conn")
			return
		}
		log.Println("Subscriber channel was closed")
		if panicInfo := recover(); panicInfo != nil {
			log.Println(panicInfo)
		}
	}()
	for resp := range subscriber {
		// read values from channel and write back to socket
		err := ws.WriteMessage(websocket.TextMessage, resp)
		if err != nil {
			log.Printf("WriteError writing message %s", err)
			return
		}
	}
}
