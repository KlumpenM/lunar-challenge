package api

import (
	"encoding/json"
	"log"
	"net/http"

	"lunar-backend-challenge/internal/errors"
	"lunar-backend-challenge/internal/middleware"
	"lunar-backend-challenge/internal/models"
	"lunar-backend-challenge/internal/sorting"
	"lunar-backend-challenge/internal/storage"
	"lunar-backend-challenge/internal/validation"
)

type ApiHandler struct {
	Repository *storage.RocketRepository
}

// MessageResponse represents the response for message processing
type MessageResponse struct {
	Status        string `json:"status" example:"success"`
	Message       string `json:"message" example:"Message processed successfully"`
	RocketID      string `json:"rocketId" example:"193270a9-c9cf-404a-8f83-838e71d9ae67"`
	MessageNumber int    `json:"messageNumber" example:"1"`
}

// DebugInfo provides debugging information about message processing
type DebugInfo struct {
	RocketID              string `json:"rocketId" example:"193270a9-c9cf-404a-8f83-838e71d9ae67"`
	ProcessedMessageCount int    `json:"processedMessageCount" example:"5"`
	PendingMessageCount   int    `json:"pendingMessageCount" example:"2"`
	PendingMessageNumbers []int  `json:"pendingMessageNumbers" example:"1,2,3"`
	LastProcessedMessage  int    `json:"lastProcessedMessage" example:"6"`
}

// NewAPIHandler creates a new API handler
func NewAPIHandler() *ApiHandler {
	return &ApiHandler{
		Repository: storage.NewRocketRepository(),
	}
}

// HandleMessage processes incoming rocket messages
// @Summary Process rocket message
// @Description Processes an incoming rocket message and updates rocket state
// @Tags Messages
// @Accept json
// @Produce json
// @Param message body models.RocketMessage true "Rocket message to process"
// @Success 200 {object} MessageResponse "Message processed successfully"
// @Failure 400 {object} errors.BadRequestError "Invalid request format or validation error"
// @Failure 422 {object} errors.MessageProcessingError "Message processing failed"
// @Router /messages [post]
func (h *ApiHandler) HandleMessage(w http.ResponseWriter, r *http.Request) {
	var message models.RocketMessage

	// Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		log.Printf("Failed to decode JSON: %v", err)
		middleware.WriteErrorResponse(w, errors.NewAPIError(http.StatusBadRequest, "Invalid JSON format", err.Error()))
		return
	}

	// Validate message
	if err := validation.ValidateRocketMessage(&message); err != nil {
		log.Printf("Message validation failed: %v", err)
		middleware.WriteErrorResponse(w, err)
		return
	}

	// Log incoming message for debugging
	log.Printf("Received message: Channel=%s, MsgNum=%d, Type=%s",
		message.GetChannel(), message.GetMessageNumber(), message.GetMessageType())

	// Process the message
	success := h.Repository.ProcessMessage(&message)
	if !success {
		processingErr := errors.NewMessageProcessingError(
			message.GetChannel(),
			message.GetMessageNumber(),
			message.GetMessageType(),
			"Message processing failed - may be duplicate, out-of-order, or invalid state transition",
		)
		log.Printf("Failed to process message: %v", processingErr)
		middleware.WriteErrorResponse(w, processingErr)
		return
	}

	log.Printf("Successfully processed message: Channel=%s, MsgNum=%d, Type=%s",
		message.GetChannel(), message.GetMessageNumber(), message.GetMessageType())

	middleware.WriteSuccessResponse(w, map[string]interface{}{
		"status":        "success",
		"message":       "Message processed successfully",
		"rocketId":      message.GetChannel(),
		"messageNumber": message.GetMessageNumber(),
	})
}

// HandleGetRocket returns a specific rocket by ID
// @Summary Get rocket by ID
// @Description Retrieves detailed information about a specific rocket
// @Tags Rockets
// @Produce json
// @Param id path string true "Rocket ID" example:"193270a9-c9cf-404a-8f83-838e71d9ae67"
// @Success 200 {object} models.RocketState "Rocket details"
// @Failure 400 {object} errors.BadRequestError "Invalid rocket ID format"
// @Failure 404 {object} errors.NotFoundError "Rocket not found"
// @Router /rockets/{id} [get]
func (h *ApiHandler) HandleGetRocket(w http.ResponseWriter, r *http.Request) {
	// Extract rocket ID from URL path parameter
	rocketID := r.PathValue("id")

	// Validate rocket ID
	if err := validation.ValidateRocketID(rocketID); err != nil {
		middleware.WriteErrorResponse(w, err)
		return
	}

	// Get rocket from repository
	rocket, exists := h.Repository.GetRocket(rocketID)
	if !exists {
		middleware.WriteErrorResponse(w, errors.NewAPIError(http.StatusNotFound, "Rocket not found", "No rocket found with ID: "+rocketID))
		return
	}

	middleware.WriteSuccessResponse(w, rocket)
}

