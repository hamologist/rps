package modes

import (
	"bitbucket.org/hamologist/rps/game"
	"testing"
)

func TestRockBeatsScissors(t *testing.T) {
	result, err := StandardGame.Play(rock, scissors)

	if result != game.GameStatePlayerOneWins {
		t.Fatal("Rock failed to beat scissors")
	}

	if err != nil {
		t.Fatalf("Game state should not have caused an error: %q", err)
	}
}

func TestScissorsBeatsPaper(t *testing.T) {
	result, err := StandardGame.Play(scissors, paper)

	if result != game.GameStatePlayerOneWins {
		t.Fatal("Scissors failed to beat paper")
	}

	if err != nil {
		t.Fatalf("Game state should not have caused an error: %q", err)
	}
}

func TestPaperBeatsRock(t *testing.T) {
	result, err := StandardGame.Play(paper, rock)

	if result != game.GameStatePlayerOneWins {
		t.Fatal("Paper failed to beat rock")
	}

	if err != nil {
		t.Fatalf("Game state should not have caused an error: %q", err)
	}
}
