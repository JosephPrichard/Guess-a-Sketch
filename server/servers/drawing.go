/*
 * Copyright (c) Joseph Prichard 2024
 */

package servers

import (
	"github.com/jmoiron/sqlx"
	"guessthesketch/database"
	"net/http"
)

type DrawingServer struct {
	db *sqlx.DB
}

func NewDrawingServer(db *sqlx.DB) *DrawingServer {
	return &DrawingServer{db: db}
}

func (server *DrawingServer) GetDrawings(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	username := query.Get("username")

	drawings, err := database.GetDrawings(server.db, username)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Cache-Control", "max-age=1800")
	w.WriteHeader(http.StatusOK)
	WriteJson(w, drawings)
}
