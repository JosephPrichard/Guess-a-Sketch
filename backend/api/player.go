package api

import "net/http"

type PlayerServer struct {
}

func NewPlayerServer() *PlayerServer {
	return &PlayerServer{}
}

func (server *PlayerServer) Get(w http.ResponseWriter, r *http.Request) {

}

func (server *PlayerServer) Leaderboard(w http.ResponseWriter, r *http.Request) {

}
