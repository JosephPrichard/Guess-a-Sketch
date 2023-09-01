package store

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"
)

const (
	Lobby   = 0
	Playing = 1
	Post    = 2
)

type Chat struct {
	Player        string
	Text          string
	GuessScoreInc int // if this is larger than 0, player guessed correctly
}

type Room struct {
	Code           string         // code of the room that uniquely identifies it
	CurrRound      int            // the current round
	Players        []string       // stores all players in the order they joined in
	ScoreBoard     map[string]int // maps players to scores
	ChatLog        []Chat         // stores the chat log
	Stage          int            // the current stage the room is (upports concurrent operations)
	sharedWordBank []string       // reference to the shared wordbank
	Settings       RoomSettings   // settings for the room set before game starts
	Game           Game           // if the game state is nil, no game is being played
}

func NewRoom(code string, sharedWordBank []string) Room {
	return Room{
		Code:           code,
		Players:        make([]string, 0),
		ScoreBoard:     make(map[string]int),
		ChatLog:        make([]Chat, 0),
		sharedWordBank: sharedWordBank,
		Settings:       NewSettings(),
		Game:           NewGame(),
	}
}

func (room *Room) GetCurrPlayer() string {
	if room.Game.CurrPlayerIndex < 0 {
		return ""
	}
	return room.Players[room.Game.CurrPlayerIndex]
}

func (room *Room) PlayerIsNotHost(player string) bool {
	return len(room.Players) < 1 || room.Players[0] != player
}

func (room *Room) ToMessage() []byte {
	b, err := json.Marshal(room)
	if err != nil {
		log.Printf(err.Error())
		return []byte{}
	}
	return b
}

func (room *Room) CanJoin(player string) bool {
	_, exists := room.ScoreBoard[player]
	return !exists
}

func (room *Room) PlayerIndex(playerToFind string) int {
	// find player in the slice
	index := -1
	for i, player := range room.Players {
		if player == playerToFind {
			index = i
			break
		}
	}
	return index
}

func (room *Room) Join(player string) {
	room.Players = append(room.Players, player)
	room.ScoreBoard[player] = 0
}

func (room *Room) Leave(playerToLeave string) int {
	index := room.PlayerIndex(playerToLeave)
	if index == -1 {
		// player doesn't exist in players slice - player never joined
		return -1
	}
	// delete player from the slice by creating a new slice without the index
	room.Players = append(room.Players[:index], room.Players[index+1:]...)
	return index
}

// starts the game and returns a snapshot of the settings used to start the game
func (room *Room) StartGame() error {
	room.Stage = Playing

	room.Game.ClearGuessers()
	room.Game.ClearCanvas()

	room.setNextWord()
	room.cycleCurrPlayer()

	room.Game.ResetStartTime()
	return nil
}

func (room *Room) FinishGame() {
	room.Stage = Post
}

func (room *Room) setNextWord() {
	// pick a new word from the shared or custom word bank
	index := rand.Intn(len(room.sharedWordBank) + len(room.Settings.customWordBank))
	if index < len(room.sharedWordBank) {
		room.Game.CurrWord = room.sharedWordBank[index]
	} else {
		room.Game.CurrWord = room.Settings.customWordBank[index]
	}
}

func (room *Room) cycleCurrPlayer() {
	// go to the next player, circle back around when we reach the end
	room.Game.CurrPlayerIndex += 1
	if room.Game.CurrPlayerIndex >= len(room.Players) {
		room.Game.CurrPlayerIndex = 0
		room.CurrRound += 1
	}
}

// handlers a player's guess and returns the increase in the score of player due to the guess
func (room *Room) OnGuess(player string, text string) int {
	// nothing happens if a player guesses when game is not in session
	if room.Stage != Playing {
		return 0
	}
	// current player cannot make a guess
	if player == room.GetCurrPlayer() {
		return 0
	}
	// check whether the text is a correct guess or not, if not, do not increase the score
	if !room.Game.ContainsCurrWord(text) {
		return 0
	}
	// cannot increase score of player if they already guessed
	if room.Game.guessers[player] {
		return 0
	}

	// calculate the score changes for successful guess
	timeSinceStartSecs := time.Now().Unix() - room.Game.startTimeSecs
	timeLimitSecs := room.Settings.TimeLimitSecs

	scoreInc := (timeLimitSecs-int(timeSinceStartSecs))/timeLimitSecs*400 + 50
	room.ScoreBoard[player] += scoreInc

	room.Game.SetGuesser(player)
	return scoreInc
}

func (room *Room) OnResetScoreInc() int {
	scoreInc := room.Game.CalcResetScore()
	room.ScoreBoard[room.GetCurrPlayer()] += scoreInc
	return scoreInc
}

func (room *Room) AddChat(chat Chat) {
	room.ChatLog = append(room.ChatLog, chat)
}
