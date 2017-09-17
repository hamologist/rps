package server

import (
	"net/http"
	"time"

	"github.com/satori/go.uuid"

	"bitbucket.org/hamologist/rps/game"
)

// GameServer defines a server/session/game coupling used for simulating games of RPS.
type GameServer struct {
	ServeMux            *http.ServeMux
	GameSessionsManager *SessionManager
	Game                game.Game
}

// CleanUp is intended to be run in a goroutine.
// Cleanup invokes the GameSessionsManager::CleanSessions method every minute.
func (gameServer *GameServer) CleanUp() {
	t := time.NewTicker(time.Minute)
	for {
		gameServer.GameSessionsManager.CleanSessions()
		<-t.C
	}
}

// SessionManager provides a means of managing sessions needed by the GameServer.
type SessionManager struct {
	GameSessions map[string]*GameSession
}

// GameSession defines the data used by a game session.
// GameSession's Data field is intended for storing data specific to a consumer
// (see the bitbucket.org/hamologist/rps/slack package for an example).
type GameSession struct {
	Timestamp      time.Time
	Challenger     string
	Target         string
	ChallengerMove string
	TargetMove     string
	Data           map[string]string
}

// CreateSession creates a session used by the SessionManger.
func (sessionManager *SessionManager) CreateSession(challenger, target string, data map[string]string) string {
	gameSessions := sessionManager.GameSessions
	u := uuid.NewV4().String()

	gameSessions[u] = &GameSession{
		Timestamp:  time.Now(),
		Challenger: challenger,
		Target:     target,
		Data:       data,
	}

	return u
}

// CleanSessions removes all sessions older than 30 minutes.
// CleanSessions is intended to be invoked by the GameServer's CleanUp method.
func (sessionManager *SessionManager) CleanSessions() {
	gameSessions := sessionManager.GameSessions
	for k, v := range gameSessions {
		if time.Since(v.Timestamp) > (time.Duration(30) * time.Minute) {
			delete(gameSessions, k)
		}
	}
}

// NewGameServer creates a GameServer.
func NewGameServer(game game.Game) *GameServer {
	return &GameServer{
		ServeMux:            http.NewServeMux(),
		GameSessionsManager: newSessionManager(),
		Game:                game,
	}
}

func newSessionManager() *SessionManager {
	return &SessionManager{
		GameSessions: make(map[string]*GameSession),
	}
}
