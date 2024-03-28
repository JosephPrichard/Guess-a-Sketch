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
	StateCode   = 11
	ErrorCode   = 12

	MinChatLen = 5
	MaxChatLen = 50
	MaxX       = 1000
	MaxY       = 1000
	MaxRadius  = 8
	MaxColor   = 8
)

var ErrUnMarshal = errors.New("Failed to unmarshal input data")
var ErrMarshal = errors.New("Failed to marshal output data")

type InputPayload[T any] struct {
	Code    int `json:"code"`
	Msg     T   `json:"msg"`
	TraceID string
}

type OutputPayload[T any] struct {
	Code    int `json:"code"`
	Msg     T   `json:"msg"`
	TraceID string
}

func (room *Room) HandleMessage(message []byte, player Player) ([]byte, error) {
	// deserialize payload message from json
	var payload InputPayload[json.RawMessage]
	err := json.Unmarshal(message, &payload)
	if err != nil {
		return nil, err
	}

	switch payload.Code {
	case StartCode:
		return room.handleStartMessage(player, payload.TraceID)
	case TextCode:
		var inputMsg TextMsg
		err = json.Unmarshal(payload.Msg, &inputMsg)
		if err != nil {
			return nil, ErrUnMarshal
		}
		return room.handleTextMessage(inputMsg, player, payload.TraceID)
	case DrawCode:
		var inputMsg DrawMsg
		err = json.Unmarshal(payload.Msg, &inputMsg)
		if err != nil {
			return nil, ErrUnMarshal
		}
		return room.handleDrawMessage(inputMsg, player, payload.TraceID)
	case SaveCode:
		capture := room.state.Capture(player)
		room.handler.DoCapture(capture)
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

func (room *Room) handleStartMessage(player Player, traceID string) ([]byte, error) {
	state := &room.state

	if state.PlayerIsNotHost(player) {
		return nil, errors.New("Player must be the host to start the game")
	}
	if state.stage == Playing {
		return nil, errors.New("Cannot start a game that is already started")
	}

	state.StartGame()

	room.startResetTimer(state.settings.TimeLimitSecs)
	room.postponeExpiration()

	msg := BeginMsg{NextWord: state.turn.currWord, NextPlayerIndex: state.turn.currPlayerIndex}
	return createTracedResponse(BeginCode, msg, traceID)
}

type TextMsg struct {
	Text string `json:"text"`
}

func (room *Room) handleTextMessage(msg TextMsg, player Player, traceID string) ([]byte, error) {
	text := msg.Text
	if len(text) > MaxChatLen || len(text) < MinChatLen {
		return nil, fmt.Errorf("Chat message must be less than %d characters in length and more than %d", MaxChatLen, MinChatLen)
	}

	chat := room.state.TryGuess(player, text)
	log.Printf("Chat message, %+v: %s", player, msg.Text)

	return createTracedResponse(ChatCode, chat, traceID)
}

type DrawMsg = Circle

func (room *Room) handleDrawMessage(msg DrawMsg, player Player, traceID string) ([]byte, error) {
	state := &room.state

	if state.stage != Playing {
		return nil, errors.New("Can't draw on canvas when game is not being played")
	}
	if player.ID != state.GetCurrPlayer().ID {
		return nil, errors.New("Player cannot draw on the canvas")
	}
	if msg.X < 0 || msg.X > MaxX || msg.Y < 0 || msg.Y > MaxY {
		return nil, errors.New("Cannot draw outside canvas")
	}
	if msg.Radius < 0 || msg.Radius > MaxRadius {
		return nil, fmt.Errorf("Unknown code for radius %d", msg.Radius)
	}
	if msg.Color < 0 || msg.Color > MaxColor {
		return nil, fmt.Errorf("Unknown code for color %d", msg.Color)
	}

	state.Draw(msg)
	return createTracedResponse(DrawCode, msg, traceID)
}

type PlayerMsg struct {
	PlayerIndex int    `json:"playerIndex"` // ensures ordering of players on client and server are the same
	Player      Player `json:"player"`
}

func (room *Room) HandleJoin(player Player) ([]byte, error) {
	state := &room.state

	err := state.Join(player)
	if err != nil {
		return nil, err
	}

	lastIndex := len(state.players) - 1
	msg := PlayerMsg{PlayerIndex: lastIndex, Player: player}
	return createResponse(JoinCode, msg)
}

func HandleLeave(state *GameState, player Player) ([]byte, error) {
	leaveIndex := state.Leave(player)
	if leaveIndex < 0 {
		return nil, errors.New("Failed to leave the state, player couldn't be found")
	}

	msg := PlayerMsg{PlayerIndex: leaveIndex, Player: player}
	return createResponse(LeaveCode, msg)
}

type FinishMsg struct {
	BeginMsg     *BeginMsg `json:"beginMsg"`
	DrawScoreInc int       `json:"drawScoreInc"`
}

func (room *Room) HandleReset() ([]byte, error) {
	state := &room.state
	log.Printf("Resetting the game for code %s", state.code)

	room.postponeExpiration()

	pointsInc := state.OnReset()

	var beginMsg *BeginMsg = nil
	if state.HasMoreRounds() {
		state.StartGame()
		room.startResetTimer(state.settings.TimeLimitSecs)

		beginMsg = &BeginMsg{NextWord: state.turn.currWord, NextPlayerIndex: state.turn.currPlayerIndex}
	} else {
		state.FinishGame()
	}

	msg := FinishMsg{BeginMsg: beginMsg, DrawScoreInc: pointsInc}
	return createResponse(FinishCode, msg)
}
 	
func (room *Room) HandleState() ([]byte, error) {
	state := &room.state
	bytes := state.MarshalJson()
	return createResponse[json.RawMessage](StateCode, bytes)
}

func createResponse[T any](code int, msg T) ([]byte, error) {
	return createTracedResponse(code, msg, "")
}

func createTracedResponse[T any](code int, msg T, traceID string) ([]byte, error) {
	payload := OutputPayload[T]{Code: code, Msg: msg, TraceID: traceID}
	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, ErrMarshal
	}
	return buf, nil
}
