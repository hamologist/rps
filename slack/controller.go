package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"bitbucket.org/hamologist/rps-go/game"
	"github.com/gorilla/schema"
)

var (
	commandName string
	debug       bool
	decoder     = schema.NewDecoder()
	encoder     = schema.NewEncoder()
)

type controller struct {
	gameServer *gameServer
}

func (controller controller) HandleStandardGame(w http.ResponseWriter, r *http.Request) {
	var (
		action          string
		param           string
		body            slackBody
		payloadResponse slackPayloadResponse
		payload         slackPayload
	)

	if r.Method == "POST" {
		if debug {
			controller.logRequest(r)
		}

		err := r.ParseForm()
		if err != nil {
			log.Print(err)
			fmt.Fprint(w, "An error occurred while proccessing your response")
			return
		}

		if _, ok := r.PostForm["payload"]; ok {
			decoder.Decode(&payloadResponse, r.PostForm)
			if err != nil {
				log.Print(err)
			}
			json.Unmarshal([]byte(payloadResponse.Payload), &payload)

			controller.processPayload(payload, w)
		} else {
			err = decoder.Decode(&body, r.PostForm)
			if err != nil {
				log.Print(err)
			}
			textTokens := strings.Split(body.Text, " ")

			if len(textTokens) > 1 {
				action = textTokens[0]
				param = textTokens[1]
			}

			if action == challengeAction {
				controller.processChallengeAction(body.UserName, param, body.ResponseURL, w)
			} else if action == acceptAction {
				controller.processAcceptAction(param, body.ResponseURL, w)
			}
		}
	}
}

func (controller *controller) processChallengeAction(challenger string, target string, responseUrl string, w http.ResponseWriter) {
	target = strings.Replace(target, "@", "", 1)
	uuid := controller.gameServer.GameSessionsManager.createSession(challenger, target, responseUrl)

	resp := slackResponse{
		ResponseType: inChannelResponse,
		Text: fmt.Sprintf(
			"%v, @%v has challenged you to a game of rock, paper, scissors. "+
				"Please use `/%v accept %v` to accept.",
			target,
			challenger,
			commandName,
			uuid,
		),
	}
	js, err := json.Marshal(resp)

	if err != nil {
		log.Print(err)
		fmt.Fprint(w, "An error occurred while proccessing your challenge")
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}

}

func (controller *controller) processAcceptAction(gameSessionUUID string, responseUrl string, w http.ResponseWriter) {
	var (
		slackAttachmentActions   []slackAttachmentAction
		slackAttachmentActionMap = make(map[string]slackAttachmentAction)
		challengerResponseUrl    string
	)
	gameMoves := controller.gameServer.Game.Moves
	preferredMoveOrder := controller.gameServer.Game.PreferredOrder
	gameSessions := controller.gameServer.GameSessionsManager.GameSessions

	if v, ok := gameSessions[gameSessionUUID]; ok {

		if v.status != gameSessionInitiatedStatus {
			fmt.Fprint(w, "The provided game session id has already been accepted.")
			return
		}

		v.targetResponseUrl = responseUrl
		challengerResponseUrl = v.challengerResponseUrl

		for _, v := range gameMoves {
			jsonData, err := json.Marshal(map[string]string{
				"session_id": gameSessionUUID,
				"move":       strings.ToLower(v.Name),
			})

			if err == nil {
				slackAttachmentActionMap[v.Name] = slackAttachmentAction{
					Name:  "move",
					Text:  strings.Title(v.Name),
					Type:  "button",
					Value: string(jsonData),
				}
			} else {
				log.Print(err)
				fmt.Fprint(w, "An error occurred while building the moves for the game")
				return
			}
		}

		for _, move := range preferredMoveOrder {
			slackAttachmentActions = append(slackAttachmentActions, slackAttachmentActionMap[move])
		}

		resp := slackResponse{
			ResponseType: ephemeralResponse,
			Text:         "RPS initiated",
			Attachments: []slackAttachment{
				slackAttachment{
					Text:           "Please select your move",
					Fallback:       "You are unable to choose a move",
					CallbackID:     "player_move_selection",
					Color:          "#3AA3E3",
					AttachmentType: "default",
					Actions:        slackAttachmentActions,
				},
			},
		}
		js, err := json.Marshal(resp)

		if err != nil {
			log.Print(err)
			fmt.Fprint(w, "An error occurred while accepting the challenge")
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
			http.Post(challengerResponseUrl, "application/json", bytes.NewBuffer(js))
		}
	} else {
		fmt.Fprint(w, "The provided game session id is not valid. Game session either doesn't exist or has expired.")
	}
}

func (controller *controller) processPayload(payload slackPayload, w http.ResponseWriter) {
	if len(payload.Actions) == 0 {
		fmt.Fprint(w, "No action was submitted")
		return
	}

	var (
		payloadValue payloadValue
		playResult   string
	)
	gameSessions := controller.gameServer.GameSessionsManager.GameSessions
	user := payload.User.Name
	json.Unmarshal([]byte(payload.Actions[0].Value), &payloadValue)

	if v, ok := gameSessions[payloadValue.SessionID]; ok {
		if user == v.challenger {
			v.challengerMove = payloadValue.Move
			gameSessions[payloadValue.SessionID] = v
			fmt.Fprint(w, "Your move has been locked in")
		} else if user == v.target {
			v.targetMove = payloadValue.Move
			gameSessions[payloadValue.SessionID] = v
			fmt.Fprint(w, "Your move has been locked in")
		} else {
			fmt.Fprint(w, "A user not associated to the sessions attempted to submit a game move.")
			return
		}

		if len(v.challengerMove) != 0 && len(v.targetMove) != 0 {
			result, err := controller.gameServer.Game.Play(v.challengerMove, v.targetMove)

			if err != nil {
				fmt.Fprint(w, err)
				return
			}

			if result == game.GameStatePlayerOneWins {
				playResult = fmt.Sprintf(
					"@%v defeated @%v, %v beats %v",
					v.challenger,
					v.target,
					v.challengerMove,
					v.targetMove,
				)
			} else if result == game.GameStatePlayerTwoWins {
				playResult = fmt.Sprintf(
					"@%v defeated @%v, %v beats %v",
					v.challenger,
					v.target,
					v.targetMove,
					v.challengerMove,
				)
			} else {
				playResult = fmt.Sprintf(
					"@%v and %v had a draw. Both played %v",
					v.challenger,
					v.target,
					v.targetMove,
				)
			}

			resp := slackResponse{
				ResponseType: inChannelResponse,
				Text:         playResult,
			}
			js, err := json.Marshal(resp)

			if err != nil {
				log.Print(err)
				fmt.Fprint(w, "An error occurred while proccessing the game")
			} else {
				http.Post(v.challengerResponseUrl, "application/json", bytes.NewBuffer(js))
			}
		}
	} else {
		fmt.Fprint(w, "An invalid session id was passed with your move. Maybe the game session has expired.")
		return
	}
}

func (controller controller) logRequest(r *http.Request) {
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Print(err)
	}
	log.Print(string(requestDump))
}

func newController(gameServer *gameServer) *controller {
	return &controller{
		gameServer: gameServer,
	}
}

func init() {
	debug = os.Getenv("RPS_DEBUG") == "1"
	commandName = os.Getenv("RPS_COMMAND_NAME")

	if commandName == "" {
		commandName = defaultCommandName
	}
}
