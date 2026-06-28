package screenshotapi

import "fmt"

// APIError is the base error type for all ScreenshotAPI errors.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("screenshotapi: %s (HTTP %d)", e.Message, e.StatusCode)
}

// AuthenticationError indicates a missing or malformed API key (HTTP 401).
type AuthenticationError struct {
	APIError
}

// InsufficientCreditsError indicates the account has no credits remaining (HTTP 402).
type InsufficientCreditsError struct {
	APIError
	Balance int
}

func (e *InsufficientCreditsError) Error() string {
	return fmt.Sprintf("screenshotapi: %s (balance: %d)", e.Message, e.Balance)
}

// InvalidAPIKeyError indicates the API key is revoked or invalid (HTTP 403).
type InvalidAPIKeyError struct {
	APIError
}

// ScreenshotFailedError indicates the screenshot capture failed server-side (HTTP 500).
type ScreenshotFailedError struct {
	APIError
}
