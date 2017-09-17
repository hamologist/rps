package game

import (
	"testing"
)

const winningMove = "win"
const losingMove = "lose"
const invalidMove = "invalid"

func TestPlayerOneWins(t *testing.T) {
	game := createMockGame()
	result, err := game.Play(winningMove, losingMove)

	if result != GameStatePlayerOneWins {
		t.Fatal("Player One failed to beat Player Two")
	}

	if err != nil {
		t.Fatalf("Game state should not have caused an error: %q", err)
	}
}

func TestPlayerTwoWins(t *testing.T) {
	game := createMockGame()
	result, err := game.Play(losingMove, winningMove)

	if result != GameStatePlayerTwoWins {
		t.Fatal("Player One failed to beat Player Two")
	}

	if err != nil {
		t.Fatalf("Game state should not have caused an error: %q", err)
	}
}

func TestPlayerOneHasAnInvalidMove(t *testing.T) {
	game := createMockGame()
	result, err := game.Play(invalidMove, winningMove)

	if result != GameStateError {
		t.Fatal("Game::Play failed to return an error state")
	}

	if err == nil {
		t.Fatalf("Game state should have caused an error: %q", err)
	}
}

func TestPlayerTwoHasAnInvalidMove(t *testing.T) {
	game := createMockGame()
	result, err := game.Play(winningMove, invalidMove)

	if result != GameStateError {
		t.Fatal("Game::Play failed to return an error state")
	}

	if err == nil {
		t.Fatalf("Game state should have caused an error: %q", err)
	}
}

func TestGameEndsInDraw(t *testing.T) {
	game := createMockGame()
	result, err := game.Play(winningMove, winningMove)

	if result != GameStateDraw {
		t.Fatal("Game failed to end in a draw")
	}

	if err != nil {
		t.Fatalf("Game state should not have caused an error: %q", err)
	}
}

func createMockGame() Game {
	return Game{
		Moves: map[string]Move{
			winningMove: Move{winningMove, []string{losingMove}},
			losingMove:  Move{losingMove, []string{}},
		},
		PreferredOrder: []string{winningMove, losingMove},
	}
}
