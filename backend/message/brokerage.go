package message

import (
	"sync"
	"time"
)

type Brokerage struct {
	m sync.Map
}

func NewBrokerMap(period time.Duration) *Brokerage {
	brokerage := &Brokerage{}
	go brokerage.startCleanup(period)
	return brokerage
}

func (brokerage *Brokerage) delete(code string, broker *Broker) {
	broker.Stop <- struct{}{}
	brokerage.m.Delete(code)
}

func (brokerage *Brokerage) Load(code string) *Broker {
	v, ok := brokerage.m.Load(code)
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

func (brokerage *Brokerage) Store(code string, broker *Broker) {
	brokerage.m.Store(code, broker)
}

func (brokerage *Brokerage) startCleanup(period time.Duration) {
	// periodically cleanup expired keys from the map
	for now := range time.NewTicker(period).C {
		brokerage.m.Range(func(key, value any) bool {
			broker := value.(*Broker)
			// check if this key has expired already
			if broker.IsExpired(now) {
				brokerage.delete(key.(string), broker)
			}
			return true
		})
	}
}
