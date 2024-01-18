/*
 * Copyright (c) Joseph Prichard 2023
 */

package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"guessthesketch/game"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
)

// create a completely new schema on the database
func CreateSchema(db *sqlx.DB) {
	query := `
		DROP TABLE IF EXISTS "players";
		DROP TABLE IF EXISTS "drawings";

		CREATE TABLE players (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL,
			points INTEGER NOT NULL,
			wins INTEGER NOT NULL,
			words_guessed INTEGER NOT NULL,
			drawings_guessed INTEGER NOT NULL
		);

		CREATE TABLE drawings (
			id TEXT PRIMARY KEY,
			created_by TEXT NOT NULL,
			saved_by TEXT NOT NULL,
			signature TEXT NOT NULL
		);

		CREATE INDEX idx_players_username ON players (username);
		CREATE INDEX idx_players_points ON players (points);
		CREATE INDEX idx_players_wins ON players (wins);
		CREATE INDEX idx_players_words_guessed ON players (words_guessed);
		CREATE INDEX idx_players_drawings_guessed ON players (drawings_guessed);`

	_ = db.MustExec(query)
}

func InsertPlayer(db *sqlx.DB, player Player) error {
	query := `
		INSERT INTO players (id, username, points, wins, words_guessed, drawings_guessed)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.Exec(query, player.ID, player.Username, player.Points, player.Wins, player.WordsGuessed, player.DrawingsGuessed)
	if err != nil {
		log.Printf("Failed to insert player stats: %v", err)
		return errors.New("Failed to insert player stats")
	}
	return nil
}

// creates a new player with a random id given a username
func CreateNewPlayer(db *sqlx.DB, name string) (*Player, error) {
	player := Player{ID: uuid.New().String(), Username: name}
	err := InsertPlayer(db, player)
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func GetPlayer(db *sqlx.DB, player *Player, username string) error {
	err := db.Get(player, "SELECT * FROM players WHERE username = $1 LIMIT 1", username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		log.Printf("Failed to get player stats: %v", err)
		return errors.New("Failed to get player stats")
	}
	return nil
}

var SortColMap = map[string]string{
	"points":   "points",
	"wins":     "wins",
	"words":    "words_guessed",
	"drawings": "drawings_guessed",
}

func GetLeaderboard(db *sqlx.DB, limit uint32, sort string) ([]Player, error) {
	if sort == "" {
		sort = "points"
	}
	col, exists := SortColMap[sort]
	if !exists {
		return nil, errors.New("Unknown sort type, must be points, wins, words, or drawings")
	}

	query := fmt.Sprintf("SELECT * FROM players ORDER BY %s DESC LIMIT $1", col)

	var players []Player
	err := db.Select(&players, query, limit)
	if err != nil {
		log.Printf("Failed to get leaderboard: %v", err)
		return nil, errors.New("Failed to get leaderboard")
	}
	return players, nil
}

// utility function to get the ordered query parameters for a given query (by index) in a batch transaction
func params(queryIdx int, parameterCount int) []interface{} {
	parameters := make([]interface{}, parameterCount)
	for j := range parameters {
		parameters[j] = queryIdx*parameterCount + j
	}
	return parameters
}

func UpdateStats(db *sqlx.DB, results []game.GameResult) error {
	var qb strings.Builder
	var args []interface{}

	// update the stats for each player individually - but use a single round-trip
	for i, r := range results {
		query := `
			UPDATE players 
			SET points = points + $%d,
				wins = wins + $%d,
				words_guessed = words_guessed + $%d,
				drawings_guessed = drawings_guessed + $%d
			WHERE id = $%d;`

		winInc := 0
		if r.Win {
			winInc = 1
		}
		parameters := params(i, 5)
		qb.WriteString(fmt.Sprintf(query, parameters...))
		args = append(args, r.Points, winInc, r.WordsGuessed, r.DrawingsGuessed, r.PlayerID)
	}

	_, err := db.Exec(qb.String(), args...)
	if err != nil {
		log.Printf("Failed to update stats: %v", err)
		return err
	}
	return nil
}

func SaveSnapshot(db *sqlx.DB, snap game.Snapshot) error {
	drawing := Drawing{
		ID:        uuid.New().String(),
		CreatedBy: snap.CreatedBy.ID.String(),
		SavedBy:   snap.SavedBy.ID.String(),
		Signature: snap.Canvas,
	}
	return InsertDrawing(db, drawing)
}

func InsertDrawing(db *sqlx.DB, drawing Drawing) error {
	query := `
		INSERT INTO drawings (created_by, saved_by, signature) 
		VALUES ($1, $2, $3)`

	_, err := db.Exec(query, drawing.CreatedBy, drawing.SavedBy, drawing.Signature)
	if err != nil {
		log.Printf("Failed to insert drawing: %v", err)
		return err
	}
	return nil
}

func GetDrawings(db *sqlx.DB, username string) ([]Drawing, error) {
	query := `
		SELECT * FROM drawings d 
		INNER JOIN players p ON 
		    d.id = p.id AND p.username = $1`

	var drawings []Drawing
	err := db.Select(&drawings, query, username)
	if err != nil {
		log.Printf("Failed to get drawings: %v", err)
		return nil, fmt.Errorf("Failed to get saved drawings for %s", username)
	}
	return drawings, nil
}
