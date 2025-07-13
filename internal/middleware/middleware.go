package middleware

import (
	"encoding/json"
	"log"
	"net/http"

	"lunar-backend-challenge/internal/errors"
)

// ErrorHandler provides centralized error handling
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				WriteErrorResponse(w, errors.NewAPIError(http.StatusInternalServerError, "Internal server error", ""))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// ContentTypeJSON ensures response content type is application/json
func ContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// WriteErrorResponse writes an error response in JSON format
func WriteErrorResponse(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	switch e := err.(type) {
	case errors.APIError:
		w.WriteHeader(e.Code)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    e.Code,
				"message": e.Message,
				"details": e.Details,
			},
		})
	case errors.ValidationError:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": "Validation failed",
				"details": e.Error(),
				"field":   e.Field,
			},
		})
	case errors.MessageProcessingError:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":          http.StatusBadRequest,
				"message":       "Message processing failed",
				"details":       e.Error(),
				"rocketId":      e.RocketID,
				"messageNumber": e.MessageNumber,
				"messageType":   e.MessageType,
			},
		})
	default:
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    http.StatusInternalServerError,
				"message": "Internal server error",
				"details": err.Error(),
			},
		})
	}
}

// WriteSuccessResponse writes a success response in JSON format
func WriteSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if data == nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		return
	}

	json.NewEncoder(w).Encode(data)
}

// ChainMiddleware chains multiple middleware functions
func ChainMiddleware(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
