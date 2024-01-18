/*
 * Copyright (c) Joseph Prichard 2023
 */

package game

import (
	"log"
	"sync"
	"time"
)

type Brokerage interface {
	Get(code string) Broker
	Set(code string, b Broker)
	Codes(offset int, limit int) []string
}

type BrokerStore struct {
	m     map[string]Broker // maps codes to brokers
	codes []string          // stores the codes of the older brokers last
	mu    sync.Mutex        // used to synchronize both structures
}

func NewBrokerStore(period time.Duration) *BrokerStore {
	rooms := &BrokerStore{
		m:     make(map[string]Broker),
		codes: make([]string, 0),
	}
	go rooms.startCleanup(period)
	return rooms
}

func (store *BrokerStore) Get(code string) Broker {
	store.mu.Lock()
	defer store.mu.Unlock()

	room, ok := store.m[code]
	if !ok || room.IsExpired(time.Now()) {
		return nil
	}
	return room
}

func (store *BrokerStore) Set(code string, b Broker) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.m[code] = b
	// only add codes for public broker into the list of all codes
	if b.IsPublic() {
		store.codes = append(store.codes, code)
	}
}

func (store *BrokerStore) Codes(offset int, limit int) []string {
	store.mu.Lock()
	defer store.mu.Unlock()

	codes := make([]string, 0)

	upperLimit := offset + limit
	if len(store.codes) < upperLimit {
		upperLimit = len(store.codes)
	}

	for i := offset; i < upperLimit; i++ {
		c := store.codes[i]
		codes = append(codes, c)
	}
	return codes
}

func (store *BrokerStore) purgeExpired(now time.Time) {
	store.mu.Lock()
	defer store.mu.Unlock()

	expiredCodes := make(map[string]bool)
	for code, room := range store.m {
		// check if this room has expired already, if so delete it
		if room.IsExpired(now) {
			log.Printf("Deleting room for code %s", code)
			// terminate the room due to expiration with a timeout code
			room.Stop(TimeoutCode)
			delete(store.m, code)
			expiredCodes[code] = true
		}
	}
	// remove all expired codes from the slice
	for i := 0; i < len(store.codes); i++ {
		code := store.codes[i]
		_, expired := expiredCodes[code]
		if expired {
			if i < len(store.codes) {
				store.codes = append(store.codes[:i], store.codes[i+1:]...)
			} else {
				store.codes = store.codes[:i]
			}
			i--
		}
	}
}

func (store *BrokerStore) startCleanup(period time.Duration) {
	// periodically cleanup expired keys from the map
	for now := range time.NewTicker(period).C {
		store.purgeExpired(now)
	}
}
