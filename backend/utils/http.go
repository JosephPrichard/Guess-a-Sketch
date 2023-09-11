package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorMsg struct {
	Code      int    `json:"code"`
	Status    int    `json:"status"`
	ErrorDesc string `json:"errorDesc"`
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

func EnableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}
