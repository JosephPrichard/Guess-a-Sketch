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

// represent a room that a client can send messages to, subscribe to, and hook onto internal events
type Room struct {
	Subscribe   chan SubscriberMsg
	Unsubscribe chan Subscriber
	SendMessage chan SentMsg
	ResetState  chan struct{}
	Term        chan int
	events      EventHandler
	state       GameState
	subscribers map[Subscriber]Player
	expireTime  atomic.Int64
	IsPublic    bool
}

// messaging layer events called that must be implemented by the caller
type EventHandler interface {
	OnShutdown(results []GameResult)
	OnSaveDrawing(drawing Drawing) error
}

type Subscriber = chan []byte

type SentMsg struct {
	Message []byte
	Sender  Subscriber
}

type SubscriberMsg struct {
	Subscriber Subscriber
	Player     Player
}

func NewRoom(code string, settings RoomSettings, events EventHandler) *Room {
	// create the room with all channels and state
	room := &Room{
		Subscribe:   make(chan SubscriberMsg),
		Unsubscribe: make(chan Subscriber),
		SendMessage: make(chan SentMsg),
		ResetState:  make(chan struct{}),
		Term:        make(chan int),
		events:      events,
		subscribers: make(map[Subscriber]Player),
		state:       NewGameRoom(code, settings),
		IsPublic:    settings.IsPublic,
	}
	room.PostponeExpiration()
	return room
}

func (room *Room) PostponeExpiration() {
	// set the expiration time for 15 minutes
	room.expireTime.Store(time.Now().Unix() + 15*60)
}

func (room *Room) IsExpired(now time.Time) bool {
	return now.Unix() >= room.expireTime.Load()
}

func (room *Room) StartResetTimer(timeSecs int) {
	go func() {
		time.Sleep(time.Duration(timeSecs) * time.Second)
		room.ResetState <- struct{}{}
	}()
}

func (room *Room) broadcast(resp []byte) {
	for s := range room.subscribers {
		s <- resp
	}
}

type ErrorMsg struct {
	ErrorDesc string `json:"errorDesc"`
}

func SendErrorMsg(ch chan []byte, errorDesc string) {
	msg := ErrorMsg{ErrorDesc: errorDesc}
	b, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to serialize error for ws message")
		return
	}
	ch <- b
}

func (room *Room) onSubscribe(subMsg SubscriberMsg) {
	resp, err := HandleJoin(&room.state, subMsg.Player)
	if err != nil {
		// only the sender should receive the error response
		SendErrorMsg(subMsg.Subscriber, err.Error())
		close(subMsg.Subscriber)
		return
	}
	log.Println("User subscribed to the room")

	room.subscribers[subMsg.Subscriber] = subMsg.Player

	room.broadcast(resp)
	subMsg.Subscriber <- room.state.ToMessage()
}

func (room *Room) onUnsubscribe(subscriber Subscriber) {
	player := room.subscribers[subscriber]
	resp, err := HandleLeave(&room.state, player)
	if err != nil {
		// only the sender should receive the error response
		SendErrorMsg(subscriber, err.Error())
		return
	}
	log.Println("User unsubscribed from the room")

	delete(room.subscribers, subscriber)
	close(subscriber)

	room.broadcast(resp)
}

func (room *Room) onMessage(sentMsg SentMsg) {
	// handle the message and get a response, then handle the error case
	player := room.subscribers[sentMsg.Sender]
	resp, err := HandleMessage(room, sentMsg.Message, player)
	if err != nil {
		// only the sender should receive the error response
		SendErrorMsg(sentMsg.Sender, err.Error())
		return
	}
	// broadcast a non error response to all subscribers
	if resp != nil {
		room.broadcast(resp)
	}
}

func (room *Room) onResetState() {
	// reset the game and get a response, then handle the error case
	resp, err := HandleReset(room)
	if err != nil {
		// if an error does exist, serialize it and replace the success message with it
		errMsg := ErrorMsg{ErrorDesc: err.Error()}
		b, err := json.Marshal(errMsg)
		if err != nil {
			log.Println("Failed to serialize error for ws message")
			return
		}
		resp = b
	}
	// broadcast the response to all subscribers - error or not
	room.broadcast(resp)
	// check to handle the shutdown event
	if !room.state.HasMoreRounds() {
		room.events.OnShutdown(room.state.CreateGameResult())
	}
}

func (room *Room) onTerminate(code int) {
	payload := OutputPayload{Code: code}
	resp, err := json.Marshal(payload)
	if err != nil {
		err = ErrMarshal
	}
	if err != nil {
		log.Println("Failed to serialize error for ws message")
		return
	}
	room.broadcast(resp)
	// delete each subscriber from table and close channel
	for s := range room.subscribers {
		delete(room.subscribers, s)
		close(s)
	}
}

func (room *Room) Start() {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			log.Println(panicInfo)
		}
	}()
	for {
		select {
		case subMsg := <-room.Subscribe:
			room.onSubscribe(subMsg)
		case subscriber := <-room.Unsubscribe:
			room.onUnsubscribe(subscriber)
		case sentMsg := <-room.SendMessage:
			room.onMessage(sentMsg)
		case <-room.ResetState:
			room.onResetState()
		case termCode := <-room.Term:
			room.onTerminate(termCode)
			return
		}
	}
}
