package server

import (
	"encoding/json"
	"fmt"
	"guessasketch/message"
	"guessasketch/utils"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type RoomCodeResp struct {
	Code string
}

type WsServerConfig struct {
	GameWordBank []string
	AuthServer   *AuthServer
	PlayerServer *PlayerServer
}

type WsServer struct {
	upgrader  websocket.Upgrader
	brokerage *message.Brokerage
	WsServerConfig
}

func NewWsServer(config WsServerConfig) *WsServer {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	return &WsServer{
		upgrader:       upgrader,
		brokerage:      message.NewBrokerMap(time.Minute),
		WsServerConfig: config,
	}
}

func (server *WsServer) CreateRoom(w http.ResponseWriter, r *http.Request) {
	utils.EnableCors(&w)

	// generate a code, create a broker, start it, then store it in the map
	code, err := utils.GenerateCode(8)
	if err != nil {
		resp := utils.ErrorMsg{Status: 500, ErrorDesc: "Failed to generate a valid error code"}
		utils.SendErrResp(w, resp)
		return
	}

	broker := message.NewBroker(code, server.GameWordBank)
	log.Printf("Starting a broker for code %s", code)
	go broker.Start()
	server.brokerage.Store(code, broker)

	roomCode := RoomCodeResp{Code: code}
	b, err := json.Marshal(roomCode)
	if err != nil {
		log.Printf("Failed to serialize create room code response")
		return
	}

	w.WriteHeader(200)
	w.Write(b)
}

func guestName() string {
	return fmt.Sprintf("Guest %d", 10+rand.Intn(89))
}

func (server *WsServer) JoinRoom(w http.ResponseWriter, r *http.Request) {
	utils.EnableCors(&w)

	session, err := server.AuthServer.GetSession(w, r)
	if err != nil {
		resp := utils.ErrorMsg{Status: 401, ErrorDesc: err.Error()}
		utils.SendErrResp(w, resp)
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
		resp := utils.ErrorMsg{Status: 400, ErrorDesc: "Player name must be 15 or less characters"}
		utils.SendErrResp(w, resp)
		return
	}

	broker := server.brokerage.Load(code)
	if broker == nil {
		resp := utils.ErrorMsg{Status: 404, ErrorDesc: "Cannot find room for provided code"}
		utils.SendErrResp(w, resp)
		return
	}

	ws, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
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
		ws.Close()
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
		ws.Close()
		log.Printf("Subscriber channel was closed")
		if panicInfo := recover(); panicInfo != nil {
			log.Println(panicInfo)
		}
	}()
	for resp := range subscriber {
		// read values from channel and write back to socket
		err := ws.WriteMessage(websocket.TextMessage, resp)
		if err != nil {
			log.Printf("Error writing message %s", err)
			return
		}
	}
}
