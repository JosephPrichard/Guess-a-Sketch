package server

import (
	"encoding/json"
	"guessasketch/store"
	"log"
	"net/http"
)

const (
	OptionsCode     = 0
	StartCode       = 1
	TextCode        = 2
	DrawCode        = 3
	ChatCode        = 4
	FinishCode      = 5
	BeginCode       = 6
	JoinCode        = 7
	LeaveCode       = 8
)

type OptionsMsg = store.Options

type TextMsg struct {
	Text string
}

type DrawMsg = store.Circle

type ChatMsg = store.Chat

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
	playerIndex int
	player      string
}

type ErrorMsg struct {
	Status    int
	ErrorDesc string
}

type InputPayload struct {
	Code int
	Msg  json.RawMessage
}

type OutputPayload struct {
	Code int
	Msg  interface{}
}

func SendErrResp(w http.ResponseWriter, msg ErrorMsg) {
	b, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to serialize error for http response")
		return
	}
	w.WriteHeader(msg.Status)
	w.Write(b)
}

func SendErrMsg(ch chan []byte, errorDesc string) {
	msg := ErrorMsg{ErrorDesc: errorDesc}
	b, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to serialize error for ws message")
		return
	}
	ch <- b
}
