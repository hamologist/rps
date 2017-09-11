package main

import (
	"log"
	"net/http"
	"os"

	"bitbucket.org/hamologist/rps-slack/slack"
)

const (
	DefaultPort = ":8081"
)

var (
	applicationPort = ":" + os.Getenv("RPS_PORT")
)

func main() {
	log.Fatal(http.ListenAndServe(applicationPort, slack.GameServer.ServeMux))
}

func init() {
	if applicationPort == ":" {
		applicationPort = DefaultPort
	}
}
