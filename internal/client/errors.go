package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError represents a structured error from the ServiceNow API.
type APIError struct {
	StatusCode  int
	Code        string
	Message     string
	Detail      string
	Suggestions []string
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
	if e.Detail != "" {
		msg += " - " + e.Detail
	}
	return msg
}

// snErrorResponse is the ServiceNow error JSON shape.
type snErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Detail  string `json:"detail"`
	} `json:"error"`
}

func parseAPIError(resp *http.Response, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
	}

	var snErr snErrorResponse
	if json.Unmarshal(body, &snErr) == nil && snErr.Error.Message != "" {
		apiErr.Message = snErr.Error.Message
		apiErr.Detail = snErr.Error.Detail
	}

	switch resp.StatusCode {
	case 401:
		apiErr.Code = "AUTH_FAILURE"
		if apiErr.Message == "" {
			apiErr.Message = "Authentication failed"
		}
		apiErr.Suggestions = []string{
			"Check credentials with 'snowctl config view'",
			"For basic auth: verify username and password",
			"For OAuth: token may be expired, try 'snowctl doctor'",
		}
	case 403:
		apiErr.Code = "FORBIDDEN"
		if apiErr.Message == "" {
			apiErr.Message = "Access denied"
		}
		apiErr.Suggestions = []string{
			"The user may lack the required role for this table",
			"Check ACLs on the target table in ServiceNow",
		}
	case 404:
		apiErr.Code = "NOT_FOUND"
		if apiErr.Message == "" {
			apiErr.Message = "Resource not found"
		}
		apiErr.Suggestions = []string{
			"Verify the record sys_id or number is correct",
			"Check if the table name exists in this instance",
		}
	case 429:
		apiErr.Code = "RATE_LIMITED"
		if apiErr.Message == "" {
			apiErr.Message = "Rate limit exceeded"
		}
		apiErr.Suggestions = []string{
			"Wait a moment and retry",
			"Reduce --limit or add --query filters",
		}
	default:
		apiErr.Code = "API_ERROR"
		if apiErr.Message == "" {
			apiErr.Message = http.StatusText(resp.StatusCode)
		}
	}

	return apiErr
}
