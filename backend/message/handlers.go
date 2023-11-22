package message

import (
	"encoding/json"
	"errors"
	"fmt"
	"guessasketch/game"
	"log"
)

const (
	MinChatLen = 5
	MaxChatLen = 50
)

var ErrUnMarshal = errors.New("Failed to unmarshal input data")
var ErrMarshal = errors.New("Failed to marshal output data")

func HandleMessage(broker *Broker, message []byte, player Player) ([]byte, error) {
	// deserialize payload message from json
	var payload InputPayload
	err := json.Unmarshal(message, &payload)
	if err != nil {
		return nil, err
	}

	log.Printf("Handling message code %d", payload.Code)

	switch payload.Code {
	case StartCode:
		if err != nil {
			return nil, ErrUnMarshal
		}
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
		log.Println("Cannot handle unknown message type")
		return nil, errors.New("No matching message types for message")
	}
}

func handleStartMessage(broker *Broker, player Player) ([]byte, error) {
	return handleStartGame(broker, player)
}

func handleStartGame(broker *Broker, player Player) ([]byte, error) {
	room := &broker.room

	if room.PlayerIsNotHost(player) {
		return nil, errors.New("Player must be the host to start the game")
	}
	if room.Stage == game.Playing {
		return nil, errors.New("Cannot start a game that is already started")
	}

	room.StartGame()

	broker.StartResetTimer(room.Settings.TimeLimitSecs)
	broker.PostponeExpiration()

	beginMsg := BeginMsg{
		NextWord:        room.Turn.CurrWord,
		NextPlayerIndex: room.Turn.CurrPlayerIndex,
	}
	payload := OutputPayload{Code: StartCode, Msg: beginMsg}
	return marshalPayload(payload)
}

func handleTextMessage(room *game.Room, msg TextMsg, player Player) ([]byte, error) {
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
func handleDrawMessage(room *game.Room, msg DrawMsg, player Player) ([]byte, error) {
	if room.Stage != game.Playing {
		return nil, errors.New("Can't draw on canvas when game is not being played")
	}
	if player.ID != room.GetCurrPlayer().ID {
		return nil, errors.New("Player cannot draw on the canvas")
	}

	room.Turn.Draw(msg)

	payload := OutputPayload{Code: DrawCode, Msg: msg}
	return marshalPayload(payload)
}

func HandleJoin(room *game.Room, player game.Player) ([]byte, error) {
	err := room.Join(player)
	if err != nil {
		return nil, err
	}

	// broadcast the new player to all subscribers
	lastIndex := len(room.Players) - 1
	playerMsg := PlayerMsg{
		PlayerIndex: lastIndex,
		Player:      player,
	}
	payload := OutputPayload{Code: JoinCode, Msg: playerMsg}
	return marshalPayload(payload)
}

func HandleLeave(room *game.Room, player Player) ([]byte, error) {
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

	scoreInc := room.OnResetScoreInc()

	var beginMsg *BeginMsg = nil
	if room.CurrRound < room.Settings.TotalRounds {
		room.StartGame()
		broker.StartResetTimer(room.Settings.TimeLimitSecs)

		beginMsg = &BeginMsg{
			NextWord:        room.Turn.CurrWord,
			NextPlayerIndex: room.Turn.CurrPlayerIndex,
		}
	} else {
		room.FinishGame()
	}

	finishMsg := FinishMsg{
		BeginMsg:     beginMsg,
		DrawScoreInc: scoreInc,
	}
	payload := OutputPayload{Code: FinishCode, Msg: finishMsg}
	return marshalPayload(payload)
}

func HandleTimeoutMessage() ([]byte, error) {
	payload := OutputPayload{Code: TimeoutCode}
	return marshalPayload(payload)
}

func marshalPayload(payload OutputPayload) ([]byte, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, ErrMarshal
	}
	return b, nil
}
