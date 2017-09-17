package main

import (
	"log"
	"net/http"
	"os"

	"bitbucket.org/hamologist/rps/slack"
)

// DefaultPort defines the default application port that will be used on startup
const DefaultPort = ":8081"

var applicationPort = ":" + os.Getenv("RPS_PORT")

func main() {
	go slack.DefaultGameServer.CleanUp()
	log.Fatal(http.ListenAndServe(applicationPort, slack.DefaultGameServer.ServeMux))
}

func init() {
	if applicationPort == ":" {
		applicationPort = DefaultPort
	}
}
