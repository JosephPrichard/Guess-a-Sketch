/*
 * Copyright (c) Joseph Prichard 2023
 */

package game

import (
	"encoding/json"
	"log"
	"sync/atomic"
	"time"
)

type Broker interface {
	Start()
	Join(m SubscriberMsg)
	Leave(s chan []byte)
	SendMessage(m SentMsg)
	Stop(c int)
	IsExpired(now time.Time) bool
	IsPublic() bool
}

type EventHandler interface {
	DoShutdown(results []GameResult)
	DoCapture(snap Snapshot)
	OnTermination()
}

type Room struct {
	join        chan SubscriberMsg
	leave       chan chan []byte
	sendMessage chan SentMsg
	reset       chan struct{}
	stop        chan int

	state       GameState
	subscribers map[chan []byte]Player
	expireTime  atomic.Int64
	isPublic    bool

	handler EventHandler
}

type SentMsg struct {
	Message []byte
	Sender  chan []byte
}

type SubscriberMsg struct {
	Subscriber chan []byte
	Player     Player
}

func NewRoom(initialState GameState, isPublic bool, handler EventHandler) *Room {
	room := &Room{
		join:        make(chan SubscriberMsg),
		leave:       make(chan chan []byte),
		sendMessage: make(chan SentMsg),
		reset:       make(chan struct{}),
		stop:        make(chan int),
		handler:     handler,
		subscribers: make(map[chan []byte]Player),
		state:       initialState,
		isPublic:    isPublic,
	}
	room.postponeExpiration()
	return room
}

func (room *Room) Start() {
	defer func() {
		log.Printf("Termination finished for room %s", room.state.code)
		if panicInfo := recover(); panicInfo != nil {
			log.Println("Fatal error in room: ", panicInfo)
		}
	}()
	for {
		select {
		case subMsg := <-room.join:
			room.onSubscribe(subMsg)
		case subscriber := <-room.leave:
			room.onUnsubscribe(subscriber)
		case sentMsg := <-room.sendMessage:
			room.onMessage(sentMsg)
		case <-room.reset:
			room.onResetState()
		case termCode := <-room.stop:
			room.onTerminate(termCode)
			room.handler.OnTermination()
			return
		}
	}
}

func (room *Room) Join(m SubscriberMsg) {
	room.join <- m
}

func (room *Room) Leave(s chan []byte) {
	room.leave <- s
}

func (room *Room) SendMessage(m SentMsg) {
	room.sendMessage <- m
}

func (room *Room) Stop(c int) {
	room.stop <- c
}

func (room *Room) IsExpired(now time.Time) bool {
	return now.Unix() >= room.expireTime.Load()
}

func (room *Room) IsPublic() bool {
	return room.isPublic
}

func (room *Room) postponeExpiration() {
	// set the expiration time for 15 minutes
	//room.expireTime.Store(time.Now().Unix() + 15*60)
	room.expireTime.Store(time.Now().Unix() + 10)
}

func (room *Room) startResetTimer(timeSecs int) {
	go func() {
		time.Sleep(time.Duration(timeSecs) * time.Second)
		room.reset <- struct{}{}
	}()
}

type ErrorMsg struct {
	ErrorDesc string `json:"errorDesc"`
}

func sendErrorMsg(ch chan []byte, errorDesc string) {
	e := ErrorMsg{ErrorDesc: errorDesc}
	buf, err := createResponse[ErrorMsg](ErrorCode, e)
	if err != nil {
		log.Println("Failed to serialize error for ws message")
		return
	}
	ch <- buf
}

func (room *Room) onSubscribe(subMsg SubscriberMsg) {
	resp, err := room.HandleJoin(subMsg.Player)
	if err != nil {
		log.Printf("User %v could not subscribe to the room", subMsg.Player)
		// only the sender should receive the error response
		sendErrorMsg(subMsg.Subscriber, err.Error())
		close(subMsg.Subscriber)
		return
	}
	log.Printf("User %v subscribed to the room", subMsg.Player)

	room.subscribers[subMsg.Subscriber] = subMsg.Player

	for s := range room.subscribers {
		s <- resp
	}

	// handle the initial message for the room only send to the subscriber
	resp, err = room.HandleState()
	if err != nil {
		// only the sender should receive the error response
		sendErrorMsg(subMsg.Subscriber, err.Error())
		close(subMsg.Subscriber)
		return
	}
	subMsg.Subscriber <- resp
}

func (room *Room) onUnsubscribe(subscriber chan []byte) {
	player := room.subscribers[subscriber]

	resp, err := HandleLeave(&room.state, player)
	if err != nil {
		// only the sender should receive the error response
		sendErrorMsg(subscriber, err.Error())
		return
	}

	delete(room.subscribers, subscriber)
	close(subscriber)

	for s := range room.subscribers {
		s <- resp
	}

	log.Println("User unsubscribed from the room")
}

func (room *Room) onMessage(sentMsg SentMsg) {
	// handle the message and get a response, then handle the error case
	player := room.subscribers[sentMsg.Sender]
	resp, err := room.HandleMessage(sentMsg.Message, player)
	if err != nil {
		// only the sender should receive the error response
		sendErrorMsg(sentMsg.Sender, err.Error())
		return
	}
	// broadcast a non error response to all subscribers
	if resp != nil {
		for s := range room.subscribers {
			s <- resp
		}
	}
}

func (room *Room) onResetState() {
	// reset the game and get a response, then handle the error case
	resp, err := room.HandleReset()
	if err != nil {
		// if an error does exist, serialize it and replace the success message with it
		e := ErrorMsg{ErrorDesc: err.Error()}
		buf, err := createResponse[ErrorMsg](ErrorCode, e)
		if err != nil {
			log.Println("Failed to serialize error for ws message")
			return
		}
		resp = buf
	}
	// broadcast the response to all subscribers - error or not
	for s := range room.subscribers {
		s <- resp
	}
	// check to handle the shutdown task
	if !room.state.HasMoreRounds() {
		room.handler.DoShutdown(room.state.CreateGameResults())
	}
}

func (room *Room) onTerminate(code int) {
	payload := OutputPayload[struct{}]{Code: code}
	resp, err := json.Marshal(payload)
	if err != nil {
		log.Println("Failed to serialize error for ws message")
		return
	}
	for s := range room.subscribers {
		s <- resp
	}
	// delete each subscriber from table and close channel
	for s := range room.subscribers {
		delete(room.subscribers, s)
		close(s)
	}
}
