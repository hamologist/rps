// Package slack provides structs to ease interactions with the slack API.
package slack

// Body provides a way to structure and work with the standard slack /command response.
type Body struct {
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

// PayloadResponse is used for loading a Payload from slack (usually from a Interactive Message interaction).
// Actual data in this Payload is a JSON string which is loaded using SlackPayload
type PayloadResponse struct {
	Payload string `schema:"payload"`
}

// Payload is used for unmarshalling data from a SlackPayloadResponse.
type Payload struct {
	Actions         []PayloadAction `json:"actions"`
	CallbackID      string          `json:"string"`
	Team            Team            `json:"team"`
	Channel         Channel         `json:"channel"`
	User            User            `json:"user"`
	ActionTS        string          `json:"action_ts"`
	MessageTS       string          `json:"message_ts"`
	AttachmentID    string          `json:"attachment_id"`
	Token           string          `json:"token"`
	OriginalMessage OriginalMessage `json:"-"`
	ResponseURL     string          `json:"response_url"`
	IsAppUnfurl     string          `json:"is_app_unfurl"`
	TriggerID       string          `json:"trigger_id"`
}

// PayloadAction defines data about the action that was performed.
type PayloadAction struct {
	Name  string `schema:"name"`
	Value string `schema:"value"`
	Type  string `schema:"type"`
}

// Team defines data about the slack team the action was executed with.
type Team struct {
	ID     string `schema:"id"`
	Domain string `schema:"domain"`
}

// Channel defines data about the channel the action was executed in.
type Channel struct {
	ID   string `schema:"id"`
	Name string `schema:"name"`
}

// User defines data about the user that executed the action.
type User struct {
	ID   string `schema:"id"`
	Name string `schema:"name"`
}

// OriginalMessage is currently omitted.
type OriginalMessage struct{}

// Response is used to send either a in_channel or ephemeral response to a given user.
type Response struct {
	ResponseType string       `json:"response_type"`
	Text         string       `json:"text"`
	Attachments  []Attachment `json:"attachments,omitempty"`
}

// Attachment allows a Response to define message areas and setup Actions on a Response.
type Attachment struct {
	Text           string             `json:"text"`
	Fallback       string             `json:"fallback"`
	CallbackID     string             `json:"callback_id"`
	Color          string             `json:"color"`
	AttachmentType string             `json:"attachment_type"`
	Actions        []AttachmentAction `json:"actions"`
}

// AttachmentAction allows an Attachment to define interactive message elements.
type AttachmentAction struct {
	Name    string         `json:"name"`
	Text    string         `json:"text"`
	Style   string         `json:"style,omitempty"`
	Type    string         `json:"type"`
	Value   string         `json:"value"`
	Confirm *ConfirmButton `json:"confirm,omitempty"`
}

// ConfirmButton allows a user to attach an optional confirm button to an AttachmentAction.
type ConfirmButton struct {
	Title       string `json:"title"`
	Text        string `json:"text"`
	OkText      string `json:"ok_text"`
	DismissText string `json:"dismiss_text"`
}
