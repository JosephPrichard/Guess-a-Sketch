package server

import (
	"encoding/json"
	"errors"
	"math/rand"
	"strings"
	"time"
)

const (
	StartCode      = 0
	TextCode       = 1
	DrawCode       = 2
	MaxColor       = 8
	MaxX           = 900
	MaxY           = 500
	MaxRadius      = 8
	MinWordBank    = 10
	MinTime        = 15
	MaxTime        = 360
	MinPlayerLimit = 2
	MaxPlayerLimit = 12
	MaxChatChars   = 50
)

type Payload struct {
	Code int
	Msg  any
}

type Game struct {
	CurrWord        string          // current word to guess in session
	CurrPlayerIndex int             // index of player drawing on canvas
	Canvas          []DrawMsg       // canvas of draw actions, acts as a sparse matrix which can be used to contruct a bitmap
	guessers        map[string]bool // map storing each player who has guessed correctly this game
	startTimeSecs   int64           // start time in milliseconds (unix epoch)
}

type Room struct {
	Code           string         // code of the room that uniquely identifies it
	playerLimit    int            // max players that can join room state, all other players will be spectators
	TotalRounds    int            // total rounds for the game to go through
	Round          int            // the current round
	timeLimitSecs  int            // time given for guessing each turn
	wordBank       []string       // reference to the global wordbank
	customWordBank []string       // custom words added in the bank by host
	Players        []string       // stores all players in the order they joined in
	ScoreBoard     map[string]int // maps players to scores
	Chat           []ChatMsg      // stores the chat log
	Game           *Game          // if the game phase is nil, no game is being played
}

func NewGame() *Game {
	return &Game{
		Canvas:        make([]DrawMsg, 0),
		startTimeSecs: time.Now().UnixMilli(),
		guessers:      make(map[string]bool),
	}
}

func NewRoom(code string, wordBank []string) *Room {
	return &Room{
		Code:           code,
		wordBank:       wordBank,
		customWordBank: make([]string, 0),
		Players:        make([]string, 0),
		ScoreBoard:     make(map[string]int),
		Chat:           make([]ChatMsg, 0),
	}
}

func (room *Room) Marshal() string {
	b, err := json.Marshal(room)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (room *Room) HandleJoin(player string) {
	if len(room.Players) < room.playerLimit {
		room.Players = append(room.Players, player)
		room.ScoreBoard[player] = 0
	}
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

func (room *Room) getCurrPlayer() string {
	return room.Players[room.Game.CurrPlayerIndex]
}

func (room *Room) HandleMessage(message, player string) (string, error) {
	// deserialize payload message from json
	var payload Payload
	err := json.Unmarshal([]byte(message), &payload)
	if err != nil {
		return "", err
	}

	switch payload.Code {
	case StartCode:
		inputMsg := payload.Msg.(StartMsg)
		err = room.handleStartMessage(&inputMsg, player)
	case TextCode:
		inputMsg := payload.Msg.(TextMsg)
		message, err = room.handleTextMessage(&inputMsg, player)
	case DrawCode:
		inputMsg := payload.Msg.(DrawMsg)
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

func (room *Room) onCorrectGuess(player string) int {
	if room.Game.guessers[player] {
		return 0
	}

	// update player scores based on the win ratio algorithm
	timeSinceStartSecs := time.Now().Unix() - room.Game.startTimeSecs

	// ratio of time taken to time allowed normalized over 400 points with a minimum of 50
	scoreInc := (room.timeLimitSecs-int(timeSinceStartSecs))/room.timeLimitSecs*400 + 50
	room.ScoreBoard[player] += scoreInc

	room.Game.guessers[player] = true
	return scoreInc
}

func (room *Room) onGameEnd() {
	// update player scores based on the win ratio algorithm
	scoreInc := len(room.Game.guessers) * 50
	room.ScoreBoard[room.getCurrPlayer()] += scoreInc
	for k := range room.Game.guessers {
		delete(room.Game.guessers, k)
	}

	// pick a new word from the word pool
	index := rand.Intn(len(room.wordBank) + len(room.customWordBank))
	if index < len(room.wordBank) {
		room.Game.CurrWord = room.wordBank[index]
	} else {
		room.Game.CurrWord = room.customWordBank[index]
	}

	// resets the length of the canvas without a new allocation
	room.Game.Canvas = room.Game.Canvas[0:0]

	// go to the next player, circle back around when we reach the end
	room.Game.CurrPlayerIndex += 1
	if room.Game.CurrPlayerIndex >= len(room.Players) {
		room.Game.CurrPlayerIndex = 0
	}
}

func (room *Room) handleStartMessage(msg *StartMsg, player string) error {
	// perform validation to confirm a game can be started
	if len(room.Players) < 1 && room.Players[0] != player {
		return errors.New("Player must be the host to start the game")
	}
	if room.Game != nil {
		return errors.New("Cannot start a game already in session")
	}
	if len(msg.wordBank) < MinWordBank {
		return errors.New("Player was unable to start the game")
	}
	if msg.timeLimitSecs < MinTime || msg.timeLimitSecs > MaxTime {
		return errors.New("Time limit must be between 15 and 360 seconds")
	}
	if msg.playerLimit < MinPlayerLimit || msg.playerLimit > MaxPlayerLimit {
		return errors.New("Games can only contain between 2 and 32 players")
	}
	// initialize the start game state - the params set in the start message and the new game
	room.playerLimit = msg.playerLimit
	room.timeLimitSecs = msg.timeLimitSecs
	room.customWordBank = msg.wordBank
	room.Game = NewGame()
	return nil
}

func (room *Room) handleTextMessage(msg *TextMsg, player string) (string, error) {
	text := msg.Text
	if len(text) > MaxChatChars {
		return "", errors.New("Chat message must be less than 50 characters")
	}

	newChatMessage := ChatMsg{Text: text, Player: player}
	room.Chat = append(room.Chat, newChatMessage)

	if room.Game != nil && player != room.getCurrPlayer() {
		// check if the curr word is included in the room chat
		for _, word := range strings.Split(text, " ") {
			if word == room.Game.CurrWord {
				newChatMessage.ScoreInc = room.onCorrectGuess(player)
				break
			}
		}
	}

	b, err := json.Marshal(newChatMessage)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (room *Room) handleDrawMessage(msg *DrawMsg, player string) error {
	if room.Game == nil {
		return errors.New("Can't draw on ganvas before game has started")
	}
	if player != room.getCurrPlayer() {
		return errors.New("Player cannot draw on the canvas")
	}
	if msg.Color > MaxColor || msg.X > MaxX || msg.Y > MaxY || msg.Radius > MaxRadius {
		return errors.New("Invalid draw format: color, x, y, and radius must match constraints")
	}
	room.Game.Canvas = append(room.Game.Canvas, *msg)
	return nil
}
