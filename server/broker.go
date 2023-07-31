package server

import (
	"encoding/json"
	"guessasketch/store"
	"log"
	"time"
)

type Subscriber struct {
	Message    chan string
	Disconnect chan struct{}
}

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
	broker := &Broker{
		Subscribe:   make(chan SubscriberMsg),
		Unsubscribe: make(chan Subscriber),
		SendMessage: make(chan SentMsg),
		ResetState:  make(chan struct{}),
		Stop:        make(chan struct{}),
		subscribers: make(map[Subscriber]string),
	}
	broker.room = store.NewRoom(code, wordBank, broker.onStartEvent)
	return broker
}

func (broker *Broker) onStartEvent(settings store.RoomSettings) {
	// whenever a game starts, the timer to reset the game after the input delay must be set
	go func(timeSecs int) {
		time.Sleep(time.Duration(timeSecs) * time.Second)
		broker.ResetState <- struct{}{}
	}(settings.TimeLimitSecs)
}

func (broker *Broker) broadcast(resp string) {
	for s := range broker.subscribers {
		s.Message <- resp
	}
}

func (broker *Broker) onSubscribe(subMsg SubscriberMsg) {
	if !broker.room.CanJoin() {
		subMsg.Subscriber.Disconnect <- struct{}{}
		return
	}
	log.Printf("User subscribed to the broker")

	resp, err := HandleJoin(broker.room, subMsg.Player)
	if err != nil {
		// only the sender should receieve the error response
		SendErrMsg(subMsg.Subscriber.Message, err.Error())
		return
	}
	broker.subscribers[subMsg.Subscriber] = subMsg.Player

	broker.broadcast(resp)
	subMsg.Subscriber.Message <- broker.room.Marshal()
}

func (broker *Broker) onUnsubscribe(subscriber Subscriber) {
	log.Printf("User unsubscribed from the broker")

	player := broker.subscribers[subscriber]
	resp, err := HandleLeave(broker.room, player)
	if err != nil {
		// only the sender should receieve the error response
		SendErrMsg(subscriber.Message, err.Error())
		return
	}

	broker.broadcast(resp)
	delete(broker.subscribers, subscriber)
}

func (broker *Broker) onMessage(sentMsg SentMsg) {
	// handle the message and get a response, then handle the error case
	player := broker.subscribers[sentMsg.Sender]
	resp, err := HandleMessage(broker.room, sentMsg.Message, player)
	if err != nil {
		// only the sender should receieve the error response
		SendErrMsg(sentMsg.Sender.Message, err.Error())
		return
	}
	// broadcast a non error response to all subscribers
	broker.broadcast(resp)
}

func (broker *Broker) onResetState() {
	// reset the game and get a response, then handle the error cose
	resp, err := HandleReset(broker.room)
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
			return
		}
	}
}
