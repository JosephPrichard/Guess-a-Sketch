/*
 * Copyright (c) Joseph Prichard 2023
 */

package servers

import (
	"guessthesketch/database"
	"net/http"

	"github.com/jmoiron/sqlx"
)

type PlayerServer struct {
	db         *sqlx.DB
	authServer *AuthServer
}

func NewPlayerServer(db *sqlx.DB, authServer *AuthServer) *PlayerServer {
	return &PlayerServer{db: db, authServer: authServer}
}

func (server *PlayerServer) Get(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	username := query.Get("username")

	var player database.Player
	err := database.GetPlayer(server.db, &player, username)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if &player == nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	w.Header().Set("Cache-Control", "max-age=1800")
	w.WriteHeader(http.StatusOK)
	WriteJson(w, player)
}

func (server *PlayerServer) Leaderboard(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	sort := query.Get("sort")

	players, err := database.GetLeaderboard(server.db, 50, sort)
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	w.Header().Set("Cache-Control", "max-age=3600")
	w.WriteHeader(http.StatusOK)
	WriteJson(w, players)
}
