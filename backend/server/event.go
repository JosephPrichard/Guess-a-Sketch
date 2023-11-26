/*
 * Copyright (c) Joseph Prichard 2023
 */

package server

import (
	"github.com/jmoiron/sqlx"
	"guessasketch/database"
	"guessasketch/game"
)

type EventServer struct {
	db *sqlx.DB
}

func NewEventServer(db *sqlx.DB) *EventServer {
	return &EventServer{db}
}

func (server EventServer) OnShutdown(results []game.GameResult) {
	database.UpdateStats(server.db, results)
}

func (server EventServer) OnSaveDrawing(drawing game.Drawing) error {
	return database.SaveDrawing(server.db, drawing)
}
