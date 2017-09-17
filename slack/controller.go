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

	"bitbucket.org/hamologist/rps/game"
	"bitbucket.org/hamologist/rps/server"
	"github.com/gorilla/schema"
)

var (
	commandName string
	debug       bool
	decoder     = schema.NewDecoder()
	encoder     = schema.NewEncoder()
)

type payloadValue struct {
	Move      string `json:"move"`
	SessionID string `json:"session_id"`
}

type controller struct {
	*server.GameServer
}

func (controller *controller) HandleGameRequest(w http.ResponseWriter, r *http.Request) {
	var (
		target string
		body   Body
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

		err = decoder.Decode(&body, r.PostForm)
		if err != nil {
			log.Print(err)
		}
		textTokens := strings.Split(body.Text, " ")

		if len(textTokens) >= 1 {
			target = textTokens[0]
		}

		controller.processChallengeAction(body.UserName, target, body.ResponseURL, w)
	}
}

func (controller *controller) HandleGameAccept(w http.ResponseWriter, r *http.Request) {
	var (
		sessionID string
		body      Body
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

		err = decoder.Decode(&body, r.PostForm)
		if err != nil {
			log.Print(err)
		}
		textTokens := strings.Split(body.Text, " ")

		if len(textTokens) >= 1 {
			sessionID = textTokens[0]
		}

		controller.processAcceptAction(sessionID, body.ResponseURL, w)
	}

}

func (controller *controller) HandleGamePayload(w http.ResponseWriter, r *http.Request) {
	var (
		payloadResponse PayloadResponse
		payload         Payload
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

		decoder.Decode(&payloadResponse, r.PostForm)
		if err != nil {
			log.Print(err)
		}
		json.Unmarshal([]byte(payloadResponse.Payload), &payload)

		controller.processPayload(payload, w)
	}

}

func (controller *controller) processChallengeAction(challenger string, target string, responseURL string, w http.ResponseWriter) {
	target = strings.Replace(target, "@", "", 1)
	slackData := createSlackData(gameSessionInitiatedStatus, responseURL, "")
	uuid := controller.GameSessionsManager.CreateSession(challenger, target, slackData)

	resp := Response{
		ResponseType: inChannelResponse,
		Text: fmt.Sprintf(
			"@%v, @%v has challenged you to a game of rock, paper, scissors. "+
				"Please use `/%v %v` to accept.",
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

func (controller *controller) processAcceptAction(gameSessionUUID string, responseURL string, w http.ResponseWriter) {
	var (
		slackAttachmentActions   []AttachmentAction
		slackAttachmentActionMap = make(map[string]AttachmentAction)
		challengerResponseURL    string
		status                   string
	)
	gameMoves := controller.Game.Moves
	preferredMoveOrder := controller.Game.PreferredOrder
	gameSessions := controller.GameSessionsManager.GameSessions

	if v, ok := gameSessions[gameSessionUUID]; ok {

		if !validSlackData(v.Data) {
			fmt.Fprint(w, "The provided game session does not support slack")
			return
		}

		status = v.Data["status"]
		challengerResponseURL = v.Data["challengerResponseUrl"]

		if status != gameSessionInitiatedStatus {
			fmt.Fprint(w, "The provided game session id has already been accepted")
			return
		}

		v.Data["targetResponseUrl"] = responseURL

		for _, v := range gameMoves {
			jsonData, err := json.Marshal(map[string]string{
				"session_id": gameSessionUUID,
				"move":       strings.ToLower(v.Name),
			})

			if err == nil {
				slackAttachmentActionMap[v.Name] = AttachmentAction{
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

		resp := Response{
			ResponseType: ephemeralResponse,
			Text:         "RPS initiated",
			Attachments: []Attachment{
				Attachment{
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
			http.Post(challengerResponseURL, "application/json", bytes.NewBuffer(js))
			v.Data["status"] = gameSessionAcceptedStatus
		}
	} else {
		fmt.Fprint(w, "The provided game session id is not valid. Game session either doesn't exist or has expired.")
	}
}

func (controller *controller) processPayload(payload Payload, w http.ResponseWriter) {
	if len(payload.Actions) == 0 {
		fmt.Fprint(w, "No action was submitted")
		return
	}

	var (
		payloadValue payloadValue
		playResult   string
	)
	gameSessions := controller.GameSessionsManager.GameSessions
	user := payload.User.Name
	json.Unmarshal([]byte(payload.Actions[0].Value), &payloadValue)

	if v, ok := gameSessions[payloadValue.SessionID]; ok {
		if user == v.Challenger {
			v.ChallengerMove = payloadValue.Move
			gameSessions[payloadValue.SessionID] = v
			fmt.Fprint(w, "Your move has been locked in")
		} else if user == v.Target {
			v.TargetMove = payloadValue.Move
			gameSessions[payloadValue.SessionID] = v
			fmt.Fprint(w, "Your move has been locked in")
		} else {
			fmt.Fprint(w, "A user not associated to the sessions attempted to submit a game move.")
			return
		}

		if len(v.ChallengerMove) != 0 && len(v.TargetMove) != 0 {
			result, err := controller.Game.Play(v.ChallengerMove, v.TargetMove)

			if err != nil {
				fmt.Fprint(w, err)
				return
			}

			if result == game.GameStatePlayerOneWins {
				playResult = fmt.Sprintf(
					"@%v defeated @%v, %v beats %v",
					v.Challenger,
					v.Target,
					v.ChallengerMove,
					v.TargetMove,
				)
			} else if result == game.GameStatePlayerTwoWins {
				playResult = fmt.Sprintf(
					"@%v defeated @%v, %v beats %v",
					v.Target,
					v.Challenger,
					v.TargetMove,
					v.ChallengerMove,
				)
			} else {
				playResult = fmt.Sprintf(
					"@%v and @%v had a draw. Both played %v",
					v.Challenger,
					v.Target,
					v.TargetMove,
				)
			}

			resp := Response{
				ResponseType: inChannelResponse,
				Text:         playResult,
			}
			js, err := json.Marshal(resp)

			if err != nil {
				log.Print(err)
				fmt.Fprint(w, "An error occurred while proccessing the game")
			} else {
				http.Post(v.Data["challengerResponseUrl"], "application/json", bytes.NewBuffer(js))
			}
		}
	} else {
		fmt.Fprint(w, "An invalid session id was passed with your move. Maybe the game session has expired.")
		return
	}
}

func (controller *controller) logRequest(r *http.Request) {
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Print(err)
	}
	log.Print(string(requestDump))
}

func newController(gameServer *server.GameServer) *controller {
	return &controller{
		GameServer: DefaultGameServer,
	}
}

func init() {
	debug = os.Getenv("RPS_DEBUG") == "1"
	commandName = os.Getenv("RPS_COMMAND_NAME")

	if commandName == "" {
		commandName = defaultAcceptCommandName
	}
}
