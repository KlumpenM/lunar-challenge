package test

import (
	"testing"
	"time"

	"lunar-backend-challenge/internal/models"
	"lunar-backend-challenge/internal/storage"
)

// Helper function to create a test rocket message
func createTestMessage(channel string, messageNumber int, messageType string) *models.RocketMessage {
	msg := &models.RocketMessage{}
	msg.Metadata.Channel = channel
	msg.Metadata.MessageNumber = messageNumber
	msg.Metadata.MessageType = messageType
	msg.Metadata.MessageTime = time.Now()

	// Set message content based on type
	switch messageType {
	case models.MessageTypeRocketLaunched:
		msg.Message.Type = "Falcon Heavy"
		msg.Message.Mission = "Test Mission"
		msg.Message.LaunchSpeed = 1000
	case models.MessageTypeRocketSpeedIncreased:
		msg.Message.By = 500
	case models.MessageTypeRocketSpeedDecreased:
		msg.Message.By = 300
	case models.MessageTypeRocketExploded:
		msg.Message.Reason = "Engine failure"
	case models.MessageTypeRocketMissionChanged:
		msg.Message.NewMission = "New Mission"
	}

	return msg
}

// Test repository creation and initialization
func TestNewRocketRepository(t *testing.T) {
	repo := storage.NewRocketRepository()

	if repo == nil {
		t.Fatal("Expected repository to be created, got nil")
	}

	// Verify initialization by checking if GetAllRockets works
	rockets := repo.GetAllRockets()
	if rockets == nil {
		t.Error("Expected rockets slice to be initialized")
	}

	if len(rockets) != 0 {
		t.Errorf("Expected empty rockets list, got %d rockets", len(rockets))
	}
}

// Test rocket launch message processing
func TestProcessRocketLaunchMessage(t *testing.T) {
	repo := storage.NewRocketRepository()
	rocketID := "test-rocket-1"

	// Create launch message
	msg := createTestMessage(rocketID, 1, models.MessageTypeRocketLaunched)

	// Process the message
	success := repo.ProcessMessage(msg)
	if !success {
		t.Fatal("Expected message processing to succeed")
	}

	// Verify rocket was created
	rocket, exists := repo.GetRocket(rocketID)
	if !exists {
		t.Fatal("Expected rocket to be created")
	}

	// Verify rocket properties
	if rocket.ID != rocketID {
		t.Errorf("Expected rocket ID %s, got %s", rocketID, rocket.ID)
	}

	if rocket.Type != msg.Message.Type {
		t.Errorf("Expected rocket type %s, got %s", msg.Message.Type, rocket.Type)
	}

	if rocket.Mission != msg.Message.Mission {
		t.Errorf("Expected mission %s, got %s", msg.Message.Mission, rocket.Mission)
	}

	if rocket.Speed != msg.Message.LaunchSpeed {
		t.Errorf("Expected speed %d, got %d", msg.Message.LaunchSpeed, rocket.Speed)
	}

	if rocket.Exploded != false {
		t.Errorf("Expected exploded to be false, got %v", rocket.Exploded)
	}

	if rocket.LastProcessedMessageNumber != 1 {
		t.Errorf("Expected last processed message number 1, got %d", rocket.LastProcessedMessageNumber)
	}
}

// Test duplicate message handling
func TestDuplicateMessageHandling(t *testing.T) {
	repo := storage.NewRocketRepository()
	rocketID := "test-rocket-2"

	// Create launch message
	msg1 := createTestMessage(rocketID, 1, models.MessageTypeRocketLaunched)
	msg2 := createTestMessage(rocketID, 1, models.MessageTypeRocketLaunched)

	// Process first message
	success1 := repo.ProcessMessage(msg1)
	if !success1 {
		t.Fatal("Expected first message processing to succeed")
	}

	// Process duplicate message (should succeed but not change state)
	success2 := repo.ProcessMessage(msg2)
	if !success2 {
		t.Error("Expected duplicate message processing to succeed (but be ignored)")
	}

	// Verify rocket state hasn't changed
	rocket, exists := repo.GetRocket(rocketID)
	if !exists {
		t.Fatal("Expected rocket to exist")
	}

	if rocket.LastProcessedMessageNumber != 1 {
		t.Errorf("Expected last processed message number 1, got %d", rocket.LastProcessedMessageNumber)
	}

}

