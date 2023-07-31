package store

import (
	"encoding/json"
	"math/rand"
	"time"
)

const (
	Before = 0
	During = 1
	After  = 2
)

type StartEvent = func(settings RoomSettings)

type Chat struct {
	Player        string
	Text          string
	GuessScoreInc int // if this is larger than 0, player guessed correctly
}

type Room struct {
	Code           string         // code of the room that uniquely identifies it
	onStartGame    StartEvent     // called whenever the game is reset, takes the time limit for a game as an arg
	CurrRound      int            // the current round
	Players        []string       // stores all players in the order they joined in
	ScoreBoard     map[string]int // maps players to scores
	Chat           []Chat         // stores the chat log
	Stage          int            // the current stage the room is in
	sharedWordBank []string       // reference to the shared wordbank
	Settings       RoomSettings   // settings for the room set before game starts
	Game           *Game          // if the game state is nil, no game is being played
}

func NewRoom(code string, sharedWordBank []string, onStartGame StartEvent) *Room {
	return &Room{
		Code:           code,
		onStartGame:    onStartGame,
		Players:        make([]string, 0),
		ScoreBoard:     make(map[string]int),
		Chat:           make([]Chat, 0),
		Stage:          Before,
		sharedWordBank: sharedWordBank,
		Settings:       NewSettings(),
	}
}

func (room *Room) GetCurrPlayer() string {
	if room.Game.CurrPlayerIndex < 0 {
		return ""
	}
	return room.Players[room.Game.CurrPlayerIndex]
}

func (room *Room) PlayerIsNotHost(player string) bool {
	return len(room.Players) < 1 || room.Players[0] != player
}

func (room *Room) Marshal() string {
	b, err := json.Marshal(room)
	if err != nil {
		return "Failed to marshall room data"
	}
	return string(b)
}

func (room *Room) CanJoin() bool {
	return len(room.Players) < room.Settings.playerLimit-1
}

func (room *Room) Join(player string) {
	room.Players = append(room.Players, player)
	room.ScoreBoard[player] = 0
}

func (room *Room) Leave(playerToLeave string) int {
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
		return -1
	}
	// delete player from the slice by creating a new slice without the index
	room.Players = append(room.Players[:index], room.Players[index+1:]...)
	// delete player from scoreboard
	delete(room.ScoreBoard, playerToLeave)
	return index
}

func (room *Room) StartGame() {
	room.Stage = During

	room.Game.ClearGuessers()
	room.Game.ClearCanvas()

	room.setNextWord()
	room.cycleCurrPlayer()

	room.Game.ResetStartTime()
	room.onStartGame(room.Settings)
}

func (room *Room) FinishGame() {
	room.Game = nil
	room.Stage = After
}

func (room *Room) setNextWord() {
	// pick a new word from the shared or custom word bank
	index := rand.Intn(len(room.sharedWordBank) + len(room.Settings.customWordBank))
	if index < len(room.sharedWordBank) {
		room.Game.CurrWord = room.sharedWordBank[index]
	} else {
		room.Game.CurrWord = room.Settings.customWordBank[index]
	}
}

func (room *Room) cycleCurrPlayer() {
	// go to the next player, circle back around when we reach the end
	room.Game.CurrPlayerIndex += 1
	if room.Game.CurrPlayerIndex >= len(room.Players) {
		room.Game.CurrPlayerIndex = 0
		room.CurrRound += 1
	}
}

func (room *Room) OnCorrectGuess(player string) int {
	if room.Game.guessers[player] {
		return 0
	}

	timeSinceStartSecs := time.Now().Unix() - room.Game.startTimeSecs
	timeLimitSecs := room.Settings.TimeLimitSecs

	scoreInc := (timeLimitSecs-int(timeSinceStartSecs))/timeLimitSecs*400 + 50
	room.ScoreBoard[player] += scoreInc

	room.Game.SetGuesser(player)
	return scoreInc
}

func (room *Room) OnResetScoreInc() int {
	scoreInc := room.Game.CalcResetScore()
	room.ScoreBoard[room.GetCurrPlayer()] += scoreInc
	return scoreInc
}

func (room *Room) AddChat(chat Chat) {
	room.Chat = append(room.Chat, chat)
}
