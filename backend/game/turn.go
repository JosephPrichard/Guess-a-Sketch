package game

import (
	"strings"
	"time"
)

type Circle struct {
	Color     uint8  `json:"color"`
	Radius    uint8  `json:"radius"`
	X         uint16 `json:"x"`
	Y         uint16 `json:"y"`
	Connected bool   `json:"connected"`
}

type GameTurn struct {
	CurrWord        string          `json:"currWord"`        // current word to guess in session
	CurrPlayerIndex int             `json:"currPlayerIndex"` // index of player drawing on canvas
	Canvas          []Circle        `json:"canvas"`          // canvas of circles, acts as a sparse matrix which can be used to contruct a bitmap
	guessers        map[string]bool // map storing each player ID who has guessed correctly this game
	startTimeSecs   int64           // start time in milliseconds (unix epoch)
}

func NewGameTurn() GameTurn {
	return GameTurn{
		Canvas:          make([]Circle, 0),
		CurrPlayerIndex: -1,
		startTimeSecs:   time.Now().Unix(),
		guessers:        make(map[string]bool),
	}
}

func (turn *GameTurn) ClearGuessers() {
	for k := range turn.guessers {
		delete(turn.guessers, k)
	}
}

func (turn *GameTurn) ClearCanvas() {
	turn.Canvas = turn.Canvas[0:0]
}

func (turn *GameTurn) ResetStartTime() {
	turn.startTimeSecs = time.Now().Unix()
}

func (turn *GameTurn) CalcResetScore() int {
	return len(turn.guessers) * 50
}

func (turn *GameTurn) ContainsCurrWord(text string) bool {
	for _, word := range strings.Split(text, " ") {
		if word == turn.CurrWord {
			return true
		}
	}
	return false
}

func (turn *GameTurn) SetGuesser(playerID string) {
	turn.guessers[playerID] = true
}

func (turn *GameTurn) Draw(stroke Circle) {
	turn.Canvas = append(turn.Canvas, stroke)
}
