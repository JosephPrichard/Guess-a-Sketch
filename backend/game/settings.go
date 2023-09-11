package game

type RoomSettings struct {
	PlayerLimit    int      `json:"playerLimit"`    // max players that can join room state
	TotalRounds    int      `json:"totalRounds"`    // total rounds for the game to go through
	TimeLimitSecs  int      `json:"timeLimitSecs"`  // time given for guessing each turn
	CustomWordBank []string `json:"customWordBank"` // custom words added in the bank by host
}

func NewRoomSettings() RoomSettings {
	return RoomSettings{
		PlayerLimit:    8,
		TimeLimitSecs:  45,
		TotalRounds:    3,
		CustomWordBank: make([]string, 0),
	}
}