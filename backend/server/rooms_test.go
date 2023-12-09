/*
 * Copyright (c) Joseph Prichard 2023
 */

package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"guessasketch/game"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// stub implementation of a rooms store that only stores a single room
type StubRoomsStore struct {
	code string
	room game.Room
}

func (stub *StubRoomsStore) Load(code string) game.Room {
	if stub.code == code {
		return stub.room
	}
	return nil
}

func (stub *StubRoomsStore) Store(code string, room game.Room) {
	stub.code = code
	stub.room = room
}

func (stub *StubRoomsStore) Codes(offset int, limit int) []string {
	return []string{stub.code}
}

// stub implementation of an authenticator where any user is not authenticated
type StubAuthenticator struct{}

func (stub *StubAuthenticator) GetSession(token string) (*JwtSession, error) {
	return nil, nil
}

// no-op implementation of worker - we don't care about testing this
type FakeWorker struct{}

func (fake FakeWorker) DoShutdown(results []game.GameResult) {}

func (fake FakeWorker) DoCapture(snap game.Snapshot) {}

func TestRoomsServer_CreateRoom(t *testing.T) {
	roomsServer := NewRoomsServer(&StubRoomsStore{}, &StubAuthenticator{}, &FakeWorker{}, []string{})

	testSettings := game.DefaultSettings()

	b, err := json.Marshal(testSettings)
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("Testing create room with body %s", string(b))
	body := strings.NewReader(string(b))

	r := httptest.NewRequest("", "/", body)
	w := httptest.NewRecorder()

	roomsServer.CreateRoom(w, r)

	resp := w.Result()
	t.Logf("Testing create room finished with %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		t.Fail()
	}
}

func BeforeTestJoinRoom(t *testing.T) (*httptest.Server, *websocket.Conn) {
	testCode := "123abc"
	initialState := game.NewGameState(testCode, game.DefaultSettings())

	testRoom := game.NewGameRoom(initialState, &FakeWorker{})
	mockRooms := StubRoomsStore{}
	go testRoom.Start()
	mockRooms.Store(testCode, testRoom)

	roomsServer := NewRoomsServer(&mockRooms, &StubAuthenticator{}, &FakeWorker{}, []string{})
	s := httptest.NewServer(http.HandlerFunc(roomsServer.JoinRoom))

	u := "ws" + strings.TrimPrefix(s.URL, "http") + "?code=" + testCode

	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// check the first two messages end when joining first
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Log(string(p))
	_, p, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Log(string(p))

	return s, ws
}

func TestRoomsServer_ChatMsg(t *testing.T) {
	s, ws := BeforeTestJoinRoom(t)
	defer s.Close()
	defer ws.Close()

	msg := fmt.Sprintf(`{"code": %d, "msg": {"text": "Hello 123 Hello123"}}`, game.TextCode)

	if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		t.Fatalf("%v", err)
	}
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}

	var payload game.OutputPayload
	_ = json.Unmarshal(p, &payload)
	t.Log(string(p))
}
