package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"lunar-backend-challenge/internal/api"
	"lunar-backend-challenge/internal/models"
)

// Helper function to create test messages for handlers
func createTestHTTPMessage(channel string, messageNumber int, messageType string) *models.RocketMessage {
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

// Helper function to create JSON request body
func createJSONRequestBody(t *testing.T, data interface{}) *bytes.Buffer {
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	return bytes.NewBuffer(jsonData)
}

// Test HandleMessage - successful processing
func TestHandleMessage_Success(t *testing.T) {
	handler := api.NewAPIHandler()

	// Create test message
	msg := createTestHTTPMessage("test-rocket-1", 1, models.MessageTypeRocketLaunched)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/messages", createJSONRequestBody(t, msg))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.HandleMessage(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check content type
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected content type application/json, got %s", rr.Header().Get("Content-Type"))
	}

	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check response fields
	if response["status"] != "success" {
		t.Errorf("Expected status success, got %v", response["status"])
	}

	if response["rocketId"] != "test-rocket-1" {
		t.Errorf("Expected rocketId test-rocket-1, got %v", response["rocketId"])
	}
}

// Test HandleMessage - invalid JSON
func TestHandleMessage_InvalidJSON(t *testing.T) {
	handler := api.NewAPIHandler()

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.HandleMessage(rr, req)

	// Check status code
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	// Parse error response
	var errorResponse map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	// Check error structure
	if errorResponse["error"] == nil {
		t.Error("Expected error field in response")
	}
}

// Test HandleMessage - validation error
func TestHandleMessage_ValidationError(t *testing.T) {
	handler := api.NewAPIHandler()

	// Create message with missing required fields
	msg := &models.RocketMessage{}
	msg.Metadata.Channel = "" // Missing channel
	msg.Metadata.MessageNumber = 1
	msg.Metadata.MessageType = models.MessageTypeRocketLaunched
	msg.Metadata.MessageTime = time.Now()

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/messages", createJSONRequestBody(t, msg))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.HandleMessage(rr, req)

	// Check status code
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	// Parse error response
	var errorResponse map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	// Check error structure
	if errorResponse["error"] == nil {
		t.Error("Expected error field in response")
	}
}

// Test HandleGetRockets - empty list
func TestHandleGetRockets_EmptyList(t *testing.T) {
	handler := api.NewAPIHandler()

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/rockets", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.HandleGetRockets(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Parse response
	var rockets []models.RocketSummary
	if err := json.NewDecoder(rr.Body).Decode(&rockets); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check empty list
	if len(rockets) != 0 {
		t.Errorf("Expected empty list, got %d rockets", len(rockets))
	}
}

// Test HandleGetRockets - with rockets
func TestHandleGetRockets_WithRockets(t *testing.T) {
	handler := api.NewAPIHandler()

	// Create test rockets
	rocketIDs := []string{"rocket-1", "rocket-2", "rocket-3"}
	for _, rocketID := range rocketIDs {
		msg := createTestHTTPMessage(rocketID, 1, models.MessageTypeRocketLaunched)
		handler.Repository.ProcessMessage(msg)
	}

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/rockets", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.HandleGetRockets(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Parse response
	var rockets []models.RocketSummary
	if err := json.NewDecoder(rr.Body).Decode(&rockets); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check rocket count
	if len(rockets) != len(rocketIDs) {
		t.Errorf("Expected %d rockets, got %d", len(rocketIDs), len(rockets))
	}

	// Verify rocket IDs
	rocketMap := make(map[string]bool)
	for _, rocket := range rockets {
		rocketMap[rocket.ID] = true
	}

	for _, expectedID := range rocketIDs {
		if !rocketMap[expectedID] {
			t.Errorf("Expected rocket %s to be present", expectedID)
		}
	}
}

// Test HandleGetRocket - successful retrieval
func TestHandleGetRocket_Success(t *testing.T) {
	handler := api.NewAPIHandler()
	rocketID := "test-rocket-1"

	// Create test rocket
	msg := createTestHTTPMessage(rocketID, 1, models.MessageTypeRocketLaunched)
	handler.Repository.ProcessMessage(msg)

	// Create request with path parameter
	req := httptest.NewRequest(http.MethodGet, "/rockets/"+rocketID, nil)
	req.SetPathValue("id", rocketID)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.HandleGetRocket(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Parse response
	var rocket models.RocketState
	if err := json.NewDecoder(rr.Body).Decode(&rocket); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check rocket properties
	if rocket.ID != rocketID {
		t.Errorf("Expected rocket ID %s, got %s", rocketID, rocket.ID)
	}

	if rocket.Type != msg.Message.Type {
		t.Errorf("Expected rocket type %s, got %s", msg.Message.Type, rocket.Type)
	}
}

// Test HandleGetRocket - rocket not found
func TestHandleGetRocket_NotFound(t *testing.T) {
	handler := api.NewAPIHandler()
	rocketID := "non-existent-rocket"

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/rockets/"+rocketID, nil)
	req.SetPathValue("id", rocketID)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.HandleGetRocket(rr, req)

	// Check status code
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rr.Code)
	}

	// Parse error response
	var errorResponse map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	// Check error structure
	if errorResponse["error"] == nil {
		t.Error("Expected error field in response")
	}
}

