package message

import (
	"guessasketch/utils"
	"log"
	"sync"
	"time"
)

type Brokerage struct {
	m     map[string]*Broker // maps codes to brokers
	codes []string           // stores the codes of the brokers with older brokers last
	mu    sync.Mutex         // used to synchronize both structures
}

func NewBrokerMap(period time.Duration) *Brokerage {
	brokerage := &Brokerage{
		m:     make(map[string]*Broker),
		codes: make([]string, 0),
	}
	go brokerage.startCleanup(period)
	return brokerage
}

func (brokerage *Brokerage) Load(code string) *Broker {
	brokerage.mu.Lock()
	defer brokerage.mu.Unlock()

	broker, ok := brokerage.m[code]
	if !ok || broker.IsExpired(time.Now()) {
		return nil
	}
	return broker
}

func (brokerage *Brokerage) Store(code string, broker *Broker) {
	brokerage.mu.Lock()
	defer brokerage.mu.Unlock()

	brokerage.m[code] = broker
	// only add codes for public m into the list of all codes
	if broker.IsPublic {
		brokerage.codes = append(brokerage.codes, code)
	}
}

func (brokerage *Brokerage) Codes(offset int, limit int) []string {
	brokerage.mu.Lock()
	defer brokerage.mu.Unlock()

	codes := make([]string, 0)
	upperLimit := utils.Min(offset+limit, len(brokerage.codes))
	for i := offset; i < upperLimit; i++ {
		c := brokerage.codes[i]
		codes = append(codes, c)
	}
	return codes
}

func (brokerage *Brokerage) purgeExpired(now time.Time) {
	brokerage.mu.Lock()
	defer brokerage.mu.Unlock()

	expiredCodes := make(map[string]bool)
	for code, broker := range brokerage.m {
		// check if this broker has expired already, if so delete it
		if broker.IsExpired(now) {
			log.Printf("Deleting broker for code %s", code)
			broker.Stop <- struct{}{}
			delete(brokerage.m, code)
			expiredCodes[code] = true
		}
	}
	for i, code := range brokerage.codes {
		_, expired := expiredCodes[code]
		if expired {
			brokerage.codes = utils.Remove(brokerage.codes, i)
		}
	}
}

func (brokerage *Brokerage) startCleanup(period time.Duration) {
	// periodically cleanup expired keys from the map
	for now := range time.NewTicker(period).C {
		brokerage.purgeExpired(now)
	}
}
