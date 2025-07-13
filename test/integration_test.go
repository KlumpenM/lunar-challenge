package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"lunar-backend-challenge/internal/api"
	"lunar-backend-challenge/internal/middleware"
	"lunar-backend-challenge/internal/models"
)

// TestServer creates a test server with the same setup as main
func createTestServer() *httptest.Server {
	// Create the API handler
	apiHandler := api.NewAPIHandler()

	// Create a new ServeMux (Go 1.22+ features)
	mux := http.NewServeMux()

	// Set up API routes with Go 1.22+ patterns
	mux.HandleFunc("POST /messages", apiHandler.HandleMessage)
	mux.HandleFunc("GET /rockets", apiHandler.HandleGetRockets)
	mux.HandleFunc("GET /rockets/{id}", apiHandler.HandleGetRocket)

	// Debug routes
	mux.HandleFunc("GET /debug/rockets", apiHandler.HandleDebugAll)
	mux.HandleFunc("GET /debug/rockets/{id}", apiHandler.HandleDebugRocket)

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		middleware.WriteSuccessResponse(w, map[string]interface{}{
			"status":    "healthy",
			"service":   "lunar-rocket-api",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC(),
		})
	})

	// Apply middleware chain
	handler := middleware.ChainMiddleware(mux,
		middleware.ErrorHandler,
		middleware.ContentTypeJSON,
	)

	return httptest.NewServer(handler)
}

// Helper function to create test messages
func createIntegrationTestMessage(channel string, messageNumber int, messageType string) *models.RocketMessage {
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

// TestingInterface defines methods common to both *testing.T and *testing.B
type TestingInterface interface {
	Fatalf(format string, args ...interface{})
}

// Helper function to send HTTP requests
// Helper function to send HTTP requests
func sendHTTPRequest(t TestingInterface, method, url string, body interface{}) *http.Response {
	var req *http.Request
	var err error

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		req, _ = http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	return resp
}

// Test health check endpoint
func TestIntegration_HealthCheck(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	resp := sendHTTPRequest(t, "GET", server.URL+"/health", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var healthResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}

	if healthResp["status"] != "healthy" {
		t.Errorf("Expected status healthy, got %v", healthResp["status"])
	}

	if healthResp["service"] != "lunar-rocket-api" {
		t.Errorf("Expected service lunar-rocket-api, got %v", healthResp["service"])
	}
}

// Test single rocket lifecycle
func TestIntegration_SingleRocketLifecycle(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	rocketID := "integration-test-rocket"

	// 1. Launch rocket
	launchMsg := createIntegrationTestMessage(rocketID, 1, models.MessageTypeRocketLaunched)
	resp := sendHTTPRequest(t, "POST", server.URL+"/messages", launchMsg)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected launch status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// 2. Verify rocket exists
	resp = sendHTTPRequest(t, "GET", server.URL+"/rockets/"+rocketID, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected get rocket status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var rocket models.RocketState
	if err := json.NewDecoder(resp.Body).Decode(&rocket); err != nil {
		t.Fatalf("Failed to decode rocket response: %v", err)
	}

	if rocket.ID != rocketID {
		t.Errorf("Expected rocket ID %s, got %s", rocketID, rocket.ID)
	}

	if rocket.Speed != launchMsg.Message.LaunchSpeed {
		t.Errorf("Expected speed %d, got %d", launchMsg.Message.LaunchSpeed, rocket.Speed)
	}

	// 3. Increase speed
	speedMsg := createIntegrationTestMessage(rocketID, 2, models.MessageTypeRocketSpeedIncreased)
	resp = sendHTTPRequest(t, "POST", server.URL+"/messages", speedMsg)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected speed increase status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// 4. Verify speed change
	resp = sendHTTPRequest(t, "GET", server.URL+"/rockets/"+rocketID, nil)
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&rocket); err != nil {
		t.Fatalf("Failed to decode rocket response: %v", err)
	}

	expectedSpeed := launchMsg.Message.LaunchSpeed + speedMsg.Message.By
	if rocket.Speed != expectedSpeed {
		t.Errorf("Expected speed %d, got %d", expectedSpeed, rocket.Speed)
	}

	// 5. Explode rocket
	explodeMsg := createIntegrationTestMessage(rocketID, 3, models.MessageTypeRocketExploded)
	resp = sendHTTPRequest(t, "POST", server.URL+"/messages", explodeMsg)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected explode status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// 6. Verify explosion
	resp = sendHTTPRequest(t, "GET", server.URL+"/rockets/"+rocketID, nil)
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&rocket); err != nil {
		t.Fatalf("Failed to decode rocket response: %v", err)
	}

	if !rocket.Exploded {
		t.Error("Expected rocket to be exploded")
	}

	if rocket.Reason != explodeMsg.Message.Reason {
		t.Errorf("Expected reason %s, got %s", explodeMsg.Message.Reason, rocket.Reason)
	}
}

