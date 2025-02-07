package client

type GenericChannelActionWithoutPayload struct {
	ID          string `json:"id"`
	ChannelID   string `json:"channel_id"`
	Enabled     bool   `json:"enabled"`
	DeleteAt    int64  `json:"delete_at"`
	ActionType  string `json:"action_type"`
	TriggerType string `json:"trigger_type"`
}

type GenericChannelAction struct {
	GenericChannelActionWithoutPayload
	Payload interface{} `json:"payload"`
}

type WelcomeMessagePayload struct {
	Message string `json:"message" mapstructure:"message"`
}

type WelcomeMessageAction struct {
	GenericChannelActionWithoutPayload
	Payload WelcomeMessagePayload `json:"payload"`
}

const (
	// Action types
	ActionTypeWelcomeMessage = "send_welcome_message"

	// Trigger types
	TriggerTypeNewMemberJoins = "new_member_joins"
)

// ChannelActionListOptions specifies the optional parameters to the
// ActionsService.List method.
type ChannelActionListOptions struct {
	TriggerType string `url:"trigger_type,omitempty"`
}

// ChannelActionCreateOptions specifies the parameters for ActionsService.Create method.
type ChannelActionCreateOptions struct {
	ChannelID   string      `json:"channel_id"`
	Enabled     bool        `json:"enabled"`
	ActionType  string      `json:"action_type"`
	TriggerType string      `json:"trigger_type"`
	Payload     interface{} `json:"payload"`
}
