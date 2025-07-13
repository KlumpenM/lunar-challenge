package sorting

import (
	"sort"
	"strings"

	"lunar-backend-challenge/internal/models"
)

// Valid sorting options
var ValidSortOptions = map[string]bool{
	"id":        true,
	"type":      true,
	"speed":     true,
	"mission":   true,
	"exploded":  true,
	"updatedAt": true,
}

// Valid sorting orders
var ValidSortOrders = map[string]bool{
	"asc":  true,
	"desc": true,
}

// ValidateSortOrder validates if a sort order is valid
func ValidateSortOrder(order string) bool {
	if order == "" {
		return true // Default to asc
	}
	return ValidSortOrders[order]
}

// ValidateSortBy validates if a sort field is valid
func ValidateSortBy(field string) bool {
	if field == "" {
		return true // Default sorting allowed
	}
	return ValidSortOptions[field]
}

// SortRockets sorts a slice of RocketSummary based on the specified field and order
func SortRockets(rockets []models.RocketSummary, sortBy, sortOrder string) []models.RocketSummary {
	// Default values
	if sortBy == "" {
		sortBy = "id" // Default sort by ID
	}
	if sortOrder == "" {
		sortOrder = "asc" // Default to ascending
	}

	// Make a copy to avoid modifying the original slice
	sortedRockets := make([]models.RocketSummary, len(rockets))
	copy(sortedRockets, rockets)

	// Sort based on the specified field and order
	sort.Slice(sortedRockets, func(i, j int) bool {
		var result bool

		switch sortBy {
		case "id":
			result = strings.ToLower(sortedRockets[i].ID) < strings.ToLower(sortedRockets[j].ID)
		case "type":
			result = strings.ToLower(sortedRockets[i].Type) < strings.ToLower(sortedRockets[j].Type)
		case "speed":
			result = sortedRockets[i].Speed < sortedRockets[j].Speed
		case "mission":
			result = strings.ToLower(sortedRockets[i].Mission) < strings.ToLower(sortedRockets[j].Mission)
		case "exploded":
			// Sort exploded rockets to the end when ascending, beginning when descending
			if sortedRockets[i].Exploded != sortedRockets[j].Exploded {
				result = !sortedRockets[i].Exploded && sortedRockets[j].Exploded
			} else {
				// If both have same exploded status, sort by ID as secondary
				result = strings.ToLower(sortedRockets[i].ID) < strings.ToLower(sortedRockets[j].ID)
			}
		case "updatedAt":
			result = sortedRockets[i].UpdatedAt.Before(sortedRockets[j].UpdatedAt)
		default:
			// Fallback to ID sorting
			result = strings.ToLower(sortedRockets[i].ID) < strings.ToLower(sortedRockets[j].ID)
		}

		// Reverse the result if descending order
		if sortOrder == "desc" {
			result = !result
		}

		return result
	})

	return sortedRockets
}
