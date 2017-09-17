package slack

import (
	"fmt"
	"os"

	"github.com/nlopes/slack"

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

var (
	// DefaultGameServer is the default game server created by the slack package.
	// Used as an optional game server that supports everything needed for integrating with slack.
	DefaultGameServer *server.GameServer

	// OAuthToken is the access token needed for interacting with the Slack API.
	OAuthToken string

	// API is a package global for working with the Slack API
	API *slack.Client
)

func createSlackData(channelName, challengerName, targetName string) map[string]string {
	return map[string]string{
		"channelName":    channelName,
		"challengerName": challengerName,
		"targetName":     targetName,
	}
}

func validSlackData(data map[string]string) bool {
	if _, ok := data["channelName"]; !ok {
		return false
	}

	if _, ok := data["challengerName"]; !ok {
		return false
	}

	if _, ok := data["targetName"]; !ok {
		return false
	}

	return true
}

func init() {
	rpsGame := os.Getenv("RPS_GAME")
	OAuthToken = os.Getenv("RPS_SLACK_OAUTH")

	if registeredGame, ok := modes.RegisteredGames[rpsGame]; ok {
		DefaultGameServer = server.NewGameServer(registeredGame)
	} else {
		DefaultGameServer = server.NewGameServer(modes.StandardGame)
	}

	if OAuthToken == "" {
		fmt.Print(
			"No slack OAuth token was provided, application will have limited slack support.\n" +
				"Please consider providing an OAuth token using the \"RPS_SLACK_OAUTH\" env variable.\n",
		)
	}
	API = slack.New(OAuthToken)

	registerRoutes(DefaultGameServer)
}
