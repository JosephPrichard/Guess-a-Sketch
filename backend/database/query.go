package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"

	"github.com/jmoiron/sqlx"
)

var SortColMap = map[string]string{
	"points":   "points",
	"wins":     "wins",
	"words":    "words_guessed",
	"drawings": "drawings_guessed",
}

type StatsUpdate struct {
	playerID        uuid.UUID
	points          int32
	win             bool
	wordsGuessed    int32
	drawingsGuessed int32
}

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

func UpdateStats(db *sqlx.DB, updates []StatsUpdate) {
	var fullQuery string
	var args []interface{}

	for i, u := range updates {
		parameterCount := 5
		baseParameter := i * parameterCount

		query := `
			UPDATE players 
			SET points = points + $%d,
				wins = wins + $%d,
				words_guessed = words_guessed + $%d,
				drawings_guessed = drawings_guessed + $%d
			WHERE id = $%d;
		`

		parameters := make([]interface{}, parameterCount)
		for i := range parameters {
			parameters[i] = baseParameter + i
		}

		fullQuery += fmt.Sprintf(query, parameters...)
		args = append(args, u.points, u.win, u.wordsGuessed, u.drawingsGuessed, u.playerID)
	}

	_, err := db.Query(fullQuery, args)
	if err != nil {
		log.Printf("Failed to update stats %s", err.Error())
	}
}
