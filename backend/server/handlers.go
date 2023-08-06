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
	MaxDrawField   = 8
	MaxX           = 500
	MaxY           = 800
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

	switch payload.Code {
	case OptionsCode:
		var inputMsg OptionsMsg
		err = json.Unmarshal(payload.Msg, &inputMsg)
		if err != nil {
			return nil, ErrUnMarshal
		}
		return handleOptionsMessage(&broker.room, inputMsg, player)
	case StartCode:
		return handleStartMessage(broker, player)
	case TextCode:
		var inputMsg TextMsg
		err = json.Unmarshal(payload.Msg, &inputMsg)
		if err != nil {
			return nil, ErrUnMarshal
		}
		return handleTextMessage(&broker.room, inputMsg, player)
	case DrawCode:
		var inputMsg DrawMsg
		err = json.Unmarshal(payload.Msg, &inputMsg)
		if err != nil {
			return nil, ErrUnMarshal
		}
		return handleDrawMessage(&broker.room, inputMsg, player)
	default:
		return nil, errors.New("No matching message types for message")
	}
}

func handleOptionsMessage(room *store.Room, msg OptionsMsg, player string) ([]byte, error) {
	if room.PlayerIsNotHost(player) {
		return nil, errors.New("Player must be the host to change the game options")
	}
	if room.Game != nil {
		return nil, errors.New("Cannot modify options for a game already in session")
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
	room.Settings.UpdateSettings(msg)

	payload := OutputPayload{Code: OptionsCode, Msg: msg}
	return marshalPayload(payload)
}

func handleStartMessage(broker *Broker, player string) ([]byte, error) {
	room := &broker.room

	if room.PlayerIsNotHost(player) {
		return nil, errors.New("Player must be the host to start the game")
	}
	if room.Game != nil {
		return nil, errors.New("Cannot start a game that is already started")
	}

	room.Game = store.NewGame()
	settings := room.StartGame()

	broker.StartResetTimer(settings.TimeLimitSecs)
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

	newChatMessage := ChatMsg{Text: text, Player: player}
	room.AddChat(newChatMessage)

	if player != room.GetCurrPlayer() && room.Game != nil && room.Game.ContainsCurrWord(text) {
		newChatMessage.GuessScoreInc = room.OnCorrectGuess(player)
	}

	log.Printf("Chat message, %s: %s", player, msg.Text)

	payload := OutputPayload{Code: ChatCode, Msg: newChatMessage}
	return marshalPayload(payload)
}

func handleDrawMessage(room *store.Room, msg DrawMsg, player string) ([]byte, error) {
	if room.Game == nil {
		return nil, errors.New("Can't draw on ganvas before game has started")
	}
	if player != room.GetCurrPlayer() {
		return nil, errors.New("Player cannot draw on the canvas")
	}
	if msg.Color > MaxDrawField || msg.Radius > MaxDrawField {
		return nil, fmt.Errorf("Color and radius enums must be between 0 and %d", MaxDrawField)
	}
	if msg.X > MaxX || msg.Y > MaxY {
		return nil, fmt.Errorf("Circles must be drawn at the board between 0,0 and %d,%d", MaxX, MaxY)
	}
	room.Game.DrawCircle(msg)

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
		playerIndex: lastIndex,
		player:      player,
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
		playerIndex: leaveIndex,
		player:      player,
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
		settings := room.StartGame()
		broker.StartResetTimer(settings.TimeLimitSecs)

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
