/*
 * Copyright (c) Joseph Prichard 2023
 */

package game

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"sort"
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
	code       string              // code of the game that uniquely identifies it
	currRound  int                 // the current round
	players    []Player            // stores all players in the order they joined in
	scoreBoard map[uuid.UUID]Score // maps player IDs to scores
	chatLog    []Chat              // stores the chat log
	stage      int                 // the current stage the room is
	turn       GameTurn            // stores the current game turn
	settings   RoomSettings        // settings for the room set before game starts
}

type GameTurn struct {
	currWord        string             // current word to guess in session
	currPlayerIndex int                // index of player drawing on canvas
	canvas          []Circle           // canvas of circles, acts as a sparse matrix which can be used to construct a bitmap
	guessers        map[uuid.UUID]bool // map storing each player ID who has guessed correctly this game
	startTimeSecs   int64              // start time in milliseconds (unix epoch)
}

type StateJson struct {
	CurrRound  int                 `json:"currRound"`
	Players    []Player            `json:"players"`
	ScoreBoard map[uuid.UUID]Score `json:"scoreBoard"`
	ChatLog    []Chat              `json:"chatLog"`
	Stage      int                 `json:"stage"`
	Turn       TurnJson            `json:"turn"`
}

type TurnJson struct {
	CurrWord   string  `json:"currWord"`
	CurrPlayer *Player `json:"currPlayer"`
	Canvas     string  `json:"canvas"`
}

type Circle struct {
	Color     uint8  `json:"color"`
	Radius    uint8  `json:"radius"`
	X         uint16 `json:"x"`
	Y         uint16 `json:"y"`
	Connected bool   `json:"connected"`
}

type Snapshot struct {
	SavedBy   Player
	CreatedBy Player
	Canvas    string
}

type Score struct {
	Points   int `json:"points"`
	words    int
	drawings int
}

type Player struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	present bool
}

type Chat struct {
	Player         Player `json:"player"`
	Text           string `json:"text"`
	GuessPointsInc int    `json:"guessPointsInc"` // if this is larger than 0, player guessed correctly
}

func NewGameState(code string, settings RoomSettings) GameState {
	initialTurn := GameTurn{
		canvas:          make([]Circle, 0),
		currPlayerIndex: 0,
		startTimeSecs:   time.Now().Unix(),
		guessers:        make(map[uuid.UUID]bool),
	}
	return GameState{
		code:       code,
		players:    make([]Player, 0),
		scoreBoard: make(map[uuid.UUID]Score),
		chatLog:    make([]Chat, 0),
		settings:   settings,
		turn:       initialTurn,
	}
}

func (state *GameState) Code() string {
	return state.code
}

func (state *GameState) EncodeCanvas() string {
	if len(state.turn.canvas) == 0 {
		return ""
	}

	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, state.turn.canvas)
	if err != nil {
		log.Println(err.Error())
		return ""
	}

	base64Encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return base64Encoded
}

func (state *GameState) MarshalJson() []byte {
	currIdx := state.turn.currPlayerIndex
	var curr *Player
	if currIdx >= 0 && currIdx < len(state.players) {
		curr = &state.players[currIdx]
	}

	turnJson := TurnJson{
		CurrWord:   state.turn.currWord,
		CurrPlayer: curr,
		Canvas:     state.EncodeCanvas(),
	}
	b, err := json.Marshal(StateJson{
		CurrRound:  state.currRound,
		Players:    state.Players(),
		ScoreBoard: state.scoreBoard,
		ChatLog:    state.chatLog,
		Turn:       turnJson,
	})

	if err != nil {
		log.Println(err.Error())
		return []byte{}
	}
	return b
}

func (state *GameState) GetCurrPlayer() Player {
	if state.turn.currPlayerIndex < 0 {
		return Player{}
	}
	return state.players[state.turn.currPlayerIndex]
}

func (state *GameState) PlayerIsNotHost(player Player) bool {
	return len(state.players) < 1 || state.players[0].ID != player.ID
}

func (state *GameState) playerIndex(playerToFind Player) int {
	// find player in the slice
	index := -1
	for i, player := range state.players {
		if player.ID == playerToFind.ID {
			index = i
			break
		}
	}
	return index
}

func (state *GameState) Join(player Player) error {
	if len(state.players) >= state.settings.PlayerLimit {
		return errors.New("Player cannot join, state is at player limit")
	}

	index := state.playerIndex(player)
	if index >= 0 {
		// player already exists - they are rejoining and we mark as present
		state.players[index].present = true
	} else {
		// mark the new joined player as present and add it to the end of the players
		player.present = true
		state.players = append(state.players, player)
	}

	_, exists := state.scoreBoard[player.ID]
	if !exists {
		state.scoreBoard[player.ID] = Score{}
	}

	return nil
}

