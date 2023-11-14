package message

import (
	"encoding/json"
	"guessasketch/game"
	"guessasketch/utils"
	"log"
)

const (
	StartCode   = 1
	TextCode    = 2
	DrawCode    = 3
	ChatCode    = 4
	FinishCode  = 5
	BeginCode   = 6
	JoinCode    = 7
	LeaveCode   = 8
	TimeoutCode = 9
)

type StartMsg = game.RoomSettings

type TextMsg struct {
	Text string `json:"text"`
}

type DrawMsg = game.Circle

type ChatMsg = game.Chat

type BeginMsg struct {
	NextWord        string `json:"nextWord"`
	NextPlayerIndex int    `json:"nextPlayerIndex"`
}

type FinishMsg struct {
	BeginMsg     *BeginMsg `json:"beginMsg"`
	DrawScoreInc int       `json:"drawScoreInc"`
}

type PlayerMsg struct {
	PlayerIndex int         `json:"playerIndex"` // ensures ordering of players on client and server are the same
	Player      game.Player `json:"player"`
}

type InputPayload struct {
	Code int             `json:"code"`
	Msg  json.RawMessage `json:"msg"`
}

type OutputPayload struct {
	Code int         `json:"code"`
	Msg  interface{} `json:"msg"`
}

func SendErrMsg(ch chan []byte, errorDesc string) {
	msg := utils.ErrorResp{ErrorDesc: errorDesc}
	b, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to serialize error for ws message")
		return
	}
	ch <- b
}
