package validation

import (
	"lunar-backend-challenge/internal/errors"
	"lunar-backend-challenge/internal/models"
)

// ValidateRocketMessage validates a rocket message for required fields
func ValidateRocketMessage(msg *models.RocketMessage) error {
	// Validate metadata
	if msg.Metadata.Channel == "" {
		return errors.NewValidationError("channel", "channel is required")
	}

	if msg.Metadata.MessageNumber <= 0 {
		return errors.NewValidationError("messageNumber", "messageNumber must be positive integer")
	}

	if msg.Metadata.MessageTime.IsZero() {
		return errors.NewValidationError("messageTime", "messageTime is required")
	}

	if !isValidMessageType(msg.Metadata.MessageType) {
		return errors.NewValidationError("messageType", "invalid message type", msg.Metadata.MessageType)
	}

	// Validate message content based on type
	switch msg.Metadata.MessageType {
	case models.MessageTypeRocketLaunched:
		if msg.Message.Type == "" {
			return errors.NewValidationError("type", "rocket type is required for launch message")
		}
		if msg.Message.Mission == "" {
			return errors.NewValidationError("mission", "mission is required for launch message")
		}
		if msg.Message.LaunchSpeed < 0 {
			return errors.NewValidationError("launchSpeed", "launch speed cannot be negative")
		}

	case models.MessageTypeRocketSpeedIncreased, models.MessageTypeRocketSpeedDecreased:
		if msg.Message.By <= 0 {
			return errors.NewValidationError("by", "speed change amount must be positive")
		}

	case models.MessageTypeRocketExploded:
		if msg.Message.Reason == "" {
			return errors.NewValidationError("reason", "explosion reason is required")
		}

	case models.MessageTypeRocketMissionChanged:
		if msg.Message.NewMission == "" {
			return errors.NewValidationError("newMission", "new mission is required")
		}
	}

	return nil
}

// isValidMessageType checks if the message type is supported
func isValidMessageType(messageType string) bool {
	validTypes := []string{
		models.MessageTypeRocketLaunched,
		models.MessageTypeRocketSpeedIncreased,
		models.MessageTypeRocketSpeedDecreased,
		models.MessageTypeRocketExploded,
		models.MessageTypeRocketMissionChanged,
	}

	for _, validType := range validTypes {
		if messageType == validType {
			return true
		}
	}
	return false
}

// ValidateRocketID validates a rocket ID from URL path
func ValidateRocketID(rocketID string) error {
	if rocketID == "" {
		return errors.NewValidationError("rocketId", "rocket ID is required")
	}

	// Validate rocket ID format, here 3 is a magic number(e.g., UUID-like strings should be longer)
	if len(rocketID) < 3 {
		return errors.NewValidationError("rocketId", "rocket ID is too short", rocketID)
	}

	return nil
}
