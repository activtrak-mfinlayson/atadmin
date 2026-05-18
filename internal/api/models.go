// Package api provides the atadmin API client and shared types.
package api

// ErrorResponse represents an error payload returned by the ActivTrak API.
type ErrorResponse struct {
	Message string `json:"message"`
}
