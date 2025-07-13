package models

import "time"

// RocketState represents the state of a rocket
type RocketState struct {
	ID                         string    `json:"id" example:"193270a9-c9cf-404a-8f83-838e71d9ae67"`   // Rocket channel ID (unique identifier)
	Type                       string    `json:"type" example:"Falcon-9"`                             // Rocket type (e.g. "Falcon-9")
	Speed                      int       `json:"speed" example:"3500"`                                // Current speed
	Mission                    string    `json:"mission" example:"ARTEMIS"`                           // Current mission
	Exploded                   bool      `json:"exploded" example:"false"`                            // Status: "exploded"
	Reason                     string    `json:"reason,omitempty" example:""`                         // Reason for explosion (only if exploded)
	CreatedAt                  time.Time `json:"createdAt" example:"2024-03-14T19:39:05.86337+01:00"` // Time of first launch
	UpdatedAt                  time.Time `json:"updatedAt" example:"2024-03-14T19:45:12.12345+01:00"` // Time of last update
	LastProcessedMessageNumber int       `json:"-"`                                                   // Track message ordering (not exposed in JSON)
}

// RocketSummary, for listing purpose
type RocketSummary struct {
	ID        string    `json:"id" example:"193270a9-c9cf-404a-8f83-838e71d9ae67"`
	Type      string    `json:"type" example:"Falcon-9"`
	Speed     int       `json:"speed" example:"3500"`
	Mission   string    `json:"mission" example:"ARTEMIS"`
	Exploded  bool      `json:"exploded" example:"false"`
	UpdatedAt time.Time `json:"updatedAt" example:"2024-03-14T19:45:12.12345+01:00"`
}