// Test out-of-order message processing
func TestOutOfOrderMessageProcessing(t *testing.T) {
	repo := storage.NewRocketRepository()
	rocketID := "test-rocket-3"

	// Create messages in reverse order
	msg1 := createTestMessage(rocketID, 1, models.MessageTypeRocketLaunched)
	msg3 := createTestMessage(rocketID, 3, models.MessageTypeRocketSpeedIncreased)
	msg2 := createTestMessage(rocketID, 2, models.MessageTypeRocketSpeedDecreased)

	// Process message 1 (launch)
	success1 := repo.ProcessMessage(msg1)
	if !success1 {
		t.Fatal("Expected message 1 processing to succeed")
	}

	// Process message 3 (should be pending)
	success3 := repo.ProcessMessage(msg3)
	if !success3 {
		t.Fatal("Expected message 3 processing to succeed (pending)")
	}

	// Verify message 3 is pending

	// Verify rocket state hasn't changed from message 3
	rocket, _ := repo.GetRocket(rocketID)
	if rocket.LastProcessedMessageNumber != 1 {
		t.Errorf("Expected last processed message number 1, got %d", rocket.LastProcessedMessageNumber)
	}

	// Process message 2 (should process both 2 and 3)
	success2 := repo.ProcessMessage(msg2)
	if !success2 {
		t.Fatal("Expected message 2 processing to succeed")
	}

	// Verify all messages were processed
	rocket, _ = repo.GetRocket(rocketID)
	if rocket.LastProcessedMessageNumber != 3 {
		t.Errorf("Expected last processed message number 3, got %d", rocket.LastProcessedMessageNumber)
	}

	// Verify final rocket state includes all changes
	expectedSpeed := msg1.Message.LaunchSpeed - msg2.Message.By + msg3.Message.By
	if rocket.Speed != expectedSpeed {
		t.Errorf("Expected speed %d, got %d", expectedSpeed, rocket.Speed)
	}
}

// Test sequential message processing
func TestSequentialMessageProcessing(t *testing.T) {
	repo := storage.NewRocketRepository()
	rocketID := "test-rocket-4"

	// Create sequential messages
	msg1 := createTestMessage(rocketID, 1, models.MessageTypeRocketLaunched)
	msg2 := createTestMessage(rocketID, 2, models.MessageTypeRocketSpeedIncreased)
	msg3 := createTestMessage(rocketID, 3, models.MessageTypeRocketSpeedDecreased)
	msg4 := createTestMessage(rocketID, 4, models.MessageTypeRocketMissionChanged)
	msg5 := createTestMessage(rocketID, 5, models.MessageTypeRocketExploded)

	messages := []*models.RocketMessage{msg1, msg2, msg3, msg4, msg5}

	// Process all messages in order
	for i, msg := range messages {
		success := repo.ProcessMessage(msg)
		if !success {
			t.Fatalf("Expected message %d processing to succeed", i+1)
		}

		// Verify last processed message number
		rocket, _ := repo.GetRocket(rocketID)
		if rocket.LastProcessedMessageNumber != i+1 {
			t.Errorf("Expected last processed message number %d, got %d", i+1, rocket.LastProcessedMessageNumber)
		}
	}

	// Verify final rocket state
	rocket, _ := repo.GetRocket(rocketID)
	if rocket.Exploded != true {
		t.Errorf("Expected exploded to be true, got %v", rocket.Exploded)
	}

	if rocket.Mission != msg4.Message.NewMission {
		t.Errorf("Expected mission %s, got %s", msg4.Message.NewMission, rocket.Mission)
	}

	if rocket.Reason != msg5.Message.Reason {
		t.Errorf("Expected exploded reason %s, got %s", msg5.Message.Reason, rocket.Reason)
	}
}

