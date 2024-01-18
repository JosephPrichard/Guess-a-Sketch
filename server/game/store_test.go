/*
 * Copyright (c) Joseph Prichard 2023
 */

package game

import (
	"reflect"
	"testing"
	"time"
)

type StubBroker struct {
	code      string
	stopCode  int
	isExpired bool
}

func (stub *StubBroker) Start() {}

func (stub *StubBroker) Join(_ SubscriberMsg) {}

func (stub *StubBroker) Leave(_ chan []byte) {}

func (stub *StubBroker) SendMessage(_ SentMsg) {}

func (stub *StubBroker) Stop(code int) {
	stub.stopCode = code
}

func (stub *StubBroker) IsExpired(_ time.Time) bool {
	return stub.isExpired
}

func (stub *StubBroker) IsPublic() bool {
	return true
}

func TestBrokerStore_StoreThenLoad(t *testing.T) {
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

	rooms := NewBrokerStore(time.Second)

	for _, test := range testStore {
		rooms.Set(test, &StubBroker{code: test})
	}
	for _, test := range testLoad {
		exists := rooms.Get(test.code) != nil
		if exists != test.expExists {
			t.Fatalf("Expected loaded room exists to be %t for code %s but got %t", test.expExists, test.code, exists)
		}
	}
}

func TestBrokerStore_StoreThenCodes(t *testing.T) {
	var testStore = []string{"123", "123", "456", "789"}

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

	store := NewBrokerStore(time.Second)

	for _, test := range testStore {
		store.Set(test, &StubBroker{code: test})
	}
	for i, test := range testCodes {
		codes := store.Codes(test.offset, test.limit)
		if !reflect.DeepEqual(codes, test.expCodes) {
			t.Fatalf("Expected codes to be equal for codes test %d", i)
		}
	}
}

func TestBrokerStore_Cleanup(t *testing.T) {
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

	store := NewBrokerStore(time.Second)

	var toStop []*StubBroker
	for _, test := range testStore {
		stub := &StubBroker{code: test.code, isExpired: test.isExpired}
		// only tests for expired store should be stopped
		if test.isExpired {
			toStop = append(toStop, stub)
		}
		store.Set(test.code, stub)
	}

	store.purgeExpired(time.Now())

	for i, r := range toStop {
		if r.stopCode != TimeoutCode {
			t.Fatalf("Expected stop code for input room %d to be timeout", i)
		}
	}

	codes := store.Codes(0, 10)
	if !reflect.DeepEqual(codes, []string{"999"}) {
		t.Fatalf("Expected codes to be an empty slice after purging all expired codes")
	}
}
