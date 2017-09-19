package slack

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/nlopes/slack"

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

		controller.processChallengeAction(body.UserID, body.UserName, target, body.ChannelID, w)
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

func (controller *controller) processChallengeAction(challenger, challengerName, target, channel string, w http.ResponseWriter) {
	var (
		slackAttachmentActions   []AttachmentAction
		slackAttachmentActionMap = make(map[string]AttachmentAction)
	)
	target = strings.Split(strings.Replace(target, "<@", "", 1), "|")[0]
	targetInfo, err := API.GetUserInfo(target)

	if err != nil {
		fmt.Fprint(w,
			"The user you are attempting to challenge isn't a valid Slack user for this team.\n"+
				"Make sure you are using the \"@\" mention syntax.",
		)
		return
	}
	targetName := targetInfo.Name
	form := url.Values{}

	slackData := createSlackData(channel, challengerName, targetName)
	uuid := controller.GameSessionsManager.CreateSession(challenger, target, slackData)
	gameMoves := controller.Game.Moves
	preferredMoveOrder := controller.Game.PreferredOrder

	for _, v := range gameMoves {
		jsonData, err := json.Marshal(map[string]string{
			"session_id": uuid,
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

	respAttachment := []Attachment{
		Attachment{
			Text:           "Please select your move",
			Fallback:       "You are unable to choose a move",
			CallbackID:     "player_move_selection",
			Color:          "#3AA3E3",
			AttachmentType: "default",
			Actions:        slackAttachmentActions,
		},
	}
	js, err := json.Marshal(respAttachment)

	if err != nil {
		log.Print(err)
		fmt.Fprint(w, "An error occurred while setting up the game.")
		return
	}

	resp := PostEphemeralPayload{
		Token:       OAuthToken,
		Channel:     channel,
		Text:        fmt.Sprintf("@%v has challenged you to a game of RPS.", challengerName),
		User:        target,
		AsUser:      false,
		Attachments: string(js),
	}

	err = encoder.Encode(resp, form)
	if err != nil {
		fmt.Fprint(w, "There was a problem building the challenge for your opponent. Please try again.")
		return
	}

	_, err = http.PostForm(PostEphemeralRoute, form)
	if err != nil {
		fmt.Fprint(w, "There was a problem issuing the challenge to your opponent. Please try again.")
		return
	}

	resp = PostEphemeralPayload{
		Token:       OAuthToken,
		Channel:     channel,
		Text:        fmt.Sprintf("The challenge was submitted to @%v. They are now selecting a move.", targetName),
		User:        challenger,
		AsUser:      false,
		Attachments: string(js),
	}

	err = encoder.Encode(resp, form)
	if err != nil {
		fmt.Fprint(w, "There was a problem building the challenge for you. Please try again.")
		return
	}

	_, err = http.PostForm(PostEphemeralRoute, form)
	if err != nil {
		fmt.Fprint(w, "There was a problem returning the game back to you. Please try again.")
		return
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
	user := payload.User.ID

	err := json.Unmarshal([]byte(payload.Actions[0].Value), &payloadValue)
	if err != nil {
		fmt.Fprint(w, "There was a problem processing the game move payload.")
	}

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

			if !validSlackData(v.Data) {
				fmt.Fprint(w, "Game session does not support Slack.")
				return
			}
			channelName := v.Data["channelName"]
			challengerName := v.Data["challengerName"]
			targetName := v.Data["targetName"]

			if result == game.GameStatePlayerOneWins {
				playResult = fmt.Sprintf(
					"@%v defeated @%v, %v beats %v",
					challengerName,
					targetName,
					v.ChallengerMove,
					v.TargetMove,
				)
			} else if result == game.GameStatePlayerTwoWins {
				playResult = fmt.Sprintf(
					"@%v defeated @%v, %v beats %v",
					targetName,
					challengerName,
					v.TargetMove,
					v.ChallengerMove,
				)
			} else {
				playResult = fmt.Sprintf(
					"@%v and @%v had a draw. Both played %v",
					challengerName,
					targetName,
					v.TargetMove,
				)
			}

			_, _, err = API.PostMessage(channelName, playResult, slack.PostMessageParameters{})
			if err != nil {
				log.Print(err)
				fmt.Fprint(w, "Failed to post the game results to the channel.")
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
