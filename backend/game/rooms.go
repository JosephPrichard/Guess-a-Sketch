/*
 * Copyright (c) Joseph Prichard 2023
 */

package game

import (
	"log"
	"sync"
	"time"
)

type Rooms struct {
	m     map[string]*Room // maps codes to rooms
	codes []string         // stores the codes of the  rooms with older rooms last
	mu    sync.Mutex       // used to synchronize both structures
}

func NewRooms(period time.Duration) *Rooms {
	rooms := &Rooms{
		m:     make(map[string]*Room),
		codes: make([]string, 0),
	}
	go rooms.startCleanup(period)
	return rooms
}

func (rooms *Rooms) Load(code string) *Room {
	rooms.mu.Lock()
	defer rooms.mu.Unlock()

	room, ok := rooms.m[code]
	if !ok || room.IsExpired(time.Now()) {
		return nil
	}
	return room
}

func (rooms *Rooms) Store(code string, room *Room) {
	rooms.mu.Lock()
	defer rooms.mu.Unlock()

	rooms.m[code] = room
	// only add codes for public m into the list of all codes
	if room.IsPublic {
		rooms.codes = append(rooms.codes, code)
	}
}

func (rooms *Rooms) Codes(offset int, limit int) []string {
	rooms.mu.Lock()
	defer rooms.mu.Unlock()

	codes := make([]string, 0)

	upperLimit := offset + limit
	if len(rooms.codes) < upperLimit {
		upperLimit = len(rooms.codes)
	}

	for i := offset; i < upperLimit; i++ {
		c := rooms.codes[i]
		codes = append(codes, c)
	}
	return codes
}

func (rooms *Rooms) purgeExpired(now time.Time) {
	rooms.mu.Lock()
	defer rooms.mu.Unlock()

	expiredCodes := make(map[string]bool)
	for code, room := range rooms.m {
		// check if this room has expired already, if so delete it
		if room.IsExpired(now) {
			log.Printf("Deleting room for code %s", code)
			// terminate the room due to expiration with a timeout code
			room.Term <- TimeoutCode
			delete(rooms.m, code)
			expiredCodes[code] = true
		}
	}
	for i, code := range rooms.codes {
		_, expired := expiredCodes[code]
		if expired {
			rooms.codes = append(rooms.codes[:i], rooms.codes[i+1:]...)
		}
	}
}

func (rooms *Rooms) startCleanup(period time.Duration) {
	// periodically cleanup expired keys from the map
	for now := range time.NewTicker(period).C {
		rooms.purgeExpired(now)
	}
}
