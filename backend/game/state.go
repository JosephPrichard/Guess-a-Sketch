/*
 * Copyright (c) Joseph Prichard 2023
 */

package game

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"strings"
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

type GameTurn struct {
	CurrWord        string             `json:"currWord"`        // current word to guess in session
	CurrPlayerIndex int                `json:"currPlayerIndex"` // index of player drawing on canvas
	Canvas          []Circle           `json:"canvas"`          // canvas of circles, acts as a sparse matrix which can be used to construct a bitmap
	guessers        map[uuid.UUID]bool // map storing each player ID who has guessed correctly this game
	startTimeSecs   int64              // start time in milliseconds (unix epoch)
}

type Circle struct {
	Color     uint8  `json:"color"`
	Radius    uint8  `json:"radius"`
	X         uint16 `json:"x"`
	Y         uint16 `json:"y"`
	Connected bool   `json:"connected"`
}

type Snapshot struct {
	SavedBy   *Player
	CreatedBy *Player
	Canvas    string
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

func NewGameState(code string, settings RoomSettings) GameState {
	initialTurn := GameTurn{
		Canvas:          make([]Circle, 0),
		CurrPlayerIndex: -1,
		startTimeSecs:   time.Now().Unix(),
		guessers:        make(map[uuid.UUID]bool),
	}
	return GameState{
		Code:       code,
		Players:    make([]Player, 0),
		ScoreBoard: make(map[uuid.UUID]Score),
		ChatLog:    make([]Chat, 0),
		Settings:   settings,
		Turn:       initialTurn,
	}
}

func (state *GameState) GetCurrPlayer() *Player {
	if state.Turn.CurrPlayerIndex < 0 {
		return &Player{}
	}
	return &state.Players[state.Turn.CurrPlayerIndex]
}

func (state *GameState) PlayerIsNotHost(player Player) bool {
	return len(state.Players) < 1 || state.Players[0] != player
}

func (state *GameState) ToMessage() []byte {
	b, err := json.Marshal(state)
	if err != nil {
		log.Println(err.Error())
		return []byte{}
	}
	return b
}

func (state *GameState) PlayerIndex(playerToFind Player) int {
	// find player in the slice
	index := -1
	for i, player := range state.Players {
		if player == playerToFind {
			index = i
			break
		}
	}
	return index
}

func (state *GameState) Join(player Player) error {
	_, exists := state.ScoreBoard[player.ID]
	if exists {
		return errors.New("Player cannot join a state they are already in")
	}
	if len(state.Players) >= state.Settings.PlayerLimit {
		return errors.New("Player cannot join, state is at player limit")
	}
	state.Players = append(state.Players, player)
	state.ScoreBoard[player.ID] = Score{}
	return nil
}

func (state *GameState) Leave(playerToLeave Player) int {
	index := state.PlayerIndex(playerToLeave)
	if index == -1 {
		// player doesn't exist in players slice - player never joined
		return -1
	}
	// delete player from the slice by creating a new slice without the index
	state.Players = append(state.Players[:index], state.Players[index+1:]...)
	return index
}

// starts the game and returns a snapshot of the settings used to start the game
func (state *GameState) StartGame() {
	state.Stage = Playing

	state.ClearGuessers()
	state.ClearCanvas()

	state.setNextWord()
	state.cycleCurrPlayer()

	state.ResetStartTime()
}

func (state *GameState) FinishGame() {
	state.Stage = Post
}

func (state *GameState) setNextWord() {
	// pick a new word from the shared or custom word bank
	bank := rand.Intn(2)
	if bank == 0 {
		index := rand.Intn(len(state.Settings.SharedWordBank))
		state.Turn.CurrWord = state.Settings.SharedWordBank[index]
	} else {
		index := rand.Intn(len(state.Settings.CustomWordBank))
		state.Turn.CurrWord = state.Settings.CustomWordBank[index]
	}
}

func (state *GameState) cycleCurrPlayer() {
	// go to the next player, circle back around when we reach the end
	state.Turn.CurrPlayerIndex += 1
	if state.Turn.CurrPlayerIndex >= len(state.Players) {
		state.Turn.CurrPlayerIndex = 0
		state.CurrRound += 1
	}
}

// handlers a player's guess and returns the increase in the score of player due to the guess
func (state *GameState) OnGuess(player Player, text string) int {
	// nothing happens if a player guesses when game is not in session
	if state.Stage != Playing {
		return 0
	}
	// current player cannot make a guess
	if player.ID == state.GetCurrPlayer().ID {
		return 0
	}
	// check whether the text is a correct guess or not, if not, do not increase the score
	if !state.ContainsCurrWord(text) {
		return 0
	}
	// cannot increase score of player if they already guessed
	if state.Turn.guessers[player.ID] {
		return 0
	}

	// calculate the score increments for successful guess
	timeSinceStartSecs := time.Now().Unix() - state.Turn.startTimeSecs
	timeLimitSecs := state.Settings.TimeLimitSecs
	pointsInc := (timeLimitSecs-int(timeSinceStartSecs))/timeLimitSecs*400 + 50
	state.incScore(&player, Score{Points: pointsInc, Words: 1})

	state.SetGuesser(&player)
	return pointsInc
}

func (state *GameState) incScore(player *Player, s Score) {
	score, _ := state.ScoreBoard[player.ID]
	score.Points += s.Points
	score.Words += s.Words
	score.Drawings += s.Drawings
	state.ScoreBoard[player.ID] = score
}

func (state *GameState) OnReset() int {
	pointsInc := state.CalcResetScore()
	state.incScore(state.GetCurrPlayer(), Score{Points: pointsInc, Drawings: 1})
	return pointsInc
}

func (state *GameState) AddChat(chat Chat) {
	state.ChatLog = append(state.ChatLog, chat)
}

func (state *GameState) HasMoreRounds() bool {
	return state.CurrRound < state.Settings.TotalRounds
}

func (state *GameState) ClearGuessers() {
	for k := range state.Turn.guessers {
		delete(state.Turn.guessers, k)
	}
}

func (state *GameState) ClearCanvas() {
	state.Turn.Canvas = state.Turn.Canvas[0:0]
}

func (state *GameState) ResetStartTime() {
	state.Turn.startTimeSecs = time.Now().Unix()
}

func (state *GameState) CalcResetScore() int {
	return len(state.Turn.guessers) * 50
}

func (state *GameState) ContainsCurrWord(text string) bool {
	for _, word := range strings.Split(text, " ") {
		if word == state.Turn.CurrWord {
			return true
		}
	}
	return false
}

func (state *GameState) SetGuesser(player *Player) {
	state.Turn.guessers[player.ID] = true
}

func (state *GameState) Draw(stroke Circle) {
	state.Turn.Canvas = append(state.Turn.Canvas, stroke)
}

func (state *GameState) Capture() (Snapshot, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(state.Turn.Canvas)
	if err != nil {
		return Snapshot{}, errors.New("Failed to capture the drawing")
	}
	s := Snapshot{
		Canvas:    buf.String(),
		CreatedBy: state.GetCurrPlayer(),
	}
	return s, nil
}

type GameResult struct {
	PlayerID        uuid.UUID
	Points          int
	Win             bool
	WordsGuessed    int
	DrawingsGuessed int
}

func (state *GameState) CreateGameResult() []GameResult {
	var updates []GameResult
	for id, score := range state.ScoreBoard {
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
