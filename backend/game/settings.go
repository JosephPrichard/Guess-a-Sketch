package game

import (
	"fmt"
)

const (
	MinWordBank    = 10
	MinTimeLimit   = 15
	MaxTimeLimit   = 240
	MinPlayerLimit = 2
	MaxPlayerLimit = 12
	MaxTotalRounds = 6
)

type RoomSettings struct {
	PlayerLimit    int      `json:"playerLimit"`    // max players that can join room state
	TotalRounds    int      `json:"totalRounds"`    // total rounds for the game to go through
	TimeLimitSecs  int      `json:"timeLimitSecs"`  // time given for guessing each turn
	CustomWordBank []string `json:"customWordBank"` // custom words added in the bank by host
	SharedWordBank []string `json:"-"`              // reference to the shared word bank
	IsPublic       bool     `json:"isPublic"`       // whether the room is publicly accessible of not
}

func SettingsWithDefaults(settings *RoomSettings) {
	// set the default settings for zero valued fields
	if settings.PlayerLimit == 0 {
		settings.PlayerLimit = 8
	}
	if settings.TimeLimitSecs == 0 {
		settings.TimeLimitSecs = 45
	}
	if settings.TotalRounds == 0 {
		settings.TotalRounds = 3
	}
	if settings.CustomWordBank == nil {
		settings.CustomWordBank = make([]string, 0)
	}
}

func IsSettingsValid(settings RoomSettings) error {
	if len(settings.CustomWordBank) < MinWordBank {
		return fmt.Errorf("Word bank must have at least %d words", MinWordBank)
	}
	if settings.TimeLimitSecs < MinTimeLimit || settings.TimeLimitSecs > MaxPlayerLimit {
		return fmt.Errorf("Time limit must be between %d and %d seconds", MaxTimeLimit, MaxTimeLimit)
	}
	if settings.PlayerLimit < MinPlayerLimit || settings.PlayerLimit > MaxPlayerLimit {
		return fmt.Errorf("Games can only contain between %d and %d players", MaxPlayerLimit, MaxPlayerLimit)
	}
	if settings.TotalRounds > MaxTotalRounds {
		return fmt.Errorf("Games can only contain between %d and %d players", MaxPlayerLimit, MaxPlayerLimit)
	}
	return nil
}
