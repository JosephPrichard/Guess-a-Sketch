/*
 * Copyright (c) Joseph Prichard 2023
 */

package database

import (
	"database/sql"
	"errors"
	"fmt"
	"guessthesketch/game"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
)

func GetPlayer(db *sqlx.DB, player *Player, id string) error {
	err := db.Get(player, "SELECT * FROM players WHERE id = $1 LIMIT 1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		log.Printf("Failed to get player stats %s", err.Error())
		return errors.New("Failed to get player stats")
	}
	return nil
}

func CreatePlayer(db *sqlx.DB, player Player) error {
	query := `
		INSERT INTO players (id, username, points, wins, words_guessed, drawings_guessed)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := db.Query(query, player.ID, player.Username, player.Wins, player.WordsGuessed, player.DrawingsGuessed)
	if err != nil {
		log.Printf("Failed to insert player stats %s", err.Error())
		return errors.New("Failed to insert player stats")
	}
	return nil
}

var SortColMap = map[string]string{
	"points":   "points",
	"wins":     "wins",
	"words":    "words_guessed",
	"drawings": "drawings_guessed",
}

func GetLeaderboard(db *sqlx.DB, players []Player, limit uint32, sort string) error {
	if sort == "" {
		sort = "points"
	}
	col, exists := SortColMap[sort]
	if !exists {
		return errors.New("Unknown sort type, must be points, wins, words, or drawings")
	}
	err := db.Select(&players, "SELECT * FROM players ORDER BY $1 DESC LIMIT $2", col, limit)
	if err != nil {
		log.Printf("Failed to get leaderboard %s", err.Error())
		return errors.New("Failed to get leaderboard")
	}
	return nil
}

// utility function to get the ordered query parameters for a given query (by index) in a batch transaction
func params(queryIdx int, parameterCount int) []interface{} {
	parameters := make([]interface{}, parameterCount)
	for j := range parameters {
		parameters[j] = queryIdx*parameterCount + j
	}
	return parameters
}

func UpdateStats(db *sqlx.DB, results []game.GameResult) {
	var qb strings.Builder
	var args []interface{}

	for i, r := range results {
		query := `
			UPDATE players 
			SET points = points + $%d,
				wins = wins + $%d,
				words_guessed = words_guessed + $%d,
				drawings_guessed = drawings_guessed + $%d
			WHERE id = $%d;
		`
		winInc := 0
		if r.Win {
			winInc = 1
		}
		parameters := params(i, 5)
		qb.WriteString(fmt.Sprintf(query, parameters...))
		args = append(args, r.Points, winInc, r.WordsGuessed, r.DrawingsGuessed, r.PlayerID.String())
	}

	_, err := db.Query(qb.String(), args)
	if err != nil {
		log.Printf("Failed to update stats %s", err.Error())
	}
}

func SaveDrawing(db *sqlx.DB, d game.Snapshot) {
	drawing := Drawing{
		SavedBy:   d.SavedBy.ID.String(),
		CreatedBy: d.CreatedBy.ID.String(),
		Signature: d.Canvas,
	}
	query := "INSERT INTO players (created_by, saved_by, signature) VALUES ($1, $2, $3)"
	_, err := db.Query(query, drawing.CreatedBy, drawing.SavedBy, drawing.Signature)
	if err != nil {
		log.Printf("Failed to insert d %s", err.Error())
	}
}
