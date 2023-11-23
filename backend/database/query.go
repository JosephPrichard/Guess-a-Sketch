/*
 * Copyright (c) Joseph Prichard 2023
 */

package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
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

type StatsUpdate struct {
	playerID        uuid.UUID
	points          int32
	win             bool
	wordsGuessed    int32
	drawingsGuessed int32
}

func UpdateStats(db *sqlx.DB, updates []StatsUpdate) {
	var qb strings.Builder
	var args []interface{}

	for i, u := range updates {
		query := `
			UPDATE players 
			SET points = points + $%d,
				wins = wins + $%d,
				words_guessed = words_guessed + $%d,
				drawings_guessed = drawings_guessed + $%d
			WHERE id = $%d;
		`
		parameters := params(i, 5)
		qb.WriteString(fmt.Sprintf(query, parameters...))
		args = append(args, u.points, u.win, u.wordsGuessed, u.drawingsGuessed, u.playerID)
	}

	_, err := db.Query(qb.String(), args)
	if err != nil {
		log.Printf("Failed to update stats %s", err.Error())
	}
}

func CreateDrawings(db *sqlx.DB, drawings []Drawing) error {
	var qb strings.Builder
	var args []interface{}

	for i, d := range drawings {
		query := "INSERT INTO players (id, created_by, signature) VALUES ($%s, $%s, $%s)"
		parameters := params(i, 3)
		qb.WriteString(fmt.Sprintf(query, parameters...))
		args = append(args, d.ID, d.CreatedBy, d.Signature)
	}

	_, err := db.Query(qb.String(), args)
	if err != nil {
		log.Printf("Failed to insert drawing %s", err.Error())
		return errors.New("Failed to insert drawings")
	}
	return nil
}
