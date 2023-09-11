package message

import (
	"encoding/json"
	"guessasketch/game"
	"guessasketch/utils"
	"log"
	"sync/atomic"
	"time"
)

type Subscriber = chan []byte

type Player = game.Player

type SentMsg struct {
	Message []byte
	Sender  Subscriber
}

type SubscriberMsg struct {
	Subscriber Subscriber
	Player     Player
}

type Broker struct {
	Subscribe   chan SubscriberMsg
	Unsubscribe chan Subscriber
	SendMessage chan SentMsg
	ResetState  chan struct{}
	Stop        chan struct{}
	room        game.Room
	subscribers map[Subscriber]Player
	ExpireTime  atomic.Int64
}

func NewBroker(code string, wordBank []string) *Broker {
	// create the broker with all channels and state
	broker := &Broker{
		Subscribe:   make(chan SubscriberMsg),
		Unsubscribe: make(chan Subscriber),
		SendMessage: make(chan SentMsg),
		ResetState:  make(chan struct{}),
		Stop:        make(chan struct{}),
		subscribers: make(map[Subscriber]Player),
		room:        game.NewRoom(code, wordBank),
	}
	broker.PostponeExpiration()
	return broker
}

func (room *Broker) PostponeExpiration() {
	// set the expiration time for 15 minutes
	room.ExpireTime.Store(time.Now().Unix() + 15*60)
}

func (broker *Broker) IsExpired(now time.Time) bool {
	return now.Unix() > broker.ExpireTime.Load()
}

func (broker *Broker) StartResetTimer(timeSecs int) {
	go func() {
		time.Sleep(time.Duration(timeSecs) * time.Second)
		broker.ResetState <- struct{}{}
	}()
}

func (broker *Broker) broadcast(resp []byte) {
	for s := range broker.subscribers {
		s <- resp
	}
}

func (broker *Broker) onSubscribe(subMsg SubscriberMsg) {
	resp, err := HandleJoin(&broker.room, subMsg.Player)
	if err != nil {
		// only the sender should receieve the error response
		SendErrMsg(subMsg.Subscriber, err.Error())
		close(subMsg.Subscriber)
		return
	}
	log.Printf("User subscribed to the broker")

	broker.subscribers[subMsg.Subscriber] = subMsg.Player

	broker.broadcast(resp)
	subMsg.Subscriber <- broker.room.ToMessage()
}

func (broker *Broker) onUnsubscribe(subscriber Subscriber) {
	player := broker.subscribers[subscriber]
	resp, err := HandleLeave(&broker.room, player)
	if err != nil {
		// only the sender should receieve the error response
		SendErrMsg(subscriber, err.Error())
		return
	}
	log.Printf("User unsubscribed from the broker")

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
		errMsg := utils.ErrorMsg{ErrorDesc: err.Error()}
		b, err := json.Marshal(errMsg)
		if err != nil {
			log.Printf("Failed to serialize error for ws message")
			return
		}
		resp = b
	}
	// broadcast the response to all subscribers - error or not
	broker.broadcast(resp)
}

func (broker *Broker) onStop() {
	resp, err := HandleTimeoutMessage()
	if err != nil {
		log.Printf("Failed to serialize error for ws message")
		return
	}
	broker.broadcast(resp)
	// delete each subscriber from table and close channel
	for s := range broker.subscribers {
		delete(broker.subscribers, s)
		close(s)
	}
}

func (broker *Broker) Start() {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			log.Println(panicInfo)
		}
	}()
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
