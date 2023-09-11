package game

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"time"
)

const (
	Lobby   = 0
	Playing = 1
	Post    = 2
)

type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Chat struct {
	Player        Player `json:"player"`
	Text          string `json:"text"`
	GuessScoreInc int    `json:"guessScoreInc"` // if this is larger than 0, player guessed correctly
}

type Room struct {
	Code           string         `json:"code"`       // code of the room that uniquely identifies it
	CurrRound      int            `json:"currRound"`  // the current round
	Players        []Player       `json:"players"`    // stores all players in the order they joined in
	ScoreBoard     map[string]int `json:"scoreBoard"` // maps player IDs to scores
	ChatLog        []Chat         `json:"chatLog"`    // stores the chat log
	Stage          int            `json:"stage"`      // the current stage the room is
	Settings       RoomSettings   `json:"settings"`   // settings for the room set before game starts
	Turn           GameTurn       `json:"turn"`       // stores the current game turn
	sharedWordBank []string       // reference to the shared wordbank
}

func NewRoom(code string, sharedWordBank []string) Room {
	return Room{
		Code:           code,
		Players:        make([]Player, 0),
		ScoreBoard:     make(map[string]int),
		ChatLog:        make([]Chat, 0),
		sharedWordBank: sharedWordBank,
		Settings:       NewRoomSettings(),
		Turn:           NewGameTurn(),
	}
}

func (room *Room) GetCurrPlayer() *Player {
	if room.Turn.CurrPlayerIndex < 0 {
		return &Player{}
	}
	return &room.Players[room.Turn.CurrPlayerIndex]
}

func (room *Room) PlayerIsNotHost(player Player) bool {
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

func (room *Room) PlayerIndex(playerToFind Player) int {
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

func (room *Room) Join(player Player) error {
	_, exists := room.ScoreBoard[player.ID]
	if exists {
		return errors.New("Player cannot join, username is already in use")
	}
	if len(room.Players) >= room.Settings.PlayerLimit {
		return errors.New("Player cannot join, room is at player limit")
	}
	room.Players = append(room.Players, player)
	room.ScoreBoard[player.ID] = 0
	return nil
}

func (room *Room) Leave(playerToLeave Player) int {
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
func (room *Room) StartGame() {
	room.Stage = Playing

	room.Turn.ClearGuessers()
	room.Turn.ClearCanvas()

	room.setNextWord()
	room.cycleCurrPlayer()

	room.Turn.ResetStartTime()
}

func (room *Room) FinishGame() {
	room.Stage = Post
}

func (room *Room) setNextWord() {
	// pick a new word from the shared or custom word bank
	index := rand.Intn(len(room.sharedWordBank) + len(room.Settings.CustomWordBank))
	if index < len(room.sharedWordBank) {
		room.Turn.CurrWord = room.sharedWordBank[index]
	} else {
		room.Turn.CurrWord = room.Settings.CustomWordBank[index]
	}
}

func (room *Room) cycleCurrPlayer() {
	// go to the next player, circle back around when we reach the end
	room.Turn.CurrPlayerIndex += 1
	if room.Turn.CurrPlayerIndex >= len(room.Players) {
		room.Turn.CurrPlayerIndex = 0
		room.CurrRound += 1
	}
}

// handlers a player's guess and returns the increase in the score of player due to the guess
func (room *Room) OnGuess(player Player, text string) int {
	// nothing happens if a player guesses when game is not in session
	if room.Stage != Playing {
		return 0
	}
	// current player cannot make a guess
	if player.ID == room.GetCurrPlayer().ID {
		return 0
	}
	// check whether the text is a correct guess or not, if not, do not increase the score
	if !room.Turn.ContainsCurrWord(text) {
		return 0
	}
	// cannot increase score of player if they already guessed
	if room.Turn.guessers[player.ID] {
		return 0
	}

	// calculate the score changes for successful guess
	timeSinceStartSecs := time.Now().Unix() - room.Turn.startTimeSecs
	timeLimitSecs := room.Settings.TimeLimitSecs

	scoreInc := (timeLimitSecs-int(timeSinceStartSecs))/timeLimitSecs*400 + 50
	room.ScoreBoard[player.ID] += scoreInc

	room.Turn.SetGuesser(player.ID)
	return scoreInc
}

func (room *Room) OnResetScoreInc() int {
	scoreInc := room.Turn.CalcResetScore()
	room.ScoreBoard[room.GetCurrPlayer().ID] += scoreInc
	return scoreInc
}

func (room *Room) AddChat(chat Chat) {
	room.ChatLog = append(room.ChatLog, chat)
}