// Test HandleGetRocket - invalid rocket ID
func TestHandleGetRocket_InvalidID(t *testing.T) {
	handler := api.NewAPIHandler()
	invalidID := "abc" // Too short

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/rockets/"+invalidID, nil)
	req.SetPathValue("id", invalidID)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.HandleGetRocket(rr, req)

	// Check status code
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rr.Code)
	}

	// Parse error response
	var errorResponse map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	// Check error structure
	if errorResponse["error"] == nil {
		t.Error("Expected error field in response")
	}
}

// Test HandleDebugRocket - successful debug info
func TestHandleDebugRocket_Success(t *testing.T) {
	handler := api.NewAPIHandler()
	rocketID := "test-rocket-1"

	// Create test rocket with launch message
	msg1 := createTestHTTPMessage(rocketID, 1, models.MessageTypeRocketLaunched)
	handler.Repository.ProcessMessage(msg1)

	// Create out-of-order message
	msg3 := createTestHTTPMessage(rocketID, 3, models.MessageTypeRocketSpeedIncreased)
	handler.Repository.ProcessMessage(msg3)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/debug/rockets/"+rocketID, nil)
	req.SetPathValue("id", rocketID)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.HandleDebugRocket(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Parse response
	var debugInfo api.DebugInfo
	if err := json.NewDecoder(rr.Body).Decode(&debugInfo); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check debug info
	if debugInfo.RocketID != rocketID {
		t.Errorf("Expected rocket ID %s, got %s", rocketID, debugInfo.RocketID)
	}

	if debugInfo.ProcessedMessageCount != 1 {
		t.Errorf("Expected processed message count 1, got %d", debugInfo.ProcessedMessageCount)
	}

	if debugInfo.PendingMessageCount != 1 {
		t.Errorf("Expected pending message count 1, got %d", debugInfo.PendingMessageCount)
	}

	if len(debugInfo.PendingMessageNumbers) != 1 || debugInfo.PendingMessageNumbers[0] != 3 {
		t.Errorf("Expected pending message numbers [3], got %v", debugInfo.PendingMessageNumbers)
	}
}

// Test HandleDebugAll - debug info for all rockets
func TestHandleDebugAll_Success(t *testing.T) {
	handler := api.NewAPIHandler()

	// Create test rockets
	rocketIDs := []string{"rocket-1", "rocket-2"}
	for _, rocketID := range rocketIDs {
		msg := createTestHTTPMessage(rocketID, 1, models.MessageTypeRocketLaunched)
		handler.Repository.ProcessMessage(msg)
	}

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/debug/rockets", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.HandleDebugAll(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Parse response
	var debugInfos []api.DebugInfo
	if err := json.NewDecoder(rr.Body).Decode(&debugInfos); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check debug info count
	if len(debugInfos) != len(rocketIDs) {
		t.Errorf("Expected %d debug infos, got %d", len(rocketIDs), len(debugInfos))
	}

	// Verify rocket IDs
	debugMap := make(map[string]bool)
	for _, debugInfo := range debugInfos {
		debugMap[debugInfo.RocketID] = true
	}

	for _, expectedID := range rocketIDs {
		if !debugMap[expectedID] {
			t.Errorf("Expected debug info for rocket %s", expectedID)
		}
	}
}

// Test message processing flow
func TestMessageProcessingFlow(t *testing.T) {
	handler := api.NewAPIHandler()
	rocketID := "test-rocket-flow"

	// Test messages in sequence
	messages := []struct {
		messageNumber int
		messageType   string
		expectedCode  int
	}{
		{1, models.MessageTypeRocketLaunched, http.StatusOK},
		{2, models.MessageTypeRocketSpeedIncreased, http.StatusOK},
		{3, models.MessageTypeRocketSpeedDecreased, http.StatusOK},
		{4, models.MessageTypeRocketMissionChanged, http.StatusOK},
		{5, models.MessageTypeRocketExploded, http.StatusOK},
	}

	for _, test := range messages {
		// Create message
		msg := createTestHTTPMessage(rocketID, test.messageNumber, test.messageType)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/messages", createJSONRequestBody(t, msg))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		rr := httptest.NewRecorder()

		// Call handler
		handler.HandleMessage(rr, req)

		// Check status code
		if rr.Code != test.expectedCode {
			t.Errorf("Message %d (%s): Expected status %d, got %d",
				test.messageNumber, test.messageType, test.expectedCode, rr.Code)
		}
	}

	// Verify final rocket state
	req := httptest.NewRequest(http.MethodGet, "/rockets/"+rocketID, nil)
	req.SetPathValue("id", rocketID)
	rr := httptest.NewRecorder()

	handler.HandleGetRocket(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var rocket models.RocketState
	if err := json.NewDecoder(rr.Body).Decode(&rocket); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check final state
	if !rocket.Exploded {
		t.Error("Expected rocket to be exploded")
	}

	// Check debug info for processed message count (LastProcessedMessageNumber is not exposed in JSON)
	debugReq := httptest.NewRequest(http.MethodGet, "/debug/rockets/"+rocketID, nil)
	debugReq.SetPathValue("id", rocketID)
	debugRr := httptest.NewRecorder()

	handler.HandleDebugRocket(debugRr, debugReq)

	if debugRr.Code != http.StatusOK {
		t.Errorf("Expected debug status %d, got %d", http.StatusOK, debugRr.Code)
	}

	var debugInfo api.DebugInfo
	if err := json.NewDecoder(debugRr.Body).Decode(&debugInfo); err != nil {
		t.Fatalf("Failed to decode debug response: %v", err)
	}

	if debugInfo.LastProcessedMessage != 5 {
		t.Errorf("Expected last processed message number 5, got %d", debugInfo.LastProcessedMessage)
	}
}
