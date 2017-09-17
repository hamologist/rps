package slack

import "bitbucket.org/hamologist/rps/server"

const (
	// HandleGameRequestRoute defines the routes used by the HandleGameRequest controller method
	HandleGameRequestRoute = "/slack/rps"
	// HandleGameAcceptRoute defines the routes used by the HandleGameAccept controller method
	HandleGameAcceptRoute = "/slack/rps/accept"
	// HandleGamePayloadRoute defines the routes used by the HandleGamePayload controller method
	HandleGamePayloadRoute = "/slack/rps/payload"
)

func registerRoutes(gameServer *server.GameServer) {
	controller := newController(gameServer)
	serveMux := controller.ServeMux

	serveMux.HandleFunc(HandleGameRequestRoute, controller.HandleGameRequest)
	serveMux.HandleFunc(HandleGameAcceptRoute, controller.HandleGameAccept)
	serveMux.HandleFunc(HandleGamePayloadRoute, controller.HandleGamePayload)
}
