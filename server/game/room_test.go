/*
 * Copyright (c) Joseph Prichard 2024
 */

package game

import (
	"fmt"
	"github.com/google/uuid"
	"sync"
	"testing"
	"time"
)

// no-op implementation of handler - we don't care about testing this
type FakeHandler struct{}

func (fake FakeHandler) DoShutdown(_ []GameResult) {}

func (fake FakeHandler) DoCapture(_ Snapshot) {}

func (fake FakeHandler) OnTermination() {}

// testing the message multiplexing and synchronization works as expected
func TestRoom(t *testing.T) {
	n := 9

	settings := MockSettings()
	settings.PlayerLimit = n

	initialState := NewGameState("123", settings)
	room := NewRoom(initialState, true, FakeHandler{})
	go room.Start()

	var wg sync.WaitGroup

	// testing n room joiners and n room leavers works as expected
	for i := 0; i < n; i++ {
		wg.Add(1)

		p := Player{
			ID:   uuid.New(),
			Name: fmt.Sprintf("Player %d", i),
		}
		subscriber := make(chan []byte)
		room.Join(SubscriberMsg{Subscriber: subscriber, Player: p})

		go func(i int) {
			for range subscriber {
			}
			t.Logf("Subscriber %d is finished", i)
			wg.Done()
		}(i)

		go func(i int) {
			room.SendMessage(SentMsg{
				Message: []byte(fmt.Sprintf("Testing 123 from %d", i)),
				Sender:  subscriber,
			})
			time.Sleep(time.Second * 3)
			room.Leave(subscriber)
		}(i)
	}

	wg.Wait()

	room.Stop(0)
}
