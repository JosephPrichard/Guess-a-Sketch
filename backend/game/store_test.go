/*
 * Copyright (c) Joseph Prichard 2023
 */

package game

import (
	"reflect"
	"testing"
	"time"
)

type StubRoom struct {
	code      string
	stopCode  int
	isExpired bool
}

func (stub *StubRoom) Start() {}

func (stub *StubRoom) Join(m SubscriberMsg) {}

func (stub *StubRoom) Leave(s Subscriber) {}

func (stub *StubRoom) SendMessage(m SentMsg) {}

func (stub *StubRoom) Stop(code int) {
	stub.stopCode = code
}

func (stub *StubRoom) IsExpired(now time.Time) bool {
	return stub.isExpired
}

func (stub *StubRoom) IsPublic() bool {
	return true
}

func TestRoomsMap_StoreThenLoad(t *testing.T) {
	var testStore = []string{"123", "123", "456", "789"}

	type TestLoad struct {
		code      string
		expExists bool
	}
	testLoad := []TestLoad{
		{code: "123", expExists: true},
		{code: "456", expExists: true},
		{code: "789", expExists: true},
		{code: "555", expExists: false},
	}

	rooms := NewRoomsMap(time.Second)

	for _, test := range testStore {
		rooms.Store(test, &StubRoom{code: test})
	}
	for _, test := range testLoad {
		exists := rooms.Load(test.code) != nil
		if exists != test.expExists {
			t.Fatalf("Expected loaded room exists to be %t for code %s but got %t", test.expExists, test.code, exists)
		}
	}
}

func TestRoomsMap_StoreThenCodes(t *testing.T) {
	var testStore = []string{"123", "123", "456", "789"}

	rooms := NewRoomsMap(time.Second)

	type TestCodes struct {
		offset   int
		limit    int
		expCodes []string
	}
	testCodes := []TestCodes{
		{offset: 0, limit: 10, expCodes: []string{"123", "123", "456", "789"}},
		{offset: 2, limit: 10, expCodes: []string{"456", "789"}},
		{offset: 0, limit: 2, expCodes: []string{"123", "123"}},
	}

	for _, test := range testStore {
		rooms.Store(test, &StubRoom{code: test})
	}
	for i, test := range testCodes {
		codes := rooms.Codes(test.offset, test.limit)
		if !reflect.DeepEqual(codes, test.expCodes) {
			t.Fatalf("Expected codes to be equal for codes test %d", i)
		}
	}
}

func TestRoomsMap_Cleanup(t *testing.T) {
	type TestStore struct {
		code      string
		isExpired bool
	}
	var testStore = []TestStore{
		{code: "123", isExpired: true},
		{code: "456", isExpired: true},
		{code: "789", isExpired: true},
		{code: "999", isExpired: false},
	}

	rooms := NewRoomsMap(time.Second)

	var toStop []*StubRoom
	for _, test := range testStore {
		stub := &StubRoom{code: test.code, isExpired: test.isExpired}
		// only tests for expired rooms should be stopped
		if test.isExpired {
			toStop = append(toStop, stub)
		}
		rooms.Store(test.code, stub)
	}

	rooms.purgeExpired(time.Now())

	for i, r := range toStop {
		if r.stopCode != TimeoutCode {
			t.Fatalf("Expected stop code for input room %d to be timeout", i)
		}
	}

	codes := rooms.Codes(0, 10)
	if !reflect.DeepEqual(codes, []string{"999"}) {
		t.Fatalf("Expected codes to be an empty slice after purging all expired codes")
	}
}
