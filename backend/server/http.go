/*
 * Copyright (c) Joseph Prichard 2023
 */

package server

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

type ErrorResp struct {
	Code      int    `json:"code"`
	Status    int    `json:"status"`
	ErrorDesc string `json:"errorDesc"`
}

func WriteError(w http.ResponseWriter, status int, errorDesc string) {
	resp := ErrorResp{Status: status, ErrorDesc: errorDesc}
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

func ReadJson[T any](r *http.Request, result *T) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return errors.New("Failed to read data from request body")
	}

	err = json.Unmarshal(body, result)
	if err != nil {
		return errors.New("Invalid format for request body")
	}
	return nil
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
	header := (*w).Header()
	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("Access-Control-Allow-Headers", "*")
}
