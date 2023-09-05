package database

type Player struct {
	ID              string
	Username        string
	Points          uint32
	Wins            uint32
	WordsGuessed    uint32
	DrawingsGuessed uint32
	Avatar          string
}