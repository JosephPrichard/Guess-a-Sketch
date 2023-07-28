package server

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"strings"
	"time"
)

const (
	// message code constants
	OptionsCode     = 0
	StartCode       = 1
	TextCode        = 2
	DrawCode        = 3
	ChatCode        = 4
	FinishRoundCode = 5
	FinishGameCode  = 6
	BeginCode       = 7
	// room state constants that store room relative to game
	Before = 0
	During = 1
	After  = 2
)

type InputPayload struct {
	Code int
	Msg  json.RawMessage
}

type OutputPayload struct {
	Code int
	Msg  interface{}
}

type GameState struct {
	CurrWord        string          // current word to guess in session
	CurrPlayerIndex int             // index of player drawing on canvas
	Canvas          []DrawMsg       // canvas of draw actions, acts as a sparse matrix which can be used to contruct a bitmap
	guessers        map[string]bool // map storing each player who has guessed correctly this game
	startTimeSecs   int64           // start time in milliseconds (unix epoch)
}

type Room struct {
	Code            string         // code of the room that uniquely identifies it
	startResetTimer func(int)      // called whenever the game is reset, takes the time limit for a game as an arg
	playerLimit     int            // max players that can join room state
	TotalRounds     int            // total rounds for the game to go through
	Round           int            // the current round
	TimeLimitSecs   int            // time given for guessing each turn
	sharedWordBank  []string       // reference to the shared wordbank
	customWordBank  []string       // custom words added in the bank by host
	Players         []string       // stores all players in the order they joined in
	ScoreBoard      map[string]int // maps players to scores
	Chat            []ChatMsg      // stores the chat log
	Stage           int            // the current stage the room is in
	Game            *GameState     // if the game state is nil, no game is being played
}

func NewGame() *GameState {
	return &GameState{
		Canvas:          make([]DrawMsg, 0),
		CurrPlayerIndex: -1,
		startTimeSecs:   time.Now().Unix(),
		guessers:        make(map[string]bool),
	}
}

func NewRoom(code string, sharedWordBank []string, startResetTimer func(time int)) *Room {
	return &Room{
		Code:            code,
		startResetTimer: startResetTimer,
		sharedWordBank:  sharedWordBank,
		customWordBank:  make([]string, 0),
		Players:         make([]string, 0),
		ScoreBoard:      make(map[string]int),
		Chat:            make([]ChatMsg, 0),
		Stage:           Before,
	}
}

func (room *Room) getCurrPlayer() string {
	if room.Game.CurrPlayerIndex < 0 {
		return ""
	}
	return room.Players[room.Game.CurrPlayerIndex]
}

func (room *Room) playerIsHost(player string) bool {
	return len(room.Players) < 1 && room.Players[0] != player
}

func (room *Room) Marshal() string {
	b, err := json.Marshal(room)
	if err != nil {
		return MARSHAL_ERR_MSG
	}
	return string(b)
}

func (room *Room) HandleJoin(player string) {
	room.Players = append(room.Players, player)
	room.ScoreBoard[player] = 0
}

func (room *Room) HandleLeave(playerToLeave string) {
	// find player in the slice
	index := -1
	for i, player := range room.Players {
		if player == playerToLeave {
			index = i
			break
		}
	}
	if index == -1 {
		// player doesn't exist in players slice - player never joined
		return
	}
	// delete player from the slice by creating a new slice without the index
	room.Players = append(room.Players[:index], room.Players[index+1:]...)
	// delete player from scoreboard
	delete(room.ScoreBoard, playerToLeave)
}

func (room *Room) StartGame() {
	// clear all guessers for this game
	for k := range room.Game.guessers {
		delete(room.Game.guessers, k)
	}

	// pick a new word from the shared or custom word bank
	index := rand.Intn(len(room.sharedWordBank) + len(room.customWordBank))
	if index < len(room.sharedWordBank) {
		room.Game.CurrWord = room.sharedWordBank[index]
	} else {
		room.Game.CurrWord = room.customWordBank[index]
	}

	room.Game.Canvas = room.Game.Canvas[0:0]

	// go to the next player, circle back around when we reach the end
	room.Game.CurrPlayerIndex += 1
	if room.Game.CurrPlayerIndex >= len(room.Players) {
		room.Game.CurrPlayerIndex = 0
		room.Round += 1
	}

	room.Game.startTimeSecs = time.Now().Unix()
	go room.startResetTimer(room.TimeLimitSecs)
}

func (room *Room) ResetGame() (string, error) {
	log.Printf("Resetting the game for code %s", room.Code)
	prevPlayer := room.getCurrPlayer()

	// update player scores based on the win ratio algorithm
	scoreInc := len(room.Game.guessers) * 50
	room.ScoreBoard[room.getCurrPlayer()] += scoreInc

	var payload OutputPayload
	// only restart the game if the game has more rounds
	if room.Round < room.TotalRounds {
		room.StartGame()

		// reset message contains the state changes for the next game
		finishMsg := FinishMsg{
			BeginMsg: &BeginMsg{
				NextWord:      room.Game.CurrWord,
				NextPlayer:    room.getCurrPlayer(),
			},
			PrevPlayer:    prevPlayer,
			GuessScoreInc: scoreInc,
		}
		payload = OutputPayload{Code: FinishRoundCode, Msg: finishMsg}
	} else {
		room.Game = nil

		// finish message contains the results of the game
		finishMsg := FinishMsg{
			PrevPlayer:    prevPlayer,
			GuessScoreInc: scoreInc,
		}
		payload = OutputPayload{Code: FinishGameCode, Msg: finishMsg}
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", errors.New(MARSHAL_ERR_MSG)
	}
	return string(b), nil
}