func (state *GameState) Players() []Player {
	players := make([]Player, 0)
	for _, player := range state.players {
		if player.present {
			players = append(players, player)
		}
	}
	return players
}

func (state *GameState) Leave(player Player) int {
	index := state.playerIndex(player)
	if index == -1 {
		// player doesn't exist in players slice - player never joined
		return -1
	}
	state.players[index].present = false
	return index
}

// starts the game and returns a snapshot of the settings used to start the game
func (state *GameState) StartGame() {
	state.stage = Playing
	state.clearGuessers()
	state.clearCanvas()
	state.setNextWord()
	state.cycleCurrPlayer()
	state.resetStartTime()
}

func (state *GameState) clearGuessers() {
	for k := range state.turn.guessers {
		delete(state.turn.guessers, k)
	}
}

func (state *GameState) clearCanvas() {
	state.turn.canvas = state.turn.canvas[0:0]
}

func (state *GameState) setNextWord() {
	// pick a new word from the shared or custom word bank
	numSharedWords := len(state.settings.SharedWordBank)
	numCustomWords := len(state.settings.CustomWordBank)
	if numCustomWords < 1 || rand.Intn(2) == 0 {
		index := rand.Intn(numSharedWords)
		state.turn.currWord = state.settings.SharedWordBank[index]
	} else {
		index := rand.Intn(numCustomWords)
		state.turn.currWord = state.settings.CustomWordBank[index]
	}
}

func (state *GameState) cycleCurrPlayer() {
	// go to the next player, circle back around when we reach the end
	turn := &state.turn
	turn.currPlayerIndex += 1
	if turn.currPlayerIndex >= len(state.players) {
		turn.currPlayerIndex = 0
		state.currRound += 1
	}
}

func (state *GameState) resetStartTime() {
	state.turn.startTimeSecs = time.Now().Unix()
}

func (state *GameState) FinishGame() {
	state.stage = Post
}

func (state *GameState) TryGuess(player Player, text string) Chat {
	pointsInc := state.guess(player, text)

	chat := Chat{Player: player, GuessPointsInc: pointsInc}
	if pointsInc == 0 {
		// only set the text for a failed guess (so we cannot tell what guess is from the log or responses)
		chat.Text = text
	}

	state.chatLog = append(state.chatLog, chat)
	return chat
}

// handlers a player's guess and returns the increase in the score of player due to the guess
func (state *GameState) guess(guesser Player, text string) int {
	if state.stage != Playing {
		return 0
	}
	if guesser.ID == state.GetCurrPlayer().ID {
		return 0
	}
	if !state.containsCurrWord(text) {
		return 0
	}
	if state.turn.guessers[guesser.ID] {
		return 0
	}

	// calculate the score increments for successful guess
	elapsed := time.Now().Unix() - state.turn.startTimeSecs
	limit := state.settings.TimeLimitSecs
	pointsInc := (limit-int(elapsed))/limit*400 + 50
	state.incScore(guesser, Score{Points: pointsInc, words: 1})

	state.turn.guessers[guesser.ID] = true
	return pointsInc
}

func (state *GameState) containsCurrWord(text string) bool {
	for _, word := range strings.Split(text, " ") {
		if strings.ToLower(word) == state.turn.currWord {
			return true
		}
	}
	return false
}

func (state *GameState) incScore(player Player, s Score) {
	score, _ := state.scoreBoard[player.ID]
	score.Points += s.Points
	score.words += s.words
	score.drawings += s.drawings
	state.scoreBoard[player.ID] = score
}

func (state *GameState) OnReset() int {
	pointsInc := state.calcResetScore()
	state.incScore(state.GetCurrPlayer(), Score{Points: pointsInc, drawings: 1})
	return pointsInc
}

func (state *GameState) calcResetScore() int {
	return len(state.turn.guessers) * 50
}

func (state *GameState) HasMoreRounds() bool {
	return state.currRound < state.settings.TotalRounds
}

func (state *GameState) Draw(stroke Circle) {
	state.turn.canvas = append(state.turn.canvas, stroke)
}

func (state *GameState) Capture(player Player) Snapshot {
	return Snapshot{
		Canvas:    state.EncodeCanvas(),
		CreatedBy: state.GetCurrPlayer(),
		SavedBy:   player,
	}
}

type GameResult struct {
	PlayerID        uuid.UUID
	Points          int
	Win             bool
	WordsGuessed    int
	DrawingsGuessed int
}

func compareResults(g1, g2 GameResult) bool {
	return g1.Points > g2.Points
}

func (state *GameState) CreateGameResults() []GameResult {
	var results []GameResult
	for id, score := range state.scoreBoard {
		results = append(results, GameResult{
			PlayerID:        id,
			Points:          score.Points,
			WordsGuessed:    score.words,
			DrawingsGuessed: score.drawings,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return compareResults(results[i], results[j])
	})
	if len(results) > 0 {
		results[0].Win = true
	}

	return results
}
