package server

import (
	"encoding/json"
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
}

func NewBroker() *Broker {
	return &Broker{
		Subscribe:   make(chan SubscriberMsg),
		Unsubscribe: make(chan Subscriber),
		SendMessage: make(chan SentMsg),
		ResetState:  make(chan struct{}),
		Stop:        make(chan struct{}),
	}
}

func (broker *Broker) Start(code string, wordBank []string) {
	// whenever a game starts, the timer to reset the game after the input delay must be set
	startResetTimer := func(timeSecs int) {
		time.Sleep(time.Duration(timeSecs) * time.Second)
		broker.ResetState <- struct{}{}
	}

	room := NewRoom(code, wordBank, startResetTimer)
	subscribers := make(map[chan string]string)

	// blocks this goroutine to listen to messages on each channel until told to stop
	for {
		select {
		case subMsg := <-broker.Subscribe:
			log.Printf("User subscribed to the broker")
			// add the channel to the map, send the room state, and handle join in the room state
			subscribers[subMsg.Subscriber] = subMsg.Player
			room.HandleJoin(subMsg.Player)
			subMsg.Subscriber <- room.Marshal()

		case subscriber := <-broker.Unsubscribe:
			log.Printf("User unsubscribed from the broker")
			// handle join in the room state, delete the subscriber channel from the map and close it
			player := subscribers[subscriber]
			room.HandleLeave(player)
			delete(subscribers, subscriber)

		case sentMsg := <-broker.SendMessage:
			// handle the message and get a response, then handle the error case
			player := subscribers[sentMsg.Sender]
			resp, err := room.HandleMessage(sentMsg.Message, player)
			if err != nil {
				// only the sender should receieve the error response
				SendErrMsg(sentMsg.Sender, err.Error())
				continue
			}
			// broadcast a non error response to all subscribers
			for s := range subscribers {
				s <- resp
			}

		case <-broker.ResetState:
			// reset the game and get a response, then handle the error cose
			resp, err := room.ResetGame()
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
			for s := range subscribers {
				s <- resp
			}

		case <-broker.Stop:
			return
		}
	}
}
