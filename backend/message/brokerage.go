package message

import (
	"guessasketch/utils"
	"log"
	"sync"
	"time"
)

type Brokerage struct {
	// sync map to store the brokers themselves - allows for parallel iteration, etc
	brokers sync.Map
	// mutex and a slice to store list of codes - protected with a mutex
	mu    sync.Mutex
	codes []string
}

func NewBrokerMap(period time.Duration) *Brokerage {
	brokerage := &Brokerage{
		codes: make([]string, 0),
	}
	go brokerage.startCleanup(period)
	return brokerage
}

func (brokerage *Brokerage) delete(code string, broker *Broker) {
	log.Printf("Deleting broker with code %s", code)

	broker.Stop <- struct{}{}
	brokerage.brokers.Delete(code)

	// linear search to find a code to remove, then remove it
	brokerage.mu.Lock()
	for i, c := range brokerage.codes {
		if c == code {
			brokerage.codes = utils.Remove(brokerage.codes, i)
		}
	}
	brokerage.mu.Unlock()
}

func (brokerage *Brokerage) Load(code string) *Broker {
	v, ok := brokerage.brokers.Load(code)
	if !ok {
		return nil
	}
	broker := v.(*Broker)
	// check if this key has expired already
	if broker.IsExpired(time.Now()) {
		brokerage.delete(code, broker)
		return nil
	}
	return broker
}

func (brokerage *Brokerage) Store(code string, broker *Broker, isPublic bool) {
	brokerage.brokers.Store(code, broker)

	// only add codes for public brokers into the list of all codes
	if isPublic {
		brokerage.mu.Lock()
		brokerage.codes = append(brokerage.codes, code)
		brokerage.mu.Unlock()
	}
}

func (brokerage *Brokerage) Codes() []string {
	codes := make([]string, 0)
	brokerage.mu.Lock()
	for _, c := range brokerage.codes {
		codes = append(codes, c)
	}
	brokerage.mu.Unlock()
	return codes
}

func (brokerage *Brokerage) startCleanup(period time.Duration) {
	// periodically cleanup expired keys from the map
	for now := range time.NewTicker(period).C {
		brokerage.brokers.Range(func(key, value any) bool {
			broker := value.(*Broker)
			// check if this broker has expired already
			if broker.IsExpired(now) {
				brokerage.delete(key.(string), broker)
			}
			return true
		})
	}
}
