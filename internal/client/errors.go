package client

import (
	"fmt"
)

// APIError represents an error returned by the Sanity API.
type APIError struct {
	StatusCode int    `json:"statusCode"`
	ErrorType  string `json:"error"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("Sanity API error (%d - %s): %s [code: %s]", e.StatusCode, e.ErrorType, e.Message, e.Code)
	}
	return fmt.Sprintf("Sanity API error (%d - %s): %s", e.StatusCode, e.ErrorType, e.Message)
}

// IsNotFound returns true if the error represents an HTTP 404 Not Found response.
// This is critical for Terraform state reconciliation: when an asset returns 404,
// Terraform treats it as destroyed outside of Terraform and removes it from state.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 404
	}
	return false
}
