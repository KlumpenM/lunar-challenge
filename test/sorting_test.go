package test

import (
	"testing"
	"time"

	"lunar-backend-challenge/internal/models"
	"lunar-backend-challenge/internal/sorting"
)

func TestValidateSortBy(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected bool
	}{
		{"Empty field", "", true},
		{"Valid field - id", "id", true},
		{"Valid field - type", "type", true},
		{"Valid field - speed", "speed", true},
		{"Valid field - mission", "mission", true},
		{"Valid field - exploded", "exploded", true},
		{"Valid field - updatedAt", "updatedAt", true},
		{"Invalid field", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sorting.ValidateSortBy(tt.field); got != tt.expected {
				t.Errorf("ValidateSortBy(%q) = %v, want %v", tt.field, got, tt.expected)
			}
		})
	}
}

func TestValidateSortOrder(t *testing.T) {
	tests := []struct {
		name     string
		order    string
		expected bool
	}{
		{"Empty order", "", true},
		{"Valid order - asc", "asc", true},
		{"Valid order - desc", "desc", true},
		{"Invalid order", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sorting.ValidateSortOrder(tt.order); got != tt.expected {
				t.Errorf("ValidateSortOrder(%q) = %v, want %v", tt.order, got, tt.expected)
			}
		})
	}
}

func TestSortRockets(t *testing.T) {
	now := time.Now()
	rockets := []models.RocketSummary{
		{ID: "rocket-2", Type: "Falcon-9", Speed: 200, Mission: "Beta", Exploded: false, UpdatedAt: now.Add(-1 * time.Hour)},
		{ID: "rocket-1", Type: "Atlas", Speed: 100, Mission: "Alpha", Exploded: true, UpdatedAt: now},
		{ID: "rocket-3", Type: "Delta", Speed: 300, Mission: "Charlie", Exploded: false, UpdatedAt: now.Add(-2 * time.Hour)},
	}

	tests := []struct {
		name          string
		sortBy        string
		sortOrder     string
		checkOrdering func([]models.RocketSummary) bool
	}{
		{
			name:      "Sort by ID ascending",
			sortBy:    "id",
			sortOrder: "asc",
			checkOrdering: func(rockets []models.RocketSummary) bool {
				return rockets[0].ID == "rocket-1" && rockets[1].ID == "rocket-2" && rockets[2].ID == "rocket-3"
			},
		},
		{
			name:      "Sort by speed descending",
			sortBy:    "speed",
			sortOrder: "desc",
			checkOrdering: func(rockets []models.RocketSummary) bool {
				return rockets[0].Speed == 300 && rockets[1].Speed == 200 && rockets[2].Speed == 100
			},
		},
		{
			name:      "Sort by exploded status",
			sortBy:    "exploded",
			sortOrder: "asc",
			checkOrdering: func(rockets []models.RocketSummary) bool {
				return !rockets[0].Exploded && !rockets[1].Exploded && rockets[2].Exploded
			},
		},
		{
			name:      "Sort by updatedAt",
			sortBy:    "updatedAt",
			sortOrder: "desc",
			checkOrdering: func(rockets []models.RocketSummary) bool {
				return rockets[0].UpdatedAt.After(rockets[1].UpdatedAt) && rockets[1].UpdatedAt.After(rockets[2].UpdatedAt)
			},
		},
		{
			name:      "Default sort (by ID) when empty field",
			sortBy:    "",
			sortOrder: "",
			checkOrdering: func(rockets []models.RocketSummary) bool {
				return rockets[0].ID == "rocket-1" && rockets[1].ID == "rocket-2" && rockets[2].ID == "rocket-3"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorted := sorting.SortRockets(rockets, tt.sortBy, tt.sortOrder)
			if !tt.checkOrdering(sorted) {
				t.Errorf("SortRockets() with sortBy=%q, sortOrder=%q failed ordering check", tt.sortBy, tt.sortOrder)
			}
			if len(sorted) != len(rockets) {
				t.Errorf("SortRockets() changed slice length, got %d, want %d", len(sorted), len(rockets))
			}
		})
	}
}
