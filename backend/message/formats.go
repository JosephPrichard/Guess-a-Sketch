package message

import (
	"encoding/json"
	"guessasketch/game"
	"guessasketch/utils"
	"log"
)

const (
	OptionsCode = 0
	StartCode   = 1
	TextCode    = 2
	DrawCode    = 3
	ChatCode    = 4
	FinishCode  = 5
	BeginCode   = 6
	JoinCode    = 7
	LeaveCode   = 8
)

type OptionsMsg = game.Options

type TextMsg struct {
	Text string
}

type DrawMsg = game.Circle

type ChatMsg = game.Chat

type BeginMsg struct {
	NextWord   string
	NextPlayer string
}

type FinishMsg struct {
	BeginMsg      *BeginMsg
	PrevPlayer    string
	GuessScoreInc int
}

type PlayerMsg struct {
	// ensures ordering of players on client and server are the same
	PlayerIndex int
	Player      string
}

type InputPayload struct {
	Code int
	Msg  json.RawMessage
}

type OutputPayload struct {
	Code int
	Msg  interface{}
}

func SendErrMsg(ch chan []byte, errorDesc string) {
	msg := utils.ErrorMsg{ErrorDesc: errorDesc}
	b, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to serialize error for ws message")
		return
	}
	ch <- b
}
