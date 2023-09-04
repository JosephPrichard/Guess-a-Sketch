package message

import (
	"sync"
	"time"
)

type BrokerMap struct {
	m sync.Map
}

func NewBrokerMap(period time.Duration) *BrokerMap {
	brokerMap := &BrokerMap{}
	go brokerMap.startCleanup(period)
	return brokerMap
}

func (brokerMap *BrokerMap) delete(code string, broker *Broker) {
	broker.Stop <- struct{}{}
	brokerMap.m.Delete(code)
}

func (brokerMap *BrokerMap) Load(code string) *Broker {
	v, ok := brokerMap.m.Load(code)
	if !ok {
		return nil
	}
	broker := v.(*Broker)
	// check if this key has expired already
	if broker.IsExpired(time.Now()) {
		brokerMap.delete(code, broker)
		return nil
	}
	return broker
}

func (brokerMap *BrokerMap) Store(code string, broker *Broker) {
	brokerMap.m.Store(code, broker)
}

func (brokerMap *BrokerMap) startCleanup(period time.Duration) {
	// periodically cleanup expired keys from the map
	for now := range time.NewTicker(period).C {
		brokerMap.m.Range(func(key, value any) bool {
			broker := value.(*Broker)
			// check if this key has expired already
			if broker.IsExpired(now) {
				brokerMap.delete(key.(string), broker)
			}
			return true
		})
	}
}