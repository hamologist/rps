package slack

import (
	"net/http"
	"os"
	"time"

	"bitbucket.org/hamologist/rps-go/game"
	"bitbucket.org/hamologist/rps-go/game/modes"
	"github.com/satori/go.uuid"
)

const (
	challengeAction            = "challenge"
	acceptAction               = "accept"
	inChannelResponse          = "in_channel"
	ephemeralResponse          = "ephemeral"
	defaultCommandName         = "rps"
	gameSessionInitiatedStatus = "initiated"
	gameSessionAcceptedStatus  = "accepted"
)

var (
	GameServer *gameServer
)

type gameServer struct {
	ServeMux            *http.ServeMux
	GameSessionsManager *gameSessionsManager
	Game                game.Game
}

type slackBody struct {
	Token        string `schema:"token"`
	TeamID       string `schema:"team_id"`
	TeamDomain   string `schema:"team_domain"`
	EnterpriseID string `schema:"enterprise_id"`
	ChannelID    string `schema:"channel_id"`
	ChannelName  string `schema:"channel_name"`
	UserID       string `schema:"user_id"`
	UserName     string `schema:"user_name"`
	Command      string `schema:"command"`
	Text         string `schema:"text"`
	ResponseURL  string `schema:"response_url"`
	TriggerID    string `schema:"trigger_id"`
}

type slackPayloadResponse struct {
	Payload string `schema:"payload"`
}

type slackPayload struct {
	Actions         []slackPayloadAction `json:"actions"`
	CallbackID      string               `json:"string"`
	Team            slackTeam            `json:"team"`
	Channel         slackChannel         `json:"channel"`
	User            slackUser            `json:"user"`
	ActionTS        string               `json:"action_ts"`
	MessageTS       string               `json:"message_ts"`
	AttachmentID    string               `json:"attachment_id"`
	Token           string               `json:"token"`
	OriginalMessage slackOriginalMessage `json:"-"`
	ResponseURL     string               `json:"response_url"`
	IsAppUnfurl     string               `json:"is_app_unfurl"`
	TriggerID       string               `json:"trigger_id"`
}

type payloadValue struct {
	Move      string `json:"move"`
	SessionID string `json:"session_id"`
}

type slackPayloadAction struct {
	Name  string `schema:"name"`
	Value string `schema:"value"`
	Type  string `schema:"type"`
}

type slackTeam struct {
	ID     string `schema:"id"`
	Domain string `schema:"domain"`
}

type slackChannel struct {
	ID   string `schema:"id"`
	Name string `schema:"name"`
}

type slackUser struct {
	ID   string `schema:"id"`
	Name string `schema:"name"`
}

type slackOriginalMessage struct{}

type slackResponse struct {
	ResponseType string            `json:"response_type"`
	Text         string            `json:"text"`
	Attachments  []slackAttachment `json:"attachments,omitempty"`
}

type slackAttachment struct {
	Text           string                  `json:"text"`
	Fallback       string                  `json:"fallback"`
	CallbackID     string                  `json:"callback_id"`
	Color          string                  `json:"color"`
	AttachmentType string                  `json:"attachment_type"`
	Actions        []slackAttachmentAction `json:"actions"`
}

type slackAttachmentAction struct {
	Name    string              `json:"name"`
	Text    string              `json:"text"`
	Style   string              `json:"style,omitempty"`
	Type    string              `json:"type"`
	Value   string              `json:"value"`
	Confirm *slackConfirmButton `json:"confirm,omitempty"`
}

type slackConfirmButton struct {
	Title       string `json:"title"`
	Text        string `json:"text"`
	OkText      string `json:"ok_text"`
	DismissText string `json:"dismiss_text"`
}

type gameSessionsManager struct {
	GameSessions map[string]gameSession
}

type gameSession struct {
	timestamp             time.Time
	status                string
	challenger            string
	target                string
	challengerMove        string
	targetMove            string
	challengerResponseUrl string
	targetResponseUrl     string
}

func (gameSessionsManager *gameSessionsManager) cleanSessions() {
	gameSessions := gameSessionsManager.GameSessions
	for k, v := range gameSessions {
		if time.Since(v.timestamp) > (time.Duration(30) * time.Minute) {
			delete(gameSessions, k)
		}
	}
}

func (gameSessionsManager *gameSessionsManager) createSession(challenger string, target string, responseUrl string) string {
	gameSessions := gameSessionsManager.GameSessions
	u := uuid.NewV4().String()

	gameSessions[u] = gameSession{
		timestamp:             time.Now(),
		status:                gameSessionInitiatedStatus,
		challenger:            challenger,
		target:                target,
		challengerResponseUrl: responseUrl,
	}

	return u
}

func newGameServer(game game.Game) *gameServer {
	gs := gameServer{
		ServeMux:            http.NewServeMux(),
		GameSessionsManager: newGameSessionsManager(),
		Game:                game,
	}
	registerRoutes(&gs)

	return &gs
}

func newGameSessionsManager() *gameSessionsManager {
	return &gameSessionsManager{
		GameSessions: make(map[string]gameSession),
	}
}

func init() {
	rpsGame := os.Getenv("RPS_GAME")

	if registeredGame, ok := modes.RegisteredGames[rpsGame]; ok {
		GameServer = newGameServer(registeredGame)
	} else {
		GameServer = newGameServer(modes.StandardGame)
	}

}
