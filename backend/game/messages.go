/*
 * Copyright (c) Joseph Prichard 2023
 */

package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

const (
	StartCode   = 1
	TextCode    = 2
	DrawCode    = 3
	ChatCode    = 4
	FinishCode  = 5
	BeginCode   = 6
	JoinCode    = 7
	LeaveCode   = 8
	TimeoutCode = 9
	SaveCode    = 10
	MinChatLen  = 5
	MaxChatLen  = 50
)

var ErrUnMarshal = errors.New("Failed to unmarshal input data")
var ErrMarshal = errors.New("Failed to marshal output data")

type InputPayload struct {
	Code int             `json:"code"`
	Msg  json.RawMessage `json:"msg"`
}

type OutputPayload struct {
	Code int         `json:"code"`
	Msg  interface{} `json:"msg"`
}

func HandleMessage(room *Room, message []byte, player Player) ([]byte, error) {
	// deserialize payload message from json
	var payload InputPayload
	err := json.Unmarshal(message, &payload)
	if err != nil {
		return nil, err
	}
	log.Printf("Handling message code %d", payload.Code)

	switch payload.Code {
	case StartCode:
		return handleStartMessage(room, player)
	case TextCode:
		var inputMsg TextMsg
		err = json.Unmarshal(payload.Msg, &inputMsg)
		if err != nil {
			return nil, ErrUnMarshal
		}
		return handleTextMessage(&room.state, inputMsg, player)
	case DrawCode:
		var inputMsg DrawMsg
		err = json.Unmarshal(payload.Msg, &inputMsg)
		if err != nil {
			return nil, ErrUnMarshal
		}
		return handleDrawMessage(&room.state, inputMsg, player)
	case SaveCode:
		capture, err := room.state.Capture()
		if err != nil {
			return nil, err
		}
		capture.SavedBy = &player
		room.worker.DoCapture(capture)
		return nil, nil
	default:
		log.Println("Cannot handle unknown message type")
		return nil, errors.New("No matching message types for message")
	}
}

type BeginMsg struct {
	NextWord        string `json:"nextWord"`
	NextPlayerIndex int    `json:"nextPlayerIndex"`
}

func handleStartMessage(room *Room, player Player) ([]byte, error) {
	state := &room.state

	if state.PlayerIsNotHost(player) {
		return nil, errors.New("Player must be the host to start the game")
	}
	if state.Stage == Playing {
		return nil, errors.New("Cannot start a game that is already started")
	}

	state.StartGame()

	room.StartResetTimer(state.Settings.TimeLimitSecs)
	room.PostponeExpiration()

	beginMsg := BeginMsg{
		NextWord:        state.Turn.CurrWord,
		NextPlayerIndex: state.Turn.CurrPlayerIndex,
	}
	payload := OutputPayload{Code: StartCode, Msg: beginMsg}
	return marshalPayload(payload)
}

type TextMsg struct {
	Text string `json:"text"`
}

type ChatMsg = Chat

func handleTextMessage(state *GameState, msg TextMsg, player Player) ([]byte, error) {
	text := msg.Text
	if len(text) > MaxChatLen || len(text) < MinChatLen {
		return nil, fmt.Errorf("Chat message must be less than %d characters in length and more than %d", MaxChatLen, MinChatLen)
	}

	newChatMessage := ChatMsg{Player: player}

	pointsInc := state.OnGuess(player, text)
	if pointsInc == 0 {
		// only set the text for a failed guess
		newChatMessage.Text = text
	}
	newChatMessage.GuessPointsInc = pointsInc

	state.AddChat(newChatMessage)

	log.Printf("Chat message, %s: %s", player, msg.Text)

	payload := OutputPayload{Code: ChatCode, Msg: newChatMessage}
	return marshalPayload(payload)
}

type DrawMsg = Circle

// color, radius, x, and y are unvalidated fields for performance
func handleDrawMessage(state *GameState, msg DrawMsg, player Player) ([]byte, error) {
	if state.Stage != Playing {
		return nil, errors.New("Can't draw on canvas when game is not being played")
	}
	if player.ID != state.GetCurrPlayer().ID {
		return nil, errors.New("Player cannot draw on the canvas")
	}

	state.Draw(msg)

	payload := OutputPayload{Code: DrawCode, Msg: msg}
	return marshalPayload(payload)
}

type PlayerMsg struct {
	PlayerIndex int    `json:"playerIndex"` // ensures ordering of players on client and server are the same
	Player      Player `json:"player"`
}

func HandleJoin(state *GameState, player Player) ([]byte, error) {
	err := state.Join(player)
	if err != nil {
		return nil, err
	}

	// broadcast the new player to all subscribers
	lastIndex := len(state.Players) - 1
	playerMsg := PlayerMsg{
		PlayerIndex: lastIndex,
		Player:      player,
	}
	payload := OutputPayload{Code: JoinCode, Msg: playerMsg}
	return marshalPayload(payload)
}

func HandleLeave(state *GameState, player Player) ([]byte, error) {
	leaveIndex := state.Leave(player)
	if leaveIndex < 0 {
		return nil, errors.New("Failed to leave the state, player couldn't be found")
	}

	// broadcast the leaving player to all subscribers
	playerMsg := PlayerMsg{
		PlayerIndex: leaveIndex,
		Player:      player,
	}
	payload := OutputPayload{Code: LeaveCode, Msg: playerMsg}
	return marshalPayload(payload)
}

type FinishMsg struct {
	BeginMsg     *BeginMsg `json:"beginMsg"`
	DrawScoreInc int       `json:"drawScoreInc"`
}

func HandleReset(room *Room) ([]byte, error) {
	state := &room.state
	log.Printf("Resetting the game for code %s", state.Code)

	room.PostponeExpiration()

	pointsInc := state.OnReset()

	var beginMsg *BeginMsg = nil
	if state.HasMoreRounds() {
		state.StartGame()
		room.StartResetTimer(state.Settings.TimeLimitSecs)

		beginMsg = &BeginMsg{
			NextWord:        state.Turn.CurrWord,
			NextPlayerIndex: state.Turn.CurrPlayerIndex,
		}
	} else {
		state.FinishGame()
	}

	finishMsg := FinishMsg{
		BeginMsg:     beginMsg,
		DrawScoreInc: pointsInc,
	}
	payload := OutputPayload{Code: FinishCode, Msg: finishMsg}
	return marshalPayload(payload)
}

func marshalPayload(payload OutputPayload) ([]byte, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, ErrMarshal
	}
	return b, nil
}
