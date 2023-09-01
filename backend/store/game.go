package store

import (
	"strings"
	"time"
)

type Stroke struct {
	Color     uint8
	Radius    uint8
	X         uint16 // position of where the stroke ends on the canvas
	Y         uint16
	Connected bool // stores whether pen was picked up between strokes
}

type Game struct {
	CurrWord        string          // current word to guess in session
	CurrPlayerIndex int             // index of player drawing on canvas
	Canvas          []Stroke        // canvas of strokes, acts as a sparse matrix which can be used to contruct a bitmap
	guessers        map[string]bool // map storing each player who has guessed correctly this game
	startTimeSecs   int64           // start time in milliseconds (unix epoch)
}

func NewGame() Game {
	return Game{
		Canvas:          make([]Stroke, 0),
		CurrPlayerIndex: -1,
		startTimeSecs:   time.Now().Unix(),
		guessers:        make(map[string]bool),
	}
}

func (game *Game) ClearGuessers() {
	for k := range game.guessers {
		delete(game.guessers, k)
	}
}

func (game *Game) ClearCanvas() {
	game.Canvas = game.Canvas[0:0]
}

func (game *Game) ResetStartTime() {
	game.startTimeSecs = time.Now().Unix()
}

func (game *Game) CalcResetScore() int {
	return len(game.guessers) * 50
}

func (game *Game) ContainsCurrWord(text string) bool {
	for _, word := range strings.Split(text, " ") {
		if word == game.CurrWord {
			return true
		}
	}
	return false
}

func (game *Game) SetGuesser(player string) {
	game.guessers[player] = true
}

func (game *Game) Draw(stroke Stroke) {
	game.Canvas = append(game.Canvas, stroke)
}