// Test out-of-order message processing
func TestIntegration_OutOfOrderMessages(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	rocketID := "out-of-order-test-rocket"

	// Create messages in reverse order
	msg1 := createIntegrationTestMessage(rocketID, 1, models.MessageTypeRocketLaunched)
	msg2 := createIntegrationTestMessage(rocketID, 2, models.MessageTypeRocketSpeedIncreased)
	msg3 := createIntegrationTestMessage(rocketID, 3, models.MessageTypeRocketSpeedDecreased)

	// Send in reverse order: 3, 1, 2
	resp := sendHTTPRequest(t, "POST", server.URL+"/messages", msg3)
	resp.Body.Close()

	resp = sendHTTPRequest(t, "POST", server.URL+"/messages", msg1)
	resp.Body.Close()

	resp = sendHTTPRequest(t, "POST", server.URL+"/messages", msg2)
	resp.Body.Close()

	// Verify final state
	resp = sendHTTPRequest(t, "GET", server.URL+"/rockets/"+rocketID, nil)
	defer resp.Body.Close()

	var rocket models.RocketState
	if err := json.NewDecoder(resp.Body).Decode(&rocket); err != nil {
		t.Fatalf("Failed to decode rocket response: %v", err)
	}

	// Check final speed includes all changes
	expectedSpeed := msg1.Message.LaunchSpeed + msg2.Message.By - msg3.Message.By
	if rocket.Speed != expectedSpeed {
		t.Errorf("Expected speed %d, got %d", expectedSpeed, rocket.Speed)
	}

	// Check debug info
	resp = sendHTTPRequest(t, "GET", server.URL+"/debug/rockets/"+rocketID, nil)
	defer resp.Body.Close()

	var debugInfo api.DebugInfo
	if err := json.NewDecoder(resp.Body).Decode(&debugInfo); err != nil {
		t.Fatalf("Failed to decode debug response: %v", err)
	}

	if debugInfo.ProcessedMessageCount != 3 {
		t.Errorf("Expected 3 processed messages, got %d", debugInfo.ProcessedMessageCount)
	}

	if debugInfo.PendingMessageCount != 0 {
		t.Errorf("Expected 0 pending messages, got %d", debugInfo.PendingMessageCount)
	}
}

// Test concurrent message processing
func TestIntegration_ConcurrentMessages(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	rocketID := "concurrent-test-rocket"
	numMessages := 20

	// Create launch message
	launchMsg := createIntegrationTestMessage(rocketID, 1, models.MessageTypeRocketLaunched)
	resp := sendHTTPRequest(t, "POST", server.URL+"/messages", launchMsg)
	resp.Body.Close()

	// Create speed increase messages
	messages := make([]*models.RocketMessage, numMessages)
	for i := 0; i < numMessages; i++ {
		messages[i] = createIntegrationTestMessage(rocketID, i+2, models.MessageTypeRocketSpeedIncreased)
	}

	// Send messages concurrently
	var wg sync.WaitGroup
	for _, msg := range messages {
		wg.Add(1)
		go func(m *models.RocketMessage) {
			defer wg.Done()
			resp := sendHTTPRequest(t, "POST", server.URL+"/messages", m)
			resp.Body.Close()
		}(msg)
	}

	wg.Wait()

	// Verify all messages were processed
	resp = sendHTTPRequest(t, "GET", server.URL+"/debug/rockets/"+rocketID, nil)
	defer resp.Body.Close()

	var debugInfo api.DebugInfo
	if err := json.NewDecoder(resp.Body).Decode(&debugInfo); err != nil {
		t.Fatalf("Failed to decode debug response: %v", err)
	}

	expectedProcessedCount := numMessages + 1 // +1 for launch message
	if debugInfo.ProcessedMessageCount != expectedProcessedCount {
		t.Errorf("Expected %d processed messages, got %d", expectedProcessedCount, debugInfo.ProcessedMessageCount)
	}

	if debugInfo.PendingMessageCount != 0 {
		t.Errorf("Expected 0 pending messages, got %d", debugInfo.PendingMessageCount)
	}
}

