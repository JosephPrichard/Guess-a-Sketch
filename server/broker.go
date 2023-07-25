package server

import "log"

type Subscriber = chan string

type BroadcastMsg struct {
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
	Broadcast   chan BroadcastMsg
	Stop        chan struct{}
}

func NewBroker() *Broker {
	return &Broker{
		Subscribe:   make(chan SubscriberMsg),
		Unsubscribe: make(chan Subscriber),
		Broadcast:   make(chan BroadcastMsg),
		Stop:        make(chan struct{}),
	}
}

func (broker *Broker) Start(code string, wordBank []string) {
	room := NewRoom(code, wordBank)
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

		case broadcastMsg := <-broker.Broadcast:
			// handle the message and get a response, then handle the error case
			player := subscribers[broadcastMsg.Sender]
			resp, err := room.HandleMessage(broadcastMsg.Message, player)
			if err != nil {
				SendErrMsg(broadcastMsg.Sender, err.Error())
				continue
			}
			for s := range subscribers {
				s <- resp
			}

		case <-broker.Stop:
			return
		}
	}
}
