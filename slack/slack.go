package main

import (
	"fmt"
	"github.com/gorilla/schema"
	"log"
	"net/http"
	"net/http/httputil"
)

var decoder = schema.NewDecoder()

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
}

func main() {
	http.HandleFunc("/slack/rps", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			requestDump, err := httputil.DumpRequest(r, true)
			if err != nil {
				log.Print(err)
			}
			log.Print(string(requestDump))

			err = r.ParseForm()
			if err != nil {
				log.Print(err)
			}

			var body slackBody

			err = decoder.Decode(&body, r.PostForm)
			if err != nil {
				log.Print(err)
			}
			log.Print(body)
		} else {
			log.Print(r.Method)
		}
		fmt.Fprint(w, "Hello World")
	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}
