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

	broker := NewBroker()
	log.Printf("Starting a broker for code %s", code)
	go broker.Start(code, controller.gameWordBank)
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
	ch := make(Subscriber)
	subMsg := SubscriberMsg{Subscriber: ch, Player: player}
	broker.Subscribe <- subMsg

	ws.SetCloseHandler(func(int, string) error {
		broker.Unsubscribe <- ch
		close(ch)
		ws.Close()
		return nil
	})

	// create routines to read and write to broker
	go socketWriter(ws, ch)
	go socketReader(ws, broker, ch)
}

// reads messages from socket and sends them to the broker
func socketReader(ws *websocket.Conn, broker *Broker, ch Subscriber) {
	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		// read any message from the socket and broadcast it to the broker
		message := string(p)
		log.Printf("Read a message %s", message)
		broker.Broadcast <- BroadcastMsg{Message: message, Sender: ch}
	}
}

// recv messages from a subscribed channel and with to the socket
func socketWriter(ws *websocket.Conn, ch Subscriber) {
	for {
		// read values from channel and write back to socket
		resp := <-ch

		log.Printf("Reading a message on subscribed channel %s", resp)
		err := ws.WriteMessage(websocket.TextMessage, []byte(resp))
		if err != nil {
			log.Println(err)
			return
		}
	}
}