// HandleGetRockets returns all rockets with optional sorting
// @Summary List all rockets
// @Description Retrieves a list of all rockets with their current state, with optional sorting
// @Tags Rockets
// @Produce json
// @Param sortBy query string false "Sort field (id, type, speed, mission, exploded, updatedAt)" default(id)
// @Param sortOrder query string false "Sort order (asc, desc)" default(asc)
// @Success 200 {array} models.RocketSummary "List of rockets"
// @Failure 400 {object} errors.BadRequestError "Invalid sorting parameters"
// @Router /rockets [get]
func (h *ApiHandler) HandleGetRockets(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for sorting
	sortBy := r.URL.Query().Get("sortBy")
	sortOrder := r.URL.Query().Get("sortOrder")

	// Validate sorting parameters
	if !sorting.ValidateSortBy(sortBy) {
		middleware.WriteErrorResponse(w, errors.NewAPIError(
			http.StatusBadRequest,
			"Invalid sort field",
			"Valid sort fields are: id, type, speed, mission, exploded, updatedAt",
		))
		return
	}

	// Validate sorting orders
	if !sorting.ValidateSortOrder(sortOrder) {
		middleware.WriteErrorResponse(w, errors.NewAPIError(
			http.StatusBadRequest,
			"Invalid sort order",
			"Valid sort orders are: asc, desc",
		))
		return
	}

	// Get rockets from repository
	rockets := h.Repository.GetAllRockets()

	// Apply sorting
	sortedRockets := sorting.SortRockets(rockets, sortBy, sortOrder)

	middleware.WriteSuccessResponse(w, sortedRockets)
}

// HandleDebugRocket returns debug information for a specific rocket
// @Summary Get debug info for specific rocket
// @Description Retrieves debugging information about message processing for a specific rocket
// @Tags Debug
// @Produce json
// @Param id path string true "Rocket ID" example:"193270a9-c9cf-404a-8f83-838e71d9ae67"
// @Success 200 {object} DebugInfo "Debug information"
// @Failure 400 {object} errors.BadRequestError "Invalid rocket ID format"
// @Failure 404 {object} errors.NotFoundError "Rocket not found"
// @Router /debug/rockets/{id} [get]
func (h *ApiHandler) HandleDebugRocket(w http.ResponseWriter, r *http.Request) {
	// Extract rocket ID from URL path parameter
	rocketID := r.PathValue("id")

	// Validate rocket ID
	if err := validation.ValidateRocketID(rocketID); err != nil {
		middleware.WriteErrorResponse(w, err)
		return
	}

	// Get rocket from repository
	rocket, exists := h.Repository.GetRocket(rocketID)
	if !exists {
		middleware.WriteErrorResponse(w, errors.NewAPIError(http.StatusNotFound, "Rocket not found", "No rocket found with ID: "+rocketID))
		return
	}

	// Get debug information
	processedCount, pendingMessages := h.Repository.GetDebugInfo(rocketID)

	debugInfo := DebugInfo{
		RocketID:              rocketID,
		ProcessedMessageCount: processedCount,
		PendingMessageCount:   len(pendingMessages),
		PendingMessageNumbers: pendingMessages,
		LastProcessedMessage:  rocket.LastProcessedMessageNumber,
	}

	middleware.WriteSuccessResponse(w, debugInfo)
}

// HandleDebugAll returns debug information for all rockets
// @Summary Get debug info for all rockets
// @Description Retrieves debugging information about message processing for all rockets
// @Tags Debug
// @Produce json
// @Success 200 {array} DebugInfo "Debug information for all rockets"
// @Router /debug/rockets [get]
func (h *ApiHandler) HandleDebugAll(w http.ResponseWriter, r *http.Request) {
	rockets := h.Repository.GetAllRockets()
	debugInfos := make([]DebugInfo, len(rockets))

	for i, rocket := range rockets {
		fullRocket, _ := h.Repository.GetRocket(rocket.ID)
		debugInfos[i] = DebugInfo{
			RocketID:             rocket.ID,
			LastProcessedMessage: fullRocket.LastProcessedMessageNumber,
		}
	}

	middleware.WriteSuccessResponse(w, debugInfos)
}
