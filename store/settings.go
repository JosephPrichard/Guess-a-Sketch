package store

type Options struct {
	PlayerLimit   int
	TimeLimitSecs int
	WordBank      []string
	TotalRounds   int
}

type RoomSettings struct {
	playerLimit    int      // max players that can join room state
	TotalRounds    int      // total rounds for the game to go through
	TimeLimitSecs  int      // time given for guessing each turn
	customWordBank []string // custom words added in the bank by host
}

func NewSettings() RoomSettings {
	return RoomSettings{
		playerLimit:    8,
		TimeLimitSecs:  45,
		TotalRounds:    3,
		customWordBank: make([]string, 0),
	}
}

func (settings *RoomSettings) UpdateSettings(options *Options) {
	if options.PlayerLimit != 0 {
		settings.playerLimit = options.PlayerLimit
	}
	if options.TotalRounds != 0 {
		settings.TotalRounds = options.TotalRounds
	}
	if options.TimeLimitSecs != 0 {
		settings.TimeLimitSecs = options.TimeLimitSecs
	}
	if len(options.WordBank) != 0 {
		settings.customWordBank = options.WordBank
	}
}
