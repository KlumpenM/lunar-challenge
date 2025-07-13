package storage

import (
	"sync"

	"lunar-backend-challenge/internal/models"
)

// RocketRepository provides storage for rockets with out-of-order message handling
type RocketRepository struct {
	rockets           map[string]*models.RocketState
	processedMessages map[string]map[int]bool                  // Track processed messages for deduplication
	pendingMessages   map[string]map[int]*models.RocketMessage // Buffer for out-of-order messages
	mutex             sync.RWMutex                             // Thread-safe access
}

// NewRocketRepository creates a new rocket repository
func NewRocketRepository() *RocketRepository {
	return &RocketRepository{
		rockets:           make(map[string]*models.RocketState),
		processedMessages: make(map[string]map[int]bool),
		pendingMessages:   make(map[string]map[int]*models.RocketMessage),
	}
}

// GetRocket retrieves a rocket by its ID
func (r *RocketRepository) GetRocket(id string) (*models.RocketState, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	rocket, exists := r.rockets[id]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid data races
	rocketCopy := *rocket
	return &rocketCopy, true
}

// GetAllRockets returns all rockets as summaries
func (r *RocketRepository) GetAllRockets() []models.RocketSummary {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	summaries := make([]models.RocketSummary, 0, len(r.rockets))

	for _, rocket := range r.rockets {
		summaries = append(summaries, models.RocketSummary{
			ID:        rocket.ID,
			Type:      rocket.Type,
			Speed:     rocket.Speed,
			Mission:   rocket.Mission,
			Exploded:  rocket.Exploded,
			UpdatedAt: rocket.UpdatedAt,
		})
	}

	return summaries
}

// ProcessMessage processes a rocket message with deduplication and out-of-order handling
func (r *RocketRepository) ProcessMessage(msg *models.RocketMessage) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	rocketID := msg.GetChannel()
	msgNumber := msg.GetMessageNumber()

	// Initialize maps for this rocket if they don't exist
	if r.processedMessages[rocketID] == nil {
		r.processedMessages[rocketID] = make(map[int]bool)
	}
	if r.pendingMessages[rocketID] == nil {
		r.pendingMessages[rocketID] = make(map[int]*models.RocketMessage)
	}

	// Check for duplicate message (at-least-once guarantee)
	if r.processedMessages[rocketID][msgNumber] {
		return true // Already processed, ignore duplicate
	}

	// Get or create rocket
	rocket, exists := r.rockets[rocketID]
	if !exists {
		// Only create new rocket if it's a launch message
		if msg.GetMessageType() != models.MessageTypeRocketLaunched {
			// Buffer non-launch messages for rockets that don't exist yet
			r.pendingMessages[rocketID][msgNumber] = msg
			return true
		}
		rocket = &models.RocketState{
			ID:                         rocketID,
			LastProcessedMessageNumber: 0,
		}
		r.rockets[rocketID] = rocket
	}

	// Check if this is the next expected message in sequence
	expectedMsgNumber := rocket.LastProcessedMessageNumber + 1

	if msgNumber == expectedMsgNumber {
		// Process this message immediately
		if r.processMessageByType(rocket, msg) {
			r.processedMessages[rocketID][msgNumber] = true
			rocket.LastProcessedMessageNumber = msgNumber
			rocket.UpdatedAt = msg.GetMessageTime()

			// Try to process any pending messages that are now in sequence
			r.processPendingMessages(rocketID)
			return true
		}
		return false
	} else if msgNumber > expectedMsgNumber {
		// Message is out of order - buffer it for later processing
		r.pendingMessages[rocketID][msgNumber] = msg
		return true
	}

	// Message is older than expected (already processed or very old)
	return false
}

// processPendingMessages processes any buffered messages that are now in sequence
func (r *RocketRepository) processPendingMessages(rocketID string) {
	rocket := r.rockets[rocketID]
	pendingForRocket := r.pendingMessages[rocketID]

	// Keep processing messages in sequence until we hit a gap
	for {
		nextMsgNumber := rocket.LastProcessedMessageNumber + 1
		msg, exists := pendingForRocket[nextMsgNumber]

		if !exists {
			break // No more sequential messages available
		}

		// Check if rocket exploded and only allow relaunch messages
		if rocket.Exploded && msg.GetMessageType() != models.MessageTypeRocketLaunched {
			// Remove the message from pending and continue
			delete(pendingForRocket, nextMsgNumber)
			continue
		}

		// Process the message
		if r.processMessageByType(rocket, msg) {
			r.processedMessages[rocketID][nextMsgNumber] = true
			rocket.LastProcessedMessageNumber = nextMsgNumber
			rocket.UpdatedAt = msg.GetMessageTime()

			// Remove processed message from pending
			delete(pendingForRocket, nextMsgNumber)
		} else {
			// Failed to process - remove from pending and stop
			delete(pendingForRocket, nextMsgNumber)
			break
		}
	}
}

// processMessageByType handles different message types
func (r *RocketRepository) processMessageByType(rocket *models.RocketState, msg *models.RocketMessage) bool {
	// If rocket exploded, only allow relaunch messages
	if rocket.Exploded && msg.GetMessageType() != models.MessageTypeRocketLaunched {
		return false
	}

	switch msg.GetMessageType() {

	case models.MessageTypeRocketLaunched:
		// Validate required fields
		if msg.Message.Type == "" || msg.Message.Mission == "" {
			return false
		}

		// Reset rocket state for new launch (can relaunch exploded rockets)
		rocket.Type = msg.Message.Type
		rocket.Mission = msg.Message.Mission
		rocket.Speed = msg.Message.LaunchSpeed
		rocket.Exploded = false
		rocket.Reason = ""

		// Set created time only for first launch
		if rocket.CreatedAt.IsZero() {
			rocket.CreatedAt = msg.GetMessageTime()
		}
		return true

	case models.MessageTypeRocketSpeedIncreased:
		if msg.Message.By <= 0 {
			return false
		}
		rocket.Speed += msg.Message.By
		return true

	case models.MessageTypeRocketSpeedDecreased:
		if msg.Message.By <= 0 {
			return false
		}
		rocket.Speed -= msg.Message.By
		if rocket.Speed < 0 {
			rocket.Speed = 0
		}
		return true

	case models.MessageTypeRocketExploded:
		if msg.Message.Reason == "" {
			return false
		}
		rocket.Exploded = true
		rocket.Reason = msg.Message.Reason

		// Clear pending messages for exploded rocket (except launch messages)
		rocketID := rocket.ID
		if pendingForRocket, exists := r.pendingMessages[rocketID]; exists {
			for msgNum, pendingMsg := range pendingForRocket {
				if pendingMsg.GetMessageType() != models.MessageTypeRocketLaunched {
					delete(pendingForRocket, msgNum)
				}
			}
		}
		return true

	case models.MessageTypeRocketMissionChanged:
		if msg.Message.NewMission == "" {
			return false
		}
		rocket.Mission = msg.Message.NewMission
		return true

	default:
		return false
	}
}

// GetDebugInfo returns debug information for a rocket
func (r *RocketRepository) GetDebugInfo(rocketID string) (processedCount int, pendingMessages []int) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if processed := r.processedMessages[rocketID]; processed != nil {
		processedCount = len(processed)
	}

	if pending := r.pendingMessages[rocketID]; pending != nil {
		for msgNum := range pending {
			pendingMessages = append(pendingMessages, msgNum)
		}
	}

	return processedCount, pendingMessages
}