// Test invalid message processing
func TestInvalidMessageProcessing(t *testing.T) {
	repo := storage.NewRocketRepository()
	rocketID := "test-rocket-5"

	// Try to process non-launch message first (should be buffered)
	msg := createTestMessage(rocketID, 1, models.MessageTypeRocketSpeedIncreased)
	success := repo.ProcessMessage(msg)
	if !success {
		t.Error("Expected non-launch first message to succeed (but be buffered)")
	}

	// Verify rocket doesn't exist yet
	_, exists := repo.GetRocket(rocketID)
	if exists {
		t.Error("Expected rocket to not exist until launch message")
	}

	// Try to process speed change on exploded rocket
	launchMsg := createTestMessage(rocketID, 1, models.MessageTypeRocketLaunched)
	explodeMsg := createTestMessage(rocketID, 2, models.MessageTypeRocketExploded)
	speedMsg := createTestMessage(rocketID, 3, models.MessageTypeRocketSpeedIncreased)

	repo.ProcessMessage(launchMsg)
	repo.ProcessMessage(explodeMsg)

	success = repo.ProcessMessage(speedMsg)
	if success {
		t.Error("Expected speed change on exploded rocket to fail")
	}
}

// Test concurrent message processing
func TestConcurrentMessageProcessing(t *testing.T) {
	repo := storage.NewRocketRepository()
	rocketID := "test-rocket-6"

	// Create launch message
	launchMsg := createTestMessage(rocketID, 1, models.MessageTypeRocketLaunched)
	repo.ProcessMessage(launchMsg)

	// Create multiple speed change messages
	numMessages := 10
	messages := make([]*models.RocketMessage, numMessages)
	for i := 0; i < numMessages; i++ {
		messages[i] = createTestMessage(rocketID, i+2, models.MessageTypeRocketSpeedIncreased)
	}

	// Process messages concurrently
	done := make(chan bool, numMessages)
	for _, msg := range messages {
		go func(m *models.RocketMessage) {
			repo.ProcessMessage(m)
			done <- true
		}(msg)
	}

	// Wait for all messages to be processed
	for i := 0; i < numMessages; i++ {
		<-done
	}

	// Verify all messages were processed
	rocket, _ := repo.GetRocket(rocketID)
	if rocket.LastProcessedMessageNumber != numMessages+1 {
		t.Errorf("Expected last processed message number %d, got %d", numMessages+1, rocket.LastProcessedMessageNumber)
	}

}

// Test GetAllRockets
func TestGetAllRockets(t *testing.T) {
	repo := storage.NewRocketRepository()

	// Create multiple rockets
	rockets := []string{"rocket-1", "rocket-2", "rocket-3"}
	for _, rocketID := range rockets {
		msg := createTestMessage(rocketID, 1, models.MessageTypeRocketLaunched)
		repo.ProcessMessage(msg)
	}

	// Get all rockets
	allRockets := repo.GetAllRockets()
	if len(allRockets) != len(rockets) {
		t.Errorf("Expected %d rockets, got %d", len(rockets), len(allRockets))
	}

	// Verify all rockets are present
	rocketMap := make(map[string]bool)
	for _, rocket := range allRockets {
		rocketMap[rocket.ID] = true
	}

	for _, rocketID := range rockets {
		if !rocketMap[rocketID] {
			t.Errorf("Expected rocket %s to be present", rocketID)
		}
	}
}

// Test debug methods
func TestDebugMethods(t *testing.T) {
	repo := storage.NewRocketRepository()
	rocketID := "test-rocket-7"

	// Create and process launch message
	msg1 := createTestMessage(rocketID, 1, models.MessageTypeRocketLaunched)
	repo.ProcessMessage(msg1)

	// Create and process pending message
	msg3 := createTestMessage(rocketID, 3, models.MessageTypeRocketSpeedIncreased)
	repo.ProcessMessage(msg3)
}
