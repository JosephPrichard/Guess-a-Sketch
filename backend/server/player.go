package server

import (
	"encoding/json"
	"guessasketch/database"
	"log"
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
	database.GetPlayer(server.db, &player, id)

	b, err := json.Marshal(player)
	if err != nil {
		log.Printf("Failed to serialize player response")
		return
	}

	w.WriteHeader(200)
	w.Write(b)
}

func (server *PlayerServer) Leaderboard(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	sort := query.Get("sort")

	players := []database.Player{}
	database.GetLeaderboard(server.db, players, 50, sort)

	b, err := json.Marshal(players)
	if err != nil {
		log.Printf("Failed to serialize leaderboard response")
		return
	}

	w.WriteHeader(200)
	w.Write(b)
}
