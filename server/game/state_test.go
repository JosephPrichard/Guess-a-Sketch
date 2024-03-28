/*
 * Copyright (c) Joseph Prichard 2023
 */

package game

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"github.com/google/uuid"
	"reflect"
	"testing"
)

func TestState_Join(t *testing.T) {
	state := NewGameState("123", MockSettings())

	player1 := Player{ID: uuid.New(), present: true}
	player2 := Player{ID: uuid.New(), present: true}

	err := state.Join(player1)
	if err != nil {
		t.Fatalf("Player1 failed to join %v", err)
	}
	err = state.Join(player2)
	if err != nil {
		t.Fatalf("Player2 failed to join %v", err)
	}

	expectedScoreBoard := map[uuid.UUID]Score{player1.ID: {}, player2.ID: {}}
	if !reflect.DeepEqual(expectedScoreBoard, state.scoreBoard) {
		t.Fatalf("Scoreboard didn't contain expected expectedPlayers and scores")
	}

	expectedPlayers := []Player{player1, player2}
	if !reflect.DeepEqual(expectedPlayers, state.players) {
		t.Fatalf("players slice didn't contain expected players")
	}
}

func TestState_Join_Duplicate(t *testing.T) {
	state := NewGameState("123", MockSettings())

	player1 := Player{ID: uuid.New(), present: true}
	player2 := Player{ID: player1.ID, present: true}

	err := state.Join(player1)
	if err != nil {
		t.Fatalf("Player1 failed to join %v", err)
	}
	err = state.Join(player2)
	if err != nil {
		t.Fatalf("Player1 failed to join %v", err)
	}

	expectedPlayers := []Player{player2}
	if !reflect.DeepEqual(expectedPlayers, state.players) {
		t.Fatalf("players slice didn't contain expected players")
	}
}

func TestState_Leave(t *testing.T) {
	state := NewGameState("123", MockSettings())

	player1 := Player{ID: uuid.New(), present: true}
	player2 := Player{ID: uuid.New(), present: true}
	state.players = []Player{player1, player2}

	leaveIndex := state.Leave(player2)
	if leaveIndex != 1 {
		t.Fatalf("Expected %d playerIndex got %d", 1, leaveIndex)
	}

	if !reflect.DeepEqual([]Player{player1, {ID: player2.ID, present: false}}, state.players) {
		t.Fatalf("player2 should be marked as not present after leaving")
	}
	if !reflect.DeepEqual([]Player{player1}, state.Players()) {
		t.Fatalf("Only player1 should be present after player2 leaves")
	}
}

func TestState_Leave_NotJoined(t *testing.T) {
	state := NewGameState("123", MockSettings())

	player1 := Player{ID: uuid.New(), present: true}
	state.players = []Player{player1}

	leaveIndex := state.Leave(Player{ID: uuid.New()})
	if leaveIndex != -1 {
		t.Fatalf("Expected %d playerIndex got %d", -1, leaveIndex)
	}
}

func TestState_StartGame(t *testing.T) {
	settings := MockSettings()
	state := NewGameState("123", settings)

	player := Player{ID: uuid.New()}
	_ = state.Join(player)

	state.StartGame()

	if state.stage != Playing {
		t.Fatalf("Failed to set the state stage to Playing")
	}
	if state.currRound != 1 {
		t.Fatalf("Failed to advance the round")
	}
	if state.turn.currWord == "" {
		t.Fatalf("Failed pick a new random current word")
	}
}

func TestState_TryGuess(t *testing.T) {
	state := NewGameState("123", MockSettings())

	state.stage = Playing
	state.turn.currWord = "quick"
	guesser := Player{ID: uuid.New()}
	state.players = []Player{{ID: uuid.New()}, guesser}
	state.turn.currPlayerIndex = 0

	if state.TryGuess(guesser, "the QUICK brown fox").GuessPointsInc <= 0 {
		t.Fatalf("Guess score increment to be at least 0")
	}

	expectedGuessers := map[uuid.UUID]bool{guesser.ID: true}
	if !reflect.DeepEqual(expectedGuessers, state.turn.guessers) {
		t.Fatalf("Expected guessing player to be set as a guessers")
	}

	guesserScore, ok := state.scoreBoard[guesser.ID]
	if !ok || guesserScore.words != 1 {
		t.Fatalf("Scoreboard didn't contain expected a properly updated score for the guesser")
	}
}

func TestState_TryGuess_WrongWord(t *testing.T) {
	state := NewGameState("123", MockSettings())

	state.stage = Playing
	state.turn.currWord = "fast"
	state.players = []Player{{ID: uuid.New()}}
	state.turn.currPlayerIndex = 0

	if state.TryGuess(Player{ID: uuid.New()}, "the quick brown fox").GuessPointsInc != 0 {
		t.Fatalf("Guess should be unsuccessful due to wrong word")
	}
}

func TestState_TryGuess_IsCurrPlayer(t *testing.T) {
	state := NewGameState("123", MockSettings())

	state.stage = Playing
	state.turn.currWord = "quick"
	player1 := Player{ID: uuid.New()}
	state.players = []Player{player1}
	state.turn.currPlayerIndex = 0

	if state.TryGuess(player1, "the quick brown fox").GuessPointsInc != 0 {
		t.Fatalf("Guess should be unsuccessful due guesser is current player")
	}
}

func TestState_TryGuess_NoDoubleGuess(t *testing.T) {
	state := NewGameState("123", MockSettings())

	state.stage = Playing
	state.turn.currWord = "quick"
	state.players = []Player{{ID: uuid.New()}}
	state.turn.currPlayerIndex = 0

	player := Player{ID: uuid.New()}
	_ = state.TryGuess(player, "the quick brown fox")
	if state.TryGuess(player, "the quick brown fox").GuessPointsInc != 0 {
		t.Fatalf("Guess should be unsuccessful due to duplcate guess")
	}
}

func TestState_CreateGameResult(t *testing.T) {
	state := NewGameState("123", MockSettings())

	state.scoreBoard = map[uuid.UUID]Score{
		uuid.New(): {Points: 100, words: 1, drawings: 2},
		uuid.New(): {Points: 200, words: 2, drawings: 2},
		uuid.New(): {Points: 250, words: 3, drawings: 2},
	}

	results := state.CreateGameResults()

	if !results[0].Win {
		t.Fatalf("The top game results should be a win")
	}
	if results[0].Points != 250 || results[1].Points != 200 || results[2].Points != 100 {
		t.Fatalf("Results need to be sorted in order of points")
	}
}

func TestState_EncodeCanvas(t *testing.T) {
	state := NewGameState("123", MockSettings())
	state.turn.canvas = []Circle{
		{Color: 4, Radius: 3, X: 2, Y: 1, Connected: true},
		{Color: 5, X: 1, Y: 2, Connected: false}}

	s := state.EncodeCanvas()
	t.Logf("Canvas as string %s", s)

	var buf bytes.Buffer
	base64Decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("Error reading %v", err)
	}

	_, err = buf.Write(base64Decoded)
	if err != nil {
		t.Fatalf("Error reading %v", err)
	}

	var canvas []Circle
	err = binary.Read(&buf, binary.LittleEndian, canvas)
	if err != nil {
		t.Fatalf("Error reading %v", err)
	}

	if reflect.DeepEqual(state.turn.canvas, canvas) {
		t.Fatalf("Canvas is not the same after encoding then decoding - binary serialization does not work")
	}
}
