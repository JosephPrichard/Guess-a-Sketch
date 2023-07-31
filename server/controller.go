package server

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WsController struct {
	upgrader     websocket.Upgrader
	brokers      sync.Map
	gameWordBank []string
}

type RoomCodeResp struct {
	Code string
}

func NewWsController(gameWordBank []string) *WsController {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	return &WsController{upgrader: upgrader, gameWordBank: gameWordBank}
}

func generateCode(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (controller *WsController) CreateRoom(w http.ResponseWriter, r *http.Request) {
	// generate a code, create a broker, start it, then store it in the map
	code, err := generateCode(4)
	if err != nil {
		resp := ErrorMsg{Status: 500, ErrorDesc: "Failed to generate a valid error code"}
		SendErrResp(w, resp)
		return
	}

	broker := NewBroker(code, controller.gameWordBank)
	log.Printf("Starting a broker for code %s", code)
	go broker.Start()
	controller.brokers.Store(code, broker)

	roomCode := RoomCodeResp{Code: code}
	b, err := json.Marshal(roomCode)
	if err != nil {
		log.Printf("Failed to serialize room code response")
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

	ws, err := controller.upgrader.Upgrade(w, r, nil)
	// close socket on cleanup
	defer ws.Close()
	if err != nil {
		log.Println(err)
	}

	var broker *Broker
	v, ok := controller.brokers.Load(code)
	if ok {
		broker = v.(*Broker)
	} else {
		resp := ErrorMsg{Status: 404, ErrorDesc: "Cannot find room for provided code"}
		SendErrResp(w, resp)
		return
	}

	// create a new subscription channel and join the broker with it
	subscriber := Subscriber{
		Message:    make(chan string),
		Disconnect: make(chan struct{}),
	}
	broker.Subscribe <- SubscriberMsg{Subscriber: subscriber, Player: player}

	ws.SetCloseHandler(func(code int, text string) error {
		broker.Unsubscribe <- subscriber
		close(subscriber.Message)
		close(subscriber.Disconnect)
		log.Printf("Client closed connection with code %d and message %s", code, text)
		return nil
	})

	go socketWriter(ws, subscriber)
	go socketReader(ws, broker, subscriber)

	// wait and close connection if a disconnect signal is sent
	<-subscriber.Disconnect
}

// reads messages from socket and sends them to the broker
func socketReader(ws *websocket.Conn, broker *Broker, subscriber Subscriber) {
	// close socket on cleanup
	defer ws.Close()
	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			// stop reading the socket on error, let close handler handle the error
			return
		}
		// read any message from the socket and broadcast it to the broker
		message := string(p)
		broker.SendMessage <- SentMsg{Message: message, Sender: subscriber}
	}
}

// recv messages from a subscribed channel and with to the socket
func socketWriter(ws *websocket.Conn, subscriber Subscriber) {
	for resp := range subscriber.Message {
		// read values from channel and write back to socket
		err := ws.WriteMessage(websocket.TextMessage, []byte(resp))
		if err != nil {
			log.Printf("Error writing message %s", err)
			return
		}
	}
	log.Printf("Subscriber channel was closed")
}
