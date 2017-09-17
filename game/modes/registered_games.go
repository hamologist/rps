package modes

import (
	"bitbucket.org/hamologist/rps/game"
)

// RegisteredGames defines all available game modes.
var RegisteredGames = map[string]game.Game{
	"standard": StandardGame,
}