func (room *Room) onCorrectGuess(player string) int {
	if room.Game.guessers[player] {
		return 0
	}

	// update player scores based on the win ratio algorithm
	timeSinceStartSecs := time.Now().Unix() - room.Game.startTimeSecs

	// ratio of time taken to time allowed normalized over 400 points with a minimum of 50
	scoreInc := (room.TimeLimitSecs-int(timeSinceStartSecs))/room.TimeLimitSecs*400 + 50
	room.ScoreBoard[player] += scoreInc

	room.Game.guessers[player] = true
	return scoreInc
}

func (room *Room) HandleMessage(message, player string) (string, error) {
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
			return "", errors.New(MARSHAL_ERR_MSG)
		}
		err = room.handleOptionsMessage(&inputMsg, player)
	case StartCode:
		message, err = room.handleStartMessage(player)
	case TextCode:
		var inputMsg TextMsg
		err = json.Unmarshal(payload.Msg, &inputMsg)
		if err != nil {
			return "", errors.New(MARSHAL_ERR_MSG)
		}
		message, err = room.handleTextMessage(&inputMsg, player)
	case DrawCode:
		var inputMsg DrawMsg
		err = json.Unmarshal(payload.Msg, &inputMsg)
		if err != nil {
			return "", errors.New(MARSHAL_ERR_MSG)
		}
		err = room.handleDrawMessage(&inputMsg, player)
	default:
		err = errors.New("No matching message types for message")
	}

	if err != nil {
		return "", err
	}

	// sends back the input message back for all cases
	return message, nil
}

func (room *Room) handleOptionsMessage(msg *OptionsMsg, player string) error {
	if room.playerIsHost(player) {
		return errors.New("Player must be the host to change the game options")
	}
	if room.Game != nil {
		return errors.New("Cannot modify options for a game already in session")
	}
	if len(msg.wordBank) < 10 {
		return errors.New("Word bank must have at least 10 words")
	}
	if msg.timeLimitSecs < 15 || msg.timeLimitSecs > 360 {
		return errors.New("Time limit must be between 15 and 360 seconds")
	}
	if msg.playerLimit < 2 || msg.playerLimit > 12 {
		return errors.New("Games can only contain between 2 and 12 players")
	}
	// initialize the start game state - the params set in the start message and the new game
	room.playerLimit = msg.playerLimit
	room.TimeLimitSecs = msg.timeLimitSecs
	room.customWordBank = msg.wordBank
	return nil
}

func (room *Room) handleStartMessage(player string) (string, error) {
	if room.playerIsHost(player) {
		return "", errors.New("Player must be the host to start the game")
	}
	if room.Game != nil {
		return "", errors.New("Cannot start a game that is already started")
	}
	room.Game = NewGame()
	room.StartGame()
	beginMsg := BeginMsg{
		NextWord:   room.Game.CurrWord,
		NextPlayer: room.getCurrPlayer(),
	}
	payload := OutputPayload{Code: BeginCode, Msg: beginMsg}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", errors.New(MARSHAL_ERR_MSG)
	}
	return string(b), nil
}

func (room *Room) handleTextMessage(msg *TextMsg, player string) (string, error) {
	text := msg.Text
	if len(text) > 50 || len(text) < 5 {
		return "", errors.New("Chat message must be less than 50 characters in length and more than 5")
	}

	newChatMessage := ChatMsg{Text: text, Player: player}
	room.Chat = append(room.Chat, newChatMessage)

	if room.Game != nil && player != room.getCurrPlayer() {
		// check if the curr word is included in the room chat
		for _, word := range strings.Split(text, " ") {
			if word == room.Game.CurrWord {
				newChatMessage.GuessScoreInc = room.onCorrectGuess(player)
				break
			}
		}
	}
	payload := OutputPayload{Code: ChatCode, Msg: newChatMessage}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", errors.New(MARSHAL_ERR_MSG)
	}
	log.Printf("Chat message, %s: %s", player, msg.Text)
	return string(b), nil
}

func (room *Room) handleDrawMessage(msg *DrawMsg, player string) error {
	if room.Game == nil {
		return errors.New("Can't draw on ganvas before game has started")
	}
	if player != room.getCurrPlayer() {
		return errors.New("Player cannot draw on the canvas")
	}
	if msg.Color > 8 || msg.Radius > 8 {
		return errors.New("Color and radius enums must be between 0 and 8")
	}
	if msg.Y > 800 || msg.X > 500 {
		return errors.New("Circles must be drawn at the board between 0,0 and 800,500")
	}
	room.Game.Canvas = append(room.Game.Canvas, *msg)
	return nil
}
