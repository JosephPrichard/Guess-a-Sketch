package server

import (
	"encoding/json"
	"log"
	"net/http"
)

const MARSHAL_ERR_MSG = "Failed to unmarshal/mashall input data"

type OptionsMsg struct {
	playerLimit   int
	timeLimitSecs int
	wordBank      []string
}

type TextMsg struct {
	Text string
}

type DrawMsg struct {
	X      uint16
	Y      uint16
	Color  uint8
	Radius uint8
}

type ChatMsg struct {
	Player        string
	Text          string
	GuessScoreInc int `json:"scoreInc,omitempty"`
}

type BeginMsg struct {
	NextWord   string
	NextPlayer string
}

type FinishMsg struct {
	*BeginMsg
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
