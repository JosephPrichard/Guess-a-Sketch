/*
 * Copyright (c) Joseph Prichard 2023
 */

package database

type Player struct {
	ID              string `db:"id"`
	Username        string `db:"username"`
	Points          int    `db:"points"`
	Wins            int    `db:"wins"`
	WordsGuessed    int    `db:"words_guessed"`
	DrawingsGuessed int    `db:"drawings_guessed"`
}

type Drawing struct {
	ID        string `db:"id"`
	CreatedBy string `db:"created_by"`
	SavedBy   string `db:"saved_by"`
	Signature string `db:"signature"`
}
