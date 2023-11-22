package server

import (
	crand "crypto/rand"
	"encoding/hex"
	"guessasketch/game"
	"guessasketch/message"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type RoomCodeResp struct {
	Code     string            `json:"code"`
	Settings game.RoomSettings `json:"settings"`
}

type RoomsServer struct {
	upgrade      websocket.Upgrader
	brokerage    *message.Brokerage
	gameWordBank []string
	authServer   *AuthServer
	playerServer *PlayerServer
}

func NewRoomsServer(gameWordBank []string, authServer *AuthServer, playerServer *PlayerServer) *RoomsServer {
	upgrade := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	upgrade.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	return &RoomsServer{
		upgrade:      upgrade,
		brokerage:    message.NewBrokerMap(time.Minute),
		gameWordBank: gameWordBank,
		authServer:   authServer,
		playerServer: playerServer,
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

func (server *RoomsServer) CreateRoom(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)

	// generate a code, create a broker, start it, then store it in the map
	code, err := HexCode(8)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to generate a valid error code")
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

	broker := message.NewBroker(code, settings)
	log.Printf("Starting a broker for code %s", code)
	go broker.Start()
	server.brokerage.Store(code, broker)

	roomCode := RoomCodeResp{Code: code, Settings: settings}
	w.WriteHeader(http.StatusOK)
	WriteJson(w, roomCode)
}

func (server *RoomsServer) JoinRoom(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)

	query := r.URL.Query()
	code := query.Get("code")
	token := query.Get("token")

	var player User
	if token != "" {
		// if a session token is specified, attempt to get the id for the user
		session, err := server.authServer.GetSession(token)
		if err != nil && session != nil {
			player = session.user
		}
	} else {
		player = GuestUser()
	}

	broker := server.brokerage.Load(code)
	if broker == nil {
		WriteError(w, http.StatusNotFound, "Cannot find room for provided code")
		return
	}

	ws, err := server.upgrade.Upgrade(w, r, nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to upgrade to websocket")
		return
	}

	// create a new subscription channel and join the broker with it
	subscriber := make(message.Subscriber)
	broker.Subscribe <- message.SubscriberMsg{Subscriber: subscriber, Player: player}

	log.Printf("Joined room %s with name %s and id %s", code, player.Name, player.ID)

	go subscriberListener(ws, subscriber)
	go socketListener(ws, broker, subscriber)
}

// reads messages from socket and sends them to broker
func socketListener(ws *websocket.Conn, broker *message.Broker, subscriber message.Subscriber) {
	defer func() {
		// unsubscribes from the broker when the websocket is closed
		broker.Unsubscribe <- subscriber
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
		// read any message from the socket and broadcast it to the broker
		log.Println("Receiving message", string(p))
		broker.SendMessage <- message.SentMsg{Message: p, Sender: subscriber}
	}
}

// reads messages from a subscribed channel and sends them to socket
func subscriberListener(ws *websocket.Conn, subscriber message.Subscriber) {
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
