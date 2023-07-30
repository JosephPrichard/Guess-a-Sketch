package server

import (
	"encoding/json"
	"guessasketch/store"
	"log"
	"net/http"
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

type ErrorMsg struct {
	Status    int
	ErrorDesc string
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

func SendErrMsg(ch chan string, errorDesc string) {
	msg := ErrorMsg{ErrorDesc: errorDesc}
	b, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to serialize error for ws message")
		return
	}
	ch <- string(b)
}
