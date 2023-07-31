package server

import (
	"encoding/json"
	"errors"
	"guessasketch/store"
	"log"
)

func HandleMessage(room *store.Room, message, player string) (string, error) {
	// deserialize payload message from json
	var payload InputPayload
	err := json.Unmarshal([]byte(message), &payload)
	if err != nil {
		return "", err
	}

	switch payload.Code {
	case OptionsCode:
		var inputMsg OptionsMsg
		err = json.Unmarshal(payload.Msg, &inputMsg)
		if err != nil {
			return "", errors.New(UnMarshalErrMsg)
		}
		err = handleOptionsMessage(room, inputMsg, player)
	case StartCode:
		message, err = handleStartMessage(room, player)
	case TextCode:
		var inputMsg TextMsg
		err = json.Unmarshal(payload.Msg, &inputMsg)
		if err != nil {
			return "", errors.New(UnMarshalErrMsg)
		}
		message, err = handleTextMessage(room, inputMsg, player)
	case DrawCode:
		var inputMsg DrawMsg
		err = json.Unmarshal(payload.Msg, &inputMsg)
		if err != nil {
			return "", errors.New(UnMarshalErrMsg)
		}
		err = handleDrawMessage(room, inputMsg, player)
	default:
		err = errors.New("No matching message types for message")
	}

	if err != nil {
		return "", err
	}

	// sends back the input message back for all cases
	return message, nil
}

func handleOptionsMessage(room *store.Room, msg OptionsMsg, player string) error {
	if room.PlayerIsNotHost(player) {
		return errors.New("Player must be the host to change the game options")
	}
	if room.Game != nil {
		return errors.New("Cannot modify options for a game already in session")
	}
	if len(msg.WordBank) < 10 && len(msg.WordBank) != 0 {
		return errors.New("Word bank must have at least 10 words")
	}
	if (msg.TimeLimitSecs < 15 || msg.TimeLimitSecs > 240) && msg.TimeLimitSecs != 0 {
		return errors.New("Time limit must be between 15 and 240 seconds")
	}
	if (msg.PlayerLimit < 2 || msg.PlayerLimit > 12) && msg.PlayerLimit != 0 {
		return errors.New("Games can only contain between 2 and 12 players")
	}
	if msg.PlayerLimit < len(room.Players) && msg.PlayerLimit != 0 {
		return errors.New("Cannot reduce player limit lower than number of players currently in the room")
	}
	room.Settings.UpdateSettings(msg)
	return nil
}

func handleStartMessage(room *store.Room, player string) (string, error) {
	if room.PlayerIsNotHost(player) {
		return "", errors.New("Player must be the host to start the game")
	}
	if room.Game != nil {
		return "", errors.New("Cannot start a game that is already started")
	}

	room.Game = store.NewGame()
	room.StartGame()

	beginMsg := BeginMsg{
		NextWord:   room.Game.CurrWord,
		NextPlayer: room.GetCurrPlayer(),
	}
	payload := OutputPayload{Code: BeginCode, Msg: beginMsg}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", errors.New(MarshallErrMsg)
	}
	return string(b), nil
}

func handleTextMessage(room *store.Room, msg TextMsg, player string) (string, error) {
	text := msg.Text
	if len(text) > 50 || len(text) < 5 {
		return "", errors.New("Chat message must be less than 50 characters in length and more than 5")
	}

	newChatMessage := ChatMsg{Text: text, Player: player}
	room.AddChat(newChatMessage)

	if player != room.GetCurrPlayer() && room.Game != nil && room.Game.ContainsCurrWord(text) {
		newChatMessage.GuessScoreInc = room.OnCorrectGuess(player)
	}

	log.Printf("Chat message, %s: %s", player, msg.Text)

	payload := OutputPayload{Code: ChatCode, Msg: newChatMessage}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", errors.New(MarshallErrMsg)
	}
	return string(b), nil
}

func handleDrawMessage(room *store.Room, msg DrawMsg, player string) error {
	if room.Game == nil {
		return errors.New("Can't draw on ganvas before game has started")
	}
	if player != room.GetCurrPlayer() {
		return errors.New("Player cannot draw on the canvas")
	}
	if msg.Color > 8 || msg.Radius > 8 {
		return errors.New("Color and radius enums must be between 0 and 8")
	}
	if msg.Y > 800 || msg.X > 500 {
		return errors.New("Circles must be drawn at the board between 0,0 and 800,500")
	}
	room.Game.DrawCircle(msg)
	return nil
}

func HandleJoin(room *store.Room, player string) (string, error) {
	room.Join(player)

	// broadcast the new player to all subscribers
	lastIndex := len(room.Players) - 1
	playerMsg := PlayerMsg{
		playerIndex: lastIndex,
		player: player,
	}
	payload := OutputPayload{Code: JoinCode, Msg: playerMsg}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func HandleLeave(room *store.Room, player string) (string, error) {
	leaveIndex := room.Leave(player)
	if leaveIndex < 0 {
		return "", errors.New("Failed to leave the room, player couldn't be found")
	}

	// broadcast the leaving player to all subscribers
	playerMsg := PlayerMsg{
		playerIndex: leaveIndex,
		player: player,
	}
	payload := OutputPayload{Code: LeaveCode, Msg: playerMsg}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func HandleReset(room *store.Room) (string, error) {
	log.Printf("Resetting the game for code %s", room.Code)

	prevPlayer := room.GetCurrPlayer()
	scoreInc := room.OnResetScoreInc()

	var beginMsg *BeginMsg = nil
	if room.CurrRound < room.Settings.TotalRounds {
		room.StartGame()

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
	b, err := json.Marshal(payload)
	if err != nil {
		return "", errors.New(MarshallErrMsg)
	}
	return string(b), nil
}
