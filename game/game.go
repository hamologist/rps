package game

import (
	"fmt"
)

// GameStatePlayerOneWins state for when player one wins.
const GameStatePlayerOneWins = "player one"

// GameStatePlayerTwoWins state for when player two wins.
const GameStatePlayerTwoWins = "player two"

// GameStateDraw state for when both players make the same move.
const GameStateDraw = "draw"

// GameStateError state for when a player makes an invalid move.
const GameStateError = "error"

// Move is used for defining moves and their interactions with other valid GameMove.
type Move struct {
	Name    string
	Defeats []string
}

// Game is used to define the rules of a game of RPS.
type Game struct {
	Moves          map[string]Move // The collection used to determine valid moves in a game of RPS.
	PreferredOrder []string        // The preferred order to use when displaying moves to the user.
}

// Play defines the basic two player interaction of RPS.
func (game *Game) Play(playerOneMove string, playerTwoMove string) (string, error) {
	moves := game.Moves

	if _, ok := moves[playerOneMove]; !ok {
		return GameStateError, fmt.Errorf("Player One provided an invalid move: %v", playerOneMove)
	}

	if _, ok := moves[playerTwoMove]; !ok {
		return GameStateError, fmt.Errorf("Player Two provided an invalid move: %v", playerOneMove)
	}

	if playerOneMove == playerTwoMove {
		return GameStateDraw, nil
	}

	for _, v := range moves[playerOneMove].Defeats {
		if v == playerTwoMove {
			return GameStatePlayerOneWins, nil
		}
	}

	return GameStatePlayerTwoWins, nil
}
