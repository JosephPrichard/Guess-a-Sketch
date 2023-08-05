package server

import (
	"encoding/json"
	"guessasketch/store"
	"log"
	"time"
)

type Subscriber = chan string

type SentMsg struct {
	Message string
	Sender  Subscriber
}

type SubscriberMsg struct {
	Subscriber Subscriber
	Player     string
}

type Broker struct {
	Subscribe   chan SubscriberMsg
	Unsubscribe chan Subscriber
	SendMessage chan SentMsg
	ResetState  chan struct{}
	Stop        chan struct{}
	room        *store.Room
	subscribers map[Subscriber]string
}

func NewBroker(code string, wordBank []string) *Broker {
	// create the broker with all channels and state
	return &Broker{
		Subscribe:   make(chan SubscriberMsg),
		Unsubscribe: make(chan Subscriber),
		SendMessage: make(chan SentMsg),
		ResetState:  make(chan struct{}),
		Stop:        make(chan struct{}),
		subscribers: make(map[Subscriber]string),
		room:        store.NewRoom(code, wordBank),
	}
}

func (broker *Broker) IsExpired(now time.Time) bool {
	return now.Second() > int(broker.room.ExpireTime.Load())
}

func (broker *Broker) startResetTimer(timeSecs int) {
	go func() {
		time.Sleep(time.Duration(timeSecs) * time.Second)
		broker.ResetState <- struct{}{}
	}()
}

func (broker *Broker) broadcast(resp string) {
	for s := range broker.subscribers {
		s <- resp
	}
}

func (broker *Broker) onSubscribe(subMsg SubscriberMsg) {
	if !broker.room.CanJoin() {
		close(subMsg.Subscriber)
		return
	}
	log.Printf("User subscribed to the broker")

	resp, err := HandleJoin(broker.room, subMsg.Player)
	if err != nil {
		// only the sender should receieve the error response
		SendErrMsg(subMsg.Subscriber, err.Error())
		return
	}
	broker.subscribers[subMsg.Subscriber] = subMsg.Player

	broker.broadcast(resp)
	subMsg.Subscriber <- broker.room.Marshal()
}

func (broker *Broker) onUnsubscribe(subscriber Subscriber) {
	log.Printf("User unsubscribed from the broker")

	player := broker.subscribers[subscriber]
	resp, err := HandleLeave(broker.room, player)
	if err != nil {
		// only the sender should receieve the error response
		SendErrMsg(subscriber, err.Error())
		return
	}
	delete(broker.subscribers, subscriber)
	close(subscriber)

	broker.broadcast(resp)
}

func (broker *Broker) onMessage(sentMsg SentMsg) {
	// handle the message and get a response, then handle the error case
	player := broker.subscribers[sentMsg.Sender]
	resp, err := HandleMessage(broker, sentMsg.Message, player)
	if err != nil {
		// only the sender should receieve the error response
		SendErrMsg(sentMsg.Sender, err.Error())
		return
	}
	// broadcast a non error response to all subscribers
	broker.broadcast(resp)
}

func (broker *Broker) onResetState() {
	// reset the game and get a response, then handle the error cose
	resp, err := HandleReset(broker)
	if err != nil {
		// if an error does exist, serialize it and replace the success message with it
		errMsg := ErrorMsg{ErrorDesc: err.Error()}
		b, err := json.Marshal(errMsg)
		if err != nil {
			log.Printf("Failed to serialize error for ws message")
			return
		}
		resp = string(b)
	}
	// broadcast the response to all subscribers - error or not
	broker.broadcast(resp)
}

func (broker *Broker) onStop() {
	for s := range broker.subscribers {
		delete(broker.subscribers, s)
		close(s)
	}
}

func (broker *Broker) Start() {
	for {
		select {
		case subMsg := <-broker.Subscribe:
			broker.onSubscribe(subMsg)
		case subscriber := <-broker.Unsubscribe:
			broker.onUnsubscribe(subscriber)
		case sentMsg := <-broker.SendMessage:
			broker.onMessage(sentMsg)
		case <-broker.ResetState:
			broker.onResetState()
		case <-broker.Stop:
			broker.onStop()
			return
		}
	}
}
