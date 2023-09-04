package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

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

func EnableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func Session(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("session")

	var newValue string
	if cookie != nil && err == nil {
		newValue = cookie.Value
	} else {
		newValue = AlphaNumeric(64)
	}

	cookie = &http.Cookie{
		Name:    "session",
		Value:   newValue,
		Expires: time.Now().Add(time.Hour * 24 * 7),
	}
	http.SetCookie(w, cookie)
	return cookie.Value
}
