/*
 * Copyright (c) Joseph Prichard 2023
 */

package game

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"time"
)

const (
	Lobby   = 0
	Playing = 1
	Post    = 2
)

// represents the entire state of the game at any given point in time
type GameState struct {
	Code       string              `json:"code"`       // code of the game that uniquely identifies it
	CurrRound  int                 `json:"currRound"`  // the current round
	Players    []Player            `json:"players"`    // stores all players in the order they joined in
	ScoreBoard map[uuid.UUID]Score `json:"scoreBoard"` // maps player IDs to scores
	ChatLog    []Chat              `json:"chatLog"`    // stores the chat log
	Stage      int                 `json:"stage"`      // the current stage the room is
	Settings   RoomSettings        `json:"settings"`   // settings for the room set before game starts
	Turn       GameTurn            `json:"turn"`       // stores the current game turn
}

type Score struct {
	Points   int `json:"points"`
	Words    int
	Drawings int
	Win      bool
}

type Player struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type Chat struct {
	Player         Player `json:"player"`
	Text           string `json:"text"`
	GuessPointsInc int    `json:"guessPointsInc"` // if this is larger than 0, player guessed correctly
}

func NewGameRoom(code string, settings RoomSettings) GameState {
	return GameState{
		Code:       code,
		Players:    make([]Player, 0),
		ScoreBoard: make(map[uuid.UUID]Score),
		ChatLog:    make([]Chat, 0),
		Settings:   settings,
		Turn:       NewGameTurn(),
	}
}

func (room *GameState) GetCurrPlayer() *Player {
	if room.Turn.CurrPlayerIndex < 0 {
		return &Player{}
	}
	return &room.Players[room.Turn.CurrPlayerIndex]
}

func (room *GameState) PlayerIsNotHost(player Player) bool {
	return len(room.Players) < 1 || room.Players[0] != player
}

func (room *GameState) ToMessage() []byte {
	b, err := json.Marshal(room)
	if err != nil {
		log.Println(err.Error())
		return []byte{}
	}
	return b
}

func (room *GameState) PlayerIndex(playerToFind Player) int {
	// find player in the slice
	index := -1
	for i, player := range room.Players {
		if player == playerToFind {
			index = i
			break
		}
	}
	return index
}

func (room *GameState) Join(player Player) error {
	_, exists := room.ScoreBoard[player.ID]
	if exists {
		return errors.New("Player cannot join a room they are already in")
	}
	if len(room.Players) >= room.Settings.PlayerLimit {
		return errors.New("Player cannot join, room is at player limit")
	}
	room.Players = append(room.Players, player)
	room.ScoreBoard[player.ID] = Score{}
	return nil
}

func (room *GameState) Leave(playerToLeave Player) int {
	index := room.PlayerIndex(playerToLeave)
	if index == -1 {
		// player doesn't exist in players slice - player never joined
		return -1
	}
	// delete player from the slice by creating a new slice without the index
	room.Players = append(room.Players[:index], room.Players[index+1:]...)
	return index
}

// starts the game and returns a snapshot of the settings used to start the game
func (room *GameState) StartGame() {
	room.Stage = Playing

	room.Turn.ClearGuessers()
	room.Turn.ClearCanvas()

	room.setNextWord()
	room.cycleCurrPlayer()

	room.Turn.ResetStartTime()
}

func (room *GameState) FinishGame() {
	room.Stage = Post
}

func (room *GameState) setNextWord() {
	// pick a new word from the shared or custom word bank
	bank := rand.Intn(2)
	if bank == 0 {
		index := rand.Intn(len(room.Settings.SharedWordBank))
		room.Turn.CurrWord = room.Settings.SharedWordBank[index]
	} else {
		index := rand.Intn(len(room.Settings.CustomWordBank))
		room.Turn.CurrWord = room.Settings.CustomWordBank[index]
	}
}

func (room *GameState) cycleCurrPlayer() {
	// go to the next player, circle back around when we reach the end
	room.Turn.CurrPlayerIndex += 1
	if room.Turn.CurrPlayerIndex >= len(room.Players) {
		room.Turn.CurrPlayerIndex = 0
		room.CurrRound += 1
	}
}

// handlers a player's guess and returns the increase in the score of player due to the guess
func (room *GameState) OnGuess(player Player, text string) int {
	// nothing happens if a player guesses when game is not in session
	if room.Stage != Playing {
		return 0
	}
	// current player cannot make a guess
	if player.ID == room.GetCurrPlayer().ID {
		return 0
	}
	// check whether the text is a correct guess or not, if not, do not increase the score
	if !room.Turn.ContainsCurrWord(text) {
		return 0
	}
	// cannot increase score of player if they already guessed
	if room.Turn.guessers[player.ID] {
		return 0
	}

	// calculate the score increments for successful guess
	timeSinceStartSecs := time.Now().Unix() - room.Turn.startTimeSecs
	timeLimitSecs := room.Settings.TimeLimitSecs
	pointsInc := (timeLimitSecs-int(timeSinceStartSecs))/timeLimitSecs*400 + 50
	room.incScore(&player, Score{Points: pointsInc, Words: 1})

	room.Turn.SetGuesser(&player)
	return pointsInc
}

func (room *GameState) incScore(player *Player, s Score) {
	score, _ := room.ScoreBoard[player.ID]
	score.Points += s.Points
	score.Words += s.Words
	score.Drawings += s.Drawings
	room.ScoreBoard[player.ID] = score
}

type GameResult struct {
	PlayerID        uuid.UUID
	Points          int
	Win             bool
	WordsGuessed    int
	DrawingsGuessed int
}

func (room *GameState) CreateGameResult() []GameResult {
	var updates []GameResult
	for id, score := range room.ScoreBoard {
		updates = append(updates, GameResult{
			PlayerID:        id,
			Points:          score.Points,
			WordsGuessed:    score.Words,
			DrawingsGuessed: score.Drawings,
		})
	}

	var highestUpdate *GameResult
	highestPoints := 0
	for _, u := range updates {
		if u.Points > highestPoints {
			highestPoints = u.Points
			highestUpdate = &u
		}
	}
	if highestUpdate != nil {
		highestUpdate.Win = true
	}

	return updates
}

func (room *GameState) OnReset() int {
	pointsInc := room.Turn.CalcResetScore()
	room.incScore(room.GetCurrPlayer(), Score{Points: pointsInc, Drawings: 1})
	return pointsInc
}

func (room *GameState) AddChat(chat Chat) {
	room.ChatLog = append(room.ChatLog, chat)
}

func (room *GameState) HasMoreRounds() bool {
	return room.CurrRound < room.Settings.TotalRounds
}
