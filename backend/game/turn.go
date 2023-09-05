package game

import (
	"strings"
	"time"
)

type Circle struct {
	Color     uint8
	Radius    uint8
	X         uint16
	Y         uint16
	Connected bool
}

type GameTurn struct {
	CurrWord        string          // current word to guess in session
	CurrPlayerIndex int             // index of player drawing on canvas
	Canvas          []Circle        // canvas of circles, acts as a sparse matrix which can be used to contruct a bitmap
	guessers        map[string]bool // map storing each player who has guessed correctly this game
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

func (game *GameTurn) ClearGuessers() {
	for k := range game.guessers {
		delete(game.guessers, k)
	}
}

func (game *GameTurn) ClearCanvas() {
	game.Canvas = game.Canvas[0:0]
}

func (game *GameTurn) ResetStartTime() {
	game.startTimeSecs = time.Now().Unix()
}

func (game *GameTurn) CalcResetScore() int {
	return len(game.guessers) * 50
}

func (game *GameTurn) ContainsCurrWord(text string) bool {
	for _, word := range strings.Split(text, " ") {
		if word == game.CurrWord {
			return true
		}
	}
	return false
}

func (game *GameTurn) SetGuesser(player string) {
	game.guessers[player] = true
}

func (game *GameTurn) Draw(stroke Circle) {
	game.Canvas = append(game.Canvas, stroke)
}
