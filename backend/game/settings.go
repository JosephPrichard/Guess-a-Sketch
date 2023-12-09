/*
 * Copyright (c) Joseph Prichard 2023
 */

package game

import (
	"fmt"
)

const (
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

// applies default settings to preexisting settings struct any zero value field
func SettingsWithDefaults(settings *RoomSettings) {
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
	if settings.TimeLimitSecs < MinTimeLimit || settings.TimeLimitSecs > MaxTimeLimit {
		return fmt.Errorf("Time limit must be between %d and %d seconds", MinTimeLimit, MaxTimeLimit)
	}
	if settings.PlayerLimit < MinPlayerLimit || settings.PlayerLimit > MaxPlayerLimit {
		return fmt.Errorf("Games can only contain between %d and %d players", MinPlayerLimit, MaxPlayerLimit)
	}
	if settings.TotalRounds > MaxTotalRounds || settings.TotalRounds < 0 {
		return fmt.Errorf("Games can only have between 0 and %d rounds", MaxTotalRounds)
	}
	return nil
}
