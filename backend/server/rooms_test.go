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

// mock implementation of a rooms store that only stores a single room
type MockRoomsStore struct {
	code string
	room *game.Room
}

func (store *MockRoomsStore) Load(code string) *game.Room {
	if store.code == code {
		return store.room
	}
	return nil
}

func (store *MockRoomsStore) Store(code string, room *game.Room) {
	store.code = code
	store.room = room
}

func (store *MockRoomsStore) Codes(_ int, _ int) []string {
	return []string{store.code}
}

// mock implementation of an authenticator where any user is not authenticated
type MockAuthenticator struct {
}

func (auth *MockAuthenticator) GetSession(_ string) (*JwtSession, error) {
	return nil, nil
}

// mock implementation of event hooks that does nothing - we don't care about testing this
type MockEventHandler struct {
}

func (server MockEventHandler) OnShutdown(_ []game.GameResult) {
}

func (server MockEventHandler) OnSaveDrawing(_ game.Drawing) error {
	return nil
}

func TestCreateRoom(t *testing.T) {
	roomsServer := NewRoomsServer(&MockRoomsStore{}, &MockAuthenticator{}, &MockEventHandler{}, []string{})

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

	testRoom := game.NewRoom(initialState, &MockEventHandler{})
	mockRooms := MockRoomsStore{}
	go testRoom.Start()
	mockRooms.Store(testCode, testRoom)

	roomsServer := NewRoomsServer(&mockRooms, &MockAuthenticator{}, &MockEventHandler{}, []string{})
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

func TestChatMsg(t *testing.T) {
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
	t.Log(string(p))
}
