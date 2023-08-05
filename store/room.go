package store

import (
	"encoding/json"
	"math/rand"
	"sync/atomic"
	"time"
)

const (
	Lobby   = 0
	Playing = 1
	Post    = 2
)

type StartEvent = func(settings RoomSettings)

type Chat struct {
	Player        string
	Text          string
	GuessScoreInc int // if this is larger than 0, player guessed correctly
}

type Room struct {
	Code           string         // code of the room that uniquely identifies it
	CurrRound      int            // the current round
	Players        []string       // stores all players in the order they joined in
	ScoreBoard     map[string]int // maps players to scores
	ChatLog        []Chat         // stores the chat log
	Stage          int            // the current stage the room is (upports concurrent operations)
	sharedWordBank []string       // reference to the shared wordbank
	Settings       RoomSettings   // settings for the room set before game starts
	Game           *Game          // if the game state is nil, no game is being played
	ExpireTime     atomic.Int64   // last access of the game
}

func NewRoom(code string, sharedWordBank []string) *Room {
	room := &Room{
		Code:           code,
		Players:        make([]string, 0),
		ScoreBoard:     make(map[string]int),
		ChatLog:        make([]Chat, 0),
		sharedWordBank: sharedWordBank,
		Settings:       NewSettings(),
	}
	room.PostponeExpiration()
	return room
}

func (room *Room) PostponeExpiration() {
	// set the expiration time for 15 minutes
	room.ExpireTime.Store(time.Now().Unix() + 15*60)
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

// starts the game and returns a snapshot of the settings used to start the game
func (room *Room) StartGame() RoomSettings {
	room.Stage = Playing

	room.Game.ClearGuessers()
	room.Game.ClearCanvas()

	room.setNextWord()
	room.cycleCurrPlayer()

	room.Game.ResetStartTime()
	return room.Settings
}

func (room *Room) FinishGame() {
	room.Game = nil
	room.Stage = Post
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
	room.ChatLog = append(room.ChatLog, chat)
}
