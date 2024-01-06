/*
 * Copyright (c) Joseph Prichard 2023
 */

package database

type Player struct {
	ID              string `db:"id"`
	Username        string `db:"username"`
	Points          uint32 `db:"points"`
	Wins            uint32 `db:"wins"`
	WordsGuessed    uint32 `db:"words_guessed"`
	DrawingsGuessed uint32 `db:"drawings_guessed"`
	//Avatar          string `db:"id"`
}

type Drawing struct {
	ID        int8   `db:"id"`
	CreatedBy string `db:"created_by"`
	SavedBy   string `db:"saved_by"`
	Signature string `db:"signature"`
}
