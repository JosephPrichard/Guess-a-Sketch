package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"guessasketch/store"
	"log"
)

const (
	MinWordBank    = 10
	MinChatLen     = 5
	MaxChatLen     = 50
	MinTimeLimit   = 15
	MaxTimeLimit   = 240
	MinPlayerLimit = 2
	MaxPlayerLimit = 12
	MaxTotalRounds = 6
)

var ErrUnMarshal = errors.New("Failed to unmarshal input data")
var ErrMarshal = errors.New("Failed to marshal output data")

func HandleMessage(broker *Broker, message []byte, player string) ([]byte, error) {
	// deserialize payload message from json
	var payload InputPayload
	err := json.Unmarshal(message, &payload)
	if err != nil {
		return nil, err
	}

	log.Printf("Handling message code %d", payload.Code)

	switch payload.Code {
	case OptionsCode:
		log.Printf("Handling message type options")
		var inputMsg OptionsMsg
		err = json.Unmarshal(payload.Msg, &inputMsg)
		if err != nil {
			return nil, ErrUnMarshal
		}
		return handleOptionsMessage(&broker.room, inputMsg, player)
	case StartCode:
		log.Printf("Handling message type start")
		return handleStartMessage(broker, player)
	case TextCode:
		log.Printf("Handling message type text")
		var inputMsg TextMsg
		err = json.Unmarshal(payload.Msg, &inputMsg)
		if err != nil {
			return nil, ErrUnMarshal
		}
		return handleTextMessage(&broker.room, inputMsg, player)
	case DrawCode:
		log.Printf("Handling message type draw")
		var inputMsg DrawMsg
		err = json.Unmarshal(payload.Msg, &inputMsg)
		if err != nil {
			return nil, ErrUnMarshal
		}
		return handleDrawMessage(&broker.room, inputMsg, player)
	default:
		log.Printf("Cannot handle unknown message type")
		return nil, errors.New("No matching message types for message")
	}
}

func handleOptionsMessage(room *store.Room, msg OptionsMsg, player string) ([]byte, error) {
	if room.PlayerIsNotHost(player) {
		return nil, errors.New("Player must be the host to change the game options")
	}
	if room.Stage != store.Lobby {
		return nil, errors.New("Cannot modify options for a game after it starts")
	}
	if len(msg.WordBank) < MinPlayerLimit && len(msg.WordBank) != 0 {
		return nil, fmt.Errorf("Word bank must have at least %d words", MinWordBank)
	}
	if (msg.TimeLimitSecs < MinTimeLimit || msg.TimeLimitSecs > MaxPlayerLimit) && msg.TimeLimitSecs != 0 {
		return nil, fmt.Errorf("Time limit must be between %d and %d seconds", MaxTimeLimit, MaxTimeLimit)
	}
	if (msg.PlayerLimit < MinPlayerLimit || msg.PlayerLimit > MaxPlayerLimit) && msg.PlayerLimit != 0 {
		return nil, fmt.Errorf("Games can only contain between %d and %d players", MaxPlayerLimit, MaxPlayerLimit)
	}
	if msg.PlayerLimit < len(room.Players) && msg.PlayerLimit != 0 {
		return nil, errors.New("Cannot reduce player limit lower than number of players currently in the room")
	}
	if msg.TotalRounds > MaxTotalRounds && msg.TotalRounds != 0 {
		return nil, fmt.Errorf("Games can only contain between %d and %d players", MaxPlayerLimit, MaxPlayerLimit)
	}
	room.Settings.UpdateSettings(msg)

	payload := OutputPayload{Code: OptionsCode, Msg: msg}
	return marshalPayload(payload)
}

func handleStartMessage(broker *Broker, player string) ([]byte, error) {
	room := &broker.room

	if room.PlayerIsNotHost(player) {
		return nil, errors.New("Player must be the host to start the game")
	}
	if room.Stage == store.Playing {
		return nil, errors.New("Cannot start a game that is already started")
	}

	room.StartGame()

	broker.StartResetTimer(room.Settings.TimeLimitSecs)
	broker.PostponeExpiration()

	beginMsg := BeginMsg{
		NextWord:   room.Game.CurrWord,
		NextPlayer: room.GetCurrPlayer(),
	}
	payload := OutputPayload{Code: BeginCode, Msg: beginMsg}
	return marshalPayload(payload)
}

func handleTextMessage(room *store.Room, msg TextMsg, player string) ([]byte, error) {
	text := msg.Text
	if len(text) > MaxChatLen || len(text) < MinChatLen {
		return nil, fmt.Errorf("Chat message must be less than %d characters in length and more than %d", MaxChatLen, MinChatLen)
	}

	newChatMessage := ChatMsg{Player: player}

	scoreInc := room.OnGuess(player, text)
	if scoreInc == 0 {
		// only set the text for a failed guess
		newChatMessage.Text = text
	}
	newChatMessage.GuessScoreInc = scoreInc

	room.AddChat(newChatMessage)

	log.Printf("Chat message, %s: %s", player, msg.Text)

	payload := OutputPayload{Code: ChatCode, Msg: newChatMessage}
	return marshalPayload(payload)
}

// color, radius, x, and y are unvalidated fields for performance
func handleDrawMessage(room *store.Room, msg DrawMsg, player string) ([]byte, error) {
	if room.Stage != store.Playing {
		return nil, errors.New("Can't draw on canvas when game is not being played")
	}
	if player != room.GetCurrPlayer() {
		return nil, errors.New("Player cannot draw on the canvas")
	}

	room.Game.Draw(msg)

	payload := OutputPayload{Code: DrawCode, Msg: msg}
	return marshalPayload(payload)
}

func HandleJoin(room *store.Room, player string) ([]byte, error) {
	if !room.CanJoin(player) {
		return nil, errors.New("Player cannot join because room is at player limit")
	}
	room.Join(player)

	// broadcast the new player to all subscribers
	lastIndex := len(room.Players) - 1
	playerMsg := PlayerMsg{
		PlayerIndex: lastIndex,
		Player:      player,
	}
	payload := OutputPayload{Code: JoinCode, Msg: playerMsg}
	return marshalPayload(payload)
}

func HandleLeave(room *store.Room, player string) ([]byte, error) {
	leaveIndex := room.Leave(player)
	if leaveIndex < 0 {
		return nil, errors.New("Failed to leave the room, player couldn't be found")
	}

	// broadcast the leaving player to all subscribers
	playerMsg := PlayerMsg{
		PlayerIndex: leaveIndex,
		Player:      player,
	}
	payload := OutputPayload{Code: LeaveCode, Msg: playerMsg}
	return marshalPayload(payload)
}

func HandleReset(broker *Broker) ([]byte, error) {
	room := &broker.room
	log.Printf("Resetting the game for code %s", room.Code)

	broker.PostponeExpiration()

	prevPlayer := room.GetCurrPlayer()
	scoreInc := room.OnResetScoreInc()

	var beginMsg *BeginMsg = nil
	if room.CurrRound < room.Settings.TotalRounds {
		room.StartGame()
		broker.StartResetTimer(room.Settings.TimeLimitSecs)

		beginMsg = &BeginMsg{
			NextWord:   room.Game.CurrWord,
			NextPlayer: room.GetCurrPlayer(),
		}
	} else {
		room.FinishGame()
	}

	finishMsg := FinishMsg{
		BeginMsg:      beginMsg,
		PrevPlayer:    prevPlayer,
		GuessScoreInc: scoreInc,
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
