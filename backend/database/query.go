package database

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

var SortColMap = map[string]string{
	"points":   "points",
	"wins":     "wins",
	"words":    "words_guessed",
	"drawings": "drawings_guessed",
}

func GetPlayer(db *sqlx.DB, player *Player, id string) error {
	err := db.Get(player, "SELECT * FROM players WHERE id = ? LIMIT 1", id)
	if err != nil {
		return err
	}
	return nil
}

func GetLeaderboard(db *sqlx.DB, players []Player, limit uint32, sort string) error {
	col, exists := SortColMap[sort]
	if !exists {
		return errors.New("Unknown sort type, must be points, wins, words, or drawings")
	}
	err := db.Select(&players, fmt.Sprintf("SELECT * FROM players ORDER BY %s DESC LIMIT ?", col), limit)
	if err != nil {
		return err
	}
	return nil
}
