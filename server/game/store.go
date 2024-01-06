/*
 * Copyright (c) Joseph Prichard 2023
 */

package game

import (
	"log"
	"sync"
	"time"
)

type RoomsStore interface {
	Load(code string) Room
	Store(code string, room Room)
	Codes(offset int, limit int) []string
}

type RoomsMap struct {
	m     map[string]Room // maps codes to rooms
	codes []string        // stores the codes of the  rooms with older rooms last
	mu    sync.Mutex      // used to synchronize both structures
}

func NewRoomsMap(period time.Duration) *RoomsMap {
	rooms := &RoomsMap{
		m:     make(map[string]Room),
		codes: make([]string, 0),
	}
	go rooms.startCleanup(period)
	return rooms
}

func (m *RoomsMap) Load(code string) Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.m[code]
	if !ok || room.IsExpired(time.Now()) {
		return nil
	}
	return room
}

func (m *RoomsMap) Store(code string, room Room) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.m[code] = room
	// only add codes for public m into the list of all codes
	if room.IsPublic() {
		m.codes = append(m.codes, code)
	}
}

func (m *RoomsMap) Codes(offset int, limit int) []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	codes := make([]string, 0)

	upperLimit := offset + limit
	if len(m.codes) < upperLimit {
		upperLimit = len(m.codes)
	}

	for i := offset; i < upperLimit; i++ {
		c := m.codes[i]
		codes = append(codes, c)
	}
	return codes
}

func (m *RoomsMap) purgeExpired(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	expiredCodes := make(map[string]bool)
	for code, room := range m.m {
		// check if this room has expired already, if so delete it
		if room.IsExpired(now) {
			log.Printf("Deleting room for code %s", code)
			// terminate the room due to expiration with a timeout code
			room.Stop(TimeoutCode)
			delete(m.m, code)
			expiredCodes[code] = true
		}
	}
	// remove all expired codes from the slice
	for i := 0; i < len(m.codes); i++ {
		code := m.codes[i]
		_, expired := expiredCodes[code]
		if expired {
			if i < len(m.codes) {
				m.codes = append(m.codes[:i], m.codes[i+1:]...)
			} else {
				m.codes = m.codes[:i]
			}
			i--
		}
	}
}

func (m *RoomsMap) startCleanup(period time.Duration) {
	// periodically cleanup expired keys from the map
	for now := range time.NewTicker(period).C {
		m.purgeExpired(now)
	}
}
