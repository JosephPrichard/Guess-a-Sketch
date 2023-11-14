package server

import (
	"guessasketch/database"
	"guessasketch/utils"
	"net/http"

	"github.com/jmoiron/sqlx"
)

type PlayerServer struct {
	db *sqlx.DB
}

func NewPlayerServer(db *sqlx.DB) *PlayerServer {
	return &PlayerServer{db}
}

func (server *PlayerServer) Get(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	id := query.Get("id")

	var player database.Player
	err := database.GetPlayer(server.db, &player, id)
	if err != nil {
		resp := utils.ErrorResp{Status: http.StatusNotFound, ErrorDesc: "Failed to get player data"}
		utils.WriteError(w, resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	utils.WriteJson(w, player)
}

func (server *PlayerServer) Leaderboard(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	sort := query.Get("sort")

	var players []database.Player
	err := database.GetLeaderboard(server.db, players, 50, sort)
	if err != nil {
		resp := utils.ErrorResp{Status: http.StatusNotFound, ErrorDesc: "Failed to get the leaderboard"}
		utils.WriteError(w, resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	utils.WriteJson(w, players)
}
