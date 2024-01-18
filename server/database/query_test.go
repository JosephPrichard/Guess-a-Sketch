/*
 * Copyright (c) Joseph Prichard 2024
 */

package database

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"guessthesketch/game"
	"reflect"
	"sync"
	"testing"
)

func Test_CreateSchema(t *testing.T) {
	db, err := sqlx.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open db %v", err)
	}
	defer db.Close()
	CreateSchema(db)
}

// test many concurrent writes to check if the database connection mode is correct
func TestQuery_InsertManyPlayers(t *testing.T) {
	db, err := sqlx.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open db %v", err)
	}
	CreateSchema(db)

	count := 1000
	var wg sync.WaitGroup
	errors := make([]error, count)

	for i := 0; i < count; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			player := Player{ID: uuid.New().String(), Username: "Player",
				Points: 1, Wins: 1, WordsGuessed: 1, DrawingsGuessed: 1}
			err := InsertPlayer(db, player)
			if err != nil {
				errors[i] = err
			}
		}(i)
	}

	wg.Wait()

	for i, err := range errors {
		if err != nil {
			t.Errorf("Failed to insert player %d with error %v", i, err)
		}
	}
}

func CreateTestPlayerDb(t *testing.T) (*sqlx.DB, []Player) {
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open db %v", err)
	}
	CreateSchema(db)

	// test data for the player table
	playersTable := []Player{
		// sample data contains no duplicate values per column, so we can use single column sorting to test the sql order by
		{ID: "id1", Username: "Player1", Points: 9, Wins: 2, WordsGuessed: 4, DrawingsGuessed: 5},
		{ID: "id2", Username: "Player2", Points: 2, Wins: 1, WordsGuessed: 3, DrawingsGuessed: 10},
		{ID: "id3", Username: "Player3", Points: 1, Wins: 5, WordsGuessed: 8, DrawingsGuessed: 1},
		{ID: "id4", Username: "Player4", Points: 3, Wins: 2, WordsGuessed: 3, DrawingsGuessed: 5},
	}
	for _, player := range playersTable {
		err := InsertPlayer(db, player)
		if err != nil {
			t.Fatalf("Failed to insert player %v with error %v", player, err)
		}
	}

	return db, playersTable
}

func TestQuery_GetPlayer(t *testing.T) {
	db, playersTable := CreateTestPlayerDb(t)
	defer db.Close()

	var player Player
	err := GetPlayer(db, &player, "Player2")
	if err != nil {
		t.Fatalf("Failed to get player with error %v", err)
	}
	if !reflect.DeepEqual(player, playersTable[1]) {
		t.Fatalf("Expected to get player %v, got %v", playersTable[1], player)
	}
}

func TestQuery_PointsLeaderboard(t *testing.T) {
	db, _ := CreateTestPlayerDb(t)
	defer db.Close()

	leaderboard, err := GetLeaderboard(db, 3, "points")
	if err != nil {
		t.Fatalf("Failed to get player with error %v", err)
	}

	// validate the shape of the leaderboard but not the exact data itself
	if len(leaderboard) > 3 {
		t.Fatalf("Leaderboard with limit of 3 must be less no more than 3 elements")
	}
	for i := 1; i < len(leaderboard); i++ {
		if leaderboard[i].Points > leaderboard[i-1].Points {
			// this element is larger than the previous, so it is not in order
			t.Fatalf("Elements %d and %d in the leaderboard are not in order", i, i-1)
		}
	}
}

func TestQuery_UpdateStats(t *testing.T) {
	db, playersTable := CreateTestPlayerDb(t)
	defer db.Close()

	// update using game results for players 0 and 1
	results := []game.GameResult{
		{PlayerID: playersTable[0].ID, Points: 1, Win: true, WordsGuessed: 3},
		{PlayerID: playersTable[1].ID, Points: 2, Win: false, DrawingsGuessed: 4},
	}
	_ = UpdateStats(db, results)

	// mirror the expected update changes in our local version of the table
	for i, result := range results {
		player := &playersTable[i]
		player.Points += result.Points
		if result.Win {
			player.Wins += 1
		}
		player.WordsGuessed += result.WordsGuessed
		player.DrawingsGuessed += result.DrawingsGuessed
	}

	for _, expectedPlayer := range playersTable {
		// compare each expected expectedPlayer after update expectedPlayer with actual version in the database
		var actualPlayer Player
		err := GetPlayer(db, &actualPlayer, expectedPlayer.Username)
		if err != nil {
			t.Fatalf("Failed to get player with error %v", err)
		}
		if !reflect.DeepEqual(actualPlayer, expectedPlayer) {
			t.Fatalf("Expected to get player %v, got %v", expectedPlayer, actualPlayer)
		}
	}
}
