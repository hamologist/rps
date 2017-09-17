package slack

import (
	"os"

	"bitbucket.org/hamologist/rps/game/modes"
	"bitbucket.org/hamologist/rps/server"
)

const (
	inChannelResponse          = "in_channel"
	ephemeralResponse          = "ephemeral"
	defaultAcceptCommandName   = "rps-accept"
	gameSessionInitiatedStatus = "initiated"
	gameSessionAcceptedStatus  = "accepted"
)

// DefaultGameServer is the default game server created by the slack package.
// This is an optional game server that supports everything needed for integrating with slack.
var DefaultGameServer *server.GameServer

func createSlackData(status, challengerResponseURL, targetResponseURL string) map[string]string {
	return map[string]string{
		"status":                status,
		"challengerResponseUrl": challengerResponseURL,
		"targetResponseUrl":     targetResponseURL,
	}
}

func validSlackData(data map[string]string) bool {
	if _, ok := data["status"]; !ok {
		return false
	}

	if _, ok := data["challengerResponseUrl"]; !ok {
		return false
	}

	if _, ok := data["targetResponseUrl"]; !ok {
		return false
	}

	return true
}

func init() {
	rpsGame := os.Getenv("RPS_GAME")

	if registeredGame, ok := modes.RegisteredGames[rpsGame]; ok {
		DefaultGameServer = server.NewGameServer(registeredGame)
	} else {
		DefaultGameServer = server.NewGameServer(modes.StandardGame)
	}

	registerRoutes(DefaultGameServer)
}
