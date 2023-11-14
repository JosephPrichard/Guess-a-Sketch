package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResp struct {
	Code      int    `json:"code"`
	Status    int    `json:"status"`
	ErrorDesc string `json:"errorDesc"`
}

func WriteError(w http.ResponseWriter, resp ErrorResp) {
	b, err := json.Marshal(resp)
	if err != nil {
		log.Println("Failed to serialize error for http response")
		return
	}
	w.WriteHeader(resp.Status)
	_, err = w.Write(b)
	if err != nil {
		log.Println("Failed to write body to response")
		return
	}
}

func WriteJson(w http.ResponseWriter, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Println("Failed to marshal json response")
		return
	}
	_, err = w.Write(b)
	if err != nil {
		log.Println("Failed to write body as response")
		return
	}
}

func EnableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}
