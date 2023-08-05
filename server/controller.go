package server

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type RoomCodeResp struct {
	Code string
}

type WsController struct {
	upgrader     websocket.Upgrader
	brokerMap    *BrokerMap
	gameWordBank []string
}

func NewWsController(gameWordBank []string) *WsController {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	return &WsController{
		upgrader:     upgrader,
		brokerMap:    NewBrokerMap(time.Minute),
		gameWordBank: gameWordBank,
	}
}

func generateCode(len int) (string, error) {
	b := make([]byte, len/2)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (controller *WsController) GetRandomCode(w http.ResponseWriter, r *http.Request) {
	code := controller.brokerMap.RandomCode()

	roomCode := RoomCodeResp{Code: code}
	b, err := json.Marshal(roomCode)
	if err != nil {
		log.Printf("Failed to serialize random room code response")
		return
	}

	w.WriteHeader(200)
	w.Write(b)
}

func (controller *WsController) CreateRoom(w http.ResponseWriter, r *http.Request) {
	// generate a code, create a broker, start it, then store it in the map
	code, err := generateCode(8)
	if err != nil {
		resp := ErrorMsg{Status: 500, ErrorDesc: "Failed to generate a valid error code"}
		SendErrResp(w, resp)
		return
	}

	broker := NewBroker(code, controller.gameWordBank)
	log.Printf("Starting a broker for code %s", code)
	go broker.Start()
	controller.brokerMap.Store(code, broker)

	roomCode := RoomCodeResp{Code: code}
	b, err := json.Marshal(roomCode)
	if err != nil {
		log.Printf("Failed to serialize create room code response")
		return
	}

	w.WriteHeader(200)
	w.Write(b)
}

func (controller *WsController) JoinRoom(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	code := query.Get("code")
	player := query.Get("name")

	if len(player) < 5 || len(player) > 15 {
		resp := ErrorMsg{Status: 400, ErrorDesc: "Player name must be between 5 and 15 characters"}
		SendErrResp(w, resp)
		return
	}

	broker := controller.brokerMap.Load(code)
	if broker == nil {
		resp := ErrorMsg{Status: 404, ErrorDesc: "Cannot find room for provided code"}
		SendErrResp(w, resp)
		return
	}

	ws, err := controller.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// create a new subscription channel and join the broker with it
	subscriber := make(chan string)
	broker.Subscribe <- SubscriberMsg{Subscriber: subscriber, Player: player}

	go subscriberListener(ws, subscriber)
	go socketListener(ws, broker, subscriber)
}

// reads messages from socket and sends them to broker
func socketListener(ws *websocket.Conn, broker *Broker, subscriber Subscriber) {
	defer func() {
		broker.Unsubscribe <- subscriber
		ws.Close()
		log.Printf("Socket listener close function called")
	}()
	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Client closed connection with err %s", err.Error())
			return
		}
		// read any message from the socket and broadcast it to the broker
		message := string(p)
		broker.SendMessage <- SentMsg{Message: message, Sender: subscriber}
	}
}

// reads messages from a subscribed channel and sends them to socket
func subscriberListener(ws *websocket.Conn, subscriber Subscriber) {
	defer func() {
		ws.Close()
		log.Printf("Subscriber channel was closed")
	}()
	for resp := range subscriber {
		// read values from channel and write back to socket
		err := ws.WriteMessage(websocket.TextMessage, []byte(resp))
		if err != nil {
			log.Printf("Error writing message %s", err)
			return
		}
	}
}