// Test multiple rockets
func TestIntegration_MultipleRockets(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	rocketIDs := []string{"multi-rocket-1", "multi-rocket-2", "multi-rocket-3"}

	// Create rockets
	for _, rocketID := range rocketIDs {
		launchMsg := createIntegrationTestMessage(rocketID, 1, models.MessageTypeRocketLaunched)
		resp := sendHTTPRequest(t, "POST", server.URL+"/messages", launchMsg)
		resp.Body.Close()
	}

	// Get all rockets
	resp := sendHTTPRequest(t, "GET", server.URL+"/rockets", nil)
	defer resp.Body.Close()

	var rockets []models.RocketSummary
	if err := json.NewDecoder(resp.Body).Decode(&rockets); err != nil {
		t.Fatalf("Failed to decode rockets response: %v", err)
	}

	if len(rockets) != len(rocketIDs) {
		t.Errorf("Expected %d rockets, got %d", len(rocketIDs), len(rockets))
	}

	// Verify all rockets are present
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

// Test error handling
func TestIntegration_ErrorHandling(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	// Test invalid JSON
	resp := sendHTTPRequest(t, "POST", server.URL+"/messages", "invalid json")
	resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid JSON, got %d", http.StatusBadRequest, resp.StatusCode)
	}

	// Test missing rocket
	resp = sendHTTPRequest(t, "GET", server.URL+"/rockets/non-existent", nil)
	resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %d for missing rocket, got %d", http.StatusNotFound, resp.StatusCode)
	}

	// Test invalid rocket ID
	resp = sendHTTPRequest(t, "GET", server.URL+"/rockets/ab", nil)
	resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid rocket ID, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

// Test duplicate message handling
func TestIntegration_DuplicateMessages(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	rocketID := "duplicate-test-rocket"

	// Create and send launch message
	launchMsg := createIntegrationTestMessage(rocketID, 1, models.MessageTypeRocketLaunched)
	resp := sendHTTPRequest(t, "POST", server.URL+"/messages", launchMsg)
	resp.Body.Close()

	// Send duplicate message
	resp = sendHTTPRequest(t, "POST", server.URL+"/messages", launchMsg)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d for duplicate message, got %d", http.StatusOK, resp.StatusCode)
	}

	// Verify only one message was processed
	resp = sendHTTPRequest(t, "GET", server.URL+"/debug/rockets/"+rocketID, nil)
	defer resp.Body.Close()

	var debugInfo api.DebugInfo
	if err := json.NewDecoder(resp.Body).Decode(&debugInfo); err != nil {
		t.Fatalf("Failed to decode debug response: %v", err)
	}

	if debugInfo.ProcessedMessageCount != 1 {
		t.Errorf("Expected 1 processed message, got %d", debugInfo.ProcessedMessageCount)
	}
}

// Test timeout handling
func TestIntegration_TimeoutHandling(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", server.URL+"/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

// Benchmark message processing
func BenchmarkIntegration_MessageProcessing(b *testing.B) {
	server := createTestServer()
	defer server.Close()

	rocketID := "benchmark-rocket"

	// Create launch message
	launchMsg := createIntegrationTestMessage(rocketID, 1, models.MessageTypeRocketLaunched)
	resp := sendHTTPRequest(b, "POST", server.URL+"/messages", launchMsg)
	resp.Body.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := createIntegrationTestMessage(rocketID, i+2, models.MessageTypeRocketSpeedIncreased)
		resp := sendHTTPRequest(b, "POST", server.URL+"/messages", msg)
		resp.Body.Close()
	}
}
