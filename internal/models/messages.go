package models

import "time"

// RocketMessage represents a message about a rocket's state change
// @Description A message containing information about a rocket's state change
type RocketMessage struct {
	Metadata struct {
		Channel       string    `json:"channel" example:"193270a9-c9cf-404a-8f83-838e71d9ae67"`
		MessageNumber int       `json:"messageNumber" example:"1"`
		MessageTime   time.Time `json:"messageTime" example:"2024-03-14T19:39:05.86337+01:00"`
		MessageType   string    `json:"messageType" example:"RocketLaunched"`
	} `json:"metadata"`

	// Message content with all possible fields
	Message MessageContent `json:"message"`
}

// MessageContent contains the payload of a rocket message
// @Description The content of a rocket message, with fields varying based on message type
type MessageContent struct {
	// RocketLaunched fields
	Type        string `json:"type,omitempty" example:"Falcon-9"`
	LaunchSpeed int    `json:"launchSpeed,omitempty" example:"500"`
	Mission     string `json:"mission,omitempty" example:"ARTEMIS"`

	// RocketSpeedIncreased/Decreased fields
	By int `json:"by,omitempty" example:"3000"`

	// RocketExploded fields
	Reason string `json:"reason,omitempty" example:"PRESSURE_VESSEL_FAILURE"`

	// RocketMissionChanged fields
	NewMission string `json:"newMission,omitempty" example:"SHUTTLE_MIR"`
}

// Message type constants
const (
	MessageTypeRocketLaunched       = "RocketLaunched"
	MessageTypeRocketSpeedIncreased = "RocketSpeedIncreased"
	MessageTypeRocketSpeedDecreased = "RocketSpeedDecreased"
	MessageTypeRocketExploded       = "RocketExploded"
	MessageTypeRocketMissionChanged = "RocketMissionChanged"
)

// Rocket status constants
const (
	RocketStatusActive   = "active"
	RocketStatusExploded = "exploded"
)

func (m *RocketMessage) GetChannel() string {
	return m.Metadata.Channel
}

func (m *RocketMessage) GetMessageNumber() int {
	return m.Metadata.MessageNumber
}

func (m *RocketMessage) GetMessageTime() time.Time {
	return m.Metadata.MessageTime
}

func (m *RocketMessage) GetMessageType() string {
	return m.Metadata.MessageType
}
