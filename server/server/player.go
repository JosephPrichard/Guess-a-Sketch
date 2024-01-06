/*
 * Copyright (c) Joseph Prichard 2023
 */

package server

import (
	"github.com/google/uuid"
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
	id := query.Get("id")

	var player *database.Player
	err := database.GetPlayer(server.db, player, id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if player == nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	w.Header().Set("Cache-Control", "max-age=1800")
	w.WriteHeader(http.StatusOK)
	WriteJson(w, player)
}

func (server *PlayerServer) NewPlayer(name string) (*database.Player, error) {
	player := database.Player{ID: uuid.New().String(), Username: name}
	err := database.CreatePlayer(server.db, player)
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func (server *PlayerServer) Leaderboard(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	sort := query.Get("sort")

	players := make([]database.Player, 0)
	err := database.GetLeaderboard(server.db, players, 50, sort)
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	w.Header().Set("Cache-Control", "max-age=3600")
	w.WriteHeader(http.StatusOK)
	WriteJson(w, players)
}
