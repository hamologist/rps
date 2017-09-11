package slack

const (
	HandleStandardGameRoute = "/slack/rps"
)

func registerRoutes(gameServer *gameServer) {
	controller := newController(gameServer)
	serveMux := controller.gameServer.ServeMux

	serveMux.HandleFunc(HandleStandardGameRoute, controller.HandleStandardGame)
}
