package errors

import (
	"fmt"
	"net/http"
)

// APIError represents an API error with HTTP status code and message
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// BadRequestError represents a 400 error example
type BadRequestError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"Invalid rocket ID format"`
	Details string `json:"details" example:"Rocket ID must be a valid UUID"`
}

// NotFoundError represents a 404 error example
type NotFoundError struct {
	Code    int    `json:"code" example:"404"`
	Message string `json:"message" example:"Rocket not found"`
	Details string `json:"details" example:"No rocket found with ID: 193270a9-c9cf-404a-8f83-838e71d9ae67"`
}

func (e APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("API Error %d: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("API Error %d: %s", e.Code, e.Message)
}

// ValidationError represents input validation errors
type ValidationError struct {
	Field   string `json:"field" example:"type"`
	Message string `json:"message" example:"Invalid rocket type"`
	Value   string `json:"value,omitempty" example:"Falcon-9"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("Validation error for field '%s': %s", e.Field, e.Message)
}

// MessageProcessingError represents errors in message processing
type MessageProcessingError struct {
	RocketID      string `json:"rocketId" example:"193270a9-c9cf-404a-8f83-838e71d9ae67"`
	MessageNumber int    `json:"messageNumber" example:"3"`
	MessageType   string `json:"messageType" example:"RocketSpeedIncreased"`
	Reason        string `json:"reason" example:"Message processing failed - may be duplicate, out-of-order, or invalid state transition"`
}

func (e MessageProcessingError) Error() string {
	return fmt.Sprintf("Failed to process message %d for rocket %s (type: %s): %s",
		e.MessageNumber, e.RocketID, e.MessageType, e.Reason)
}

// Pre-defined API errors
var (
	ErrInvalidJSON = APIError{
		Code:    http.StatusBadRequest,
		Message: "Invalid JSON format",
	}

	ErrMethodNotAllowed = APIError{
		Code:    http.StatusMethodNotAllowed,
		Message: "HTTP method not allowed",
	}

	ErrRocketNotFound = APIError{
		Code:    http.StatusNotFound,
		Message: "Rocket not found",
	}

	ErrMissingRocketID = APIError{
		Code:    http.StatusBadRequest,
		Message: "Rocket ID is required",
	}

	ErrInternalServer = APIError{
		Code:    http.StatusInternalServerError,
		Message: "Internal server error",
	}
)

// NewValidationError creates a new validation error
func NewValidationError(field, message string, value ...string) ValidationError {
	var val string
	if len(value) > 0 {
		val = value[0]
	}
	return ValidationError{
		Field:   field,
		Message: message,
		Value:   val,
	}
}

// NewMessageProcessingError creates a new message processing error
func NewMessageProcessingError(rocketID string, messageNumber int, messageType, reason string) MessageProcessingError {
	return MessageProcessingError{
		RocketID:      rocketID,
		MessageNumber: messageNumber,
		MessageType:   messageType,
		Reason:        reason,
	}
}

// NewAPIError creates a new API error with custom details
func NewAPIError(code int, message, details string) APIError {
	return APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}
