package modes

import (
	"bitbucket.org/hamologist/rps/game"
)

const rock = "rock"
const paper = "paper"
const scissors = "scissors"

// StandardGame defines the default RPS game with a standard (rock, paper, scissors) moveset
var StandardGame = game.Game{
	Moves: map[string]game.Move{
		rock: game.Move{
			Name:    rock,
			Defeats: []string{scissors},
		},
		paper: game.Move{
			Name:    paper,
			Defeats: []string{rock},
		},
		scissors: game.Move{
			Name:    scissors,
			Defeats: []string{paper},
		},
	},
	PreferredOrder: []string{rock, paper, scissors},
}
