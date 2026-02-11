package utils

/*
 * File: internal/utils/errors.go
 *
 * Purpose: modeling the errors that can internally emerge from service providers
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
)

// Error types associated with API failures
var (
	ErrValidationFailed   = errors.New("validation checks did not pass")                            // HTTP 400
	ErrNotAuthorized      = errors.New("insufficient claim to the requested resource")              // HTTP 401
	ErrAccessDenied       = errors.New("insufficient permissions to access the requested resource") // HTTP 403
	ErrNoSuchResource     = errors.New("requested resource does not exist")                         // HTTP 404
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")                                       // HTTP 429
	ErrServiceUnavailable = errors.New("try again later")                                           // HTTP 503
)

// Type `ErrAPIRequestFailed` represents an error from a Tournabyte API service provider
//
// Struct members:
//   - StatusCode: the http status code to mark the response status (typically a 4xx or 5xx code)
//   - CausedBy: contextual information about the error (the underlying error)
//   - Details: additional information regarding the error
type ErrAPIRequestFailed struct {
	CausedBy error
	Details  map[string]string
}

// Function `(ErrAPIRequestFailed).Error` implements the error interface for `ErrAPIRequestFailed`
func (err ErrAPIRequestFailed) Error() string {
	return err.CausedBy.Error()
}

// Function `(ErrAPIRequestFailed).Unwrap` implements the error interface for `ErrAPIRequestFailed`
func (err ErrAPIRequestFailed) Unwrap() error {
	return err.CausedBy
}

// Function `(*ErrAPIRequestFailed).MarshalJSON` implements the JSON marshaler interface for `ErrAPIRequestFailed`
func (err *ErrAPIRequestFailed) MarshalJSON() ([]byte, error) {
	data := make(map[string]any)
	data["message"] = err.Error()
	if len(err.Details) > 0 {
		data["details"] = err.Details
	}
	return json.Marshal(data)
}

// Function `(*ErrAPIRequestFailed).StatusCode` retrieves the appropriate HTTP status code for the underlying error
//
// Returns:
//   - `int`: the corresponding 4xx or 5xx HTTP status code based on the underlying error
func (err *ErrAPIRequestFailed) StatusCode() int {
	switch {
	case errors.Is(err.CausedBy, ErrValidationFailed):
		return http.StatusBadRequest
	case errors.Is(err.CausedBy, ErrNotAuthorized):
		return http.StatusUnauthorized
	case errors.Is(err.CausedBy, ErrAccessDenied):
		return http.StatusForbidden
	case errors.Is(err.CausedBy, ErrNoSuchResource):
		return http.StatusNotFound
	case errors.Is(err.CausedBy, ErrRateLimitExceeded):
		return http.StatusTooManyRequests
	case errors.Is(err.CausedBy, ErrServiceUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// Function `ValidationFailed` constructs an ErrAPIRequestFailed corresponding to a validation failure
//
// Parameters:
//   - ...values: the sequence of string items to populate the details of the error
//
// Returns:
//   - `ErrAPIRequestFailed`: the error response with the appropriate underlying issue
func ValidationFailed(details ...string) ErrAPIRequestFailed {
	return ErrAPIRequestFailed{
		CausedBy: ErrValidationFailed,
		Details:  slicePairwiseToMapping(details...),
	}
}

// Function `NotAuthorized` constructs an ErrAPIRequestFailed corresponding to a authentication failure
//
// Parameters:
//   - ...values: the sequence of string items to populate the details of the error
//
// Returns:
//   - `ErrAPIRequestFailed`: the error response with the appropriate underlying issue
func NotAuthorized(details ...string) ErrAPIRequestFailed {
	return ErrAPIRequestFailed{
		CausedBy: ErrNotAuthorized,
		Details:  slicePairwiseToMapping(details...),
	}
}

// Function `AccessDenied` constructs an ErrAPIRequestFailed corresponding to a permissions failure
//
// Parameters:
//   - ...values: the sequence of string items to populate the details of the error
//
// Returns:
//   - `ErrAPIRequestFailed`: the error response with the appropriate underlying issue
func AccessDenied(details ...string) ErrAPIRequestFailed {
	return ErrAPIRequestFailed{
		CausedBy: ErrAccessDenied,
		Details:  slicePairwiseToMapping(details...),
	}
}

// Function `ResourceNotFound` constructs an ErrAPIRequestFailed corresponding to a missing resource failure
//
// Parameters:
//   - ...values: the sequence of string items to populate the details of the error
//
// Returns:
//   - `ErrAPIRequestFailed`: the error response with the appropriate underlying issue
func ResourceNotFound(details ...string) ErrAPIRequestFailed {
	return ErrAPIRequestFailed{
		CausedBy: ErrNoSuchResource,
		Details:  slicePairwiseToMapping(details...),
	}
}

// Function `SlowDown` constructs an ErrAPIRequestFailed corresponding to a rate limit exceeded error
//
// Parameters:
//   - ...values: the sequence of string items to populate the details of the error
//
// Returns:
//   - `ErrAPIRequestFailed`: the error response with the appropriate underlying issue
func SlowDown(details ...string) ErrAPIRequestFailed {
	return ErrAPIRequestFailed{
		CausedBy: ErrRateLimitExceeded,
		Details:  slicePairwiseToMapping(details...),
	}
}

// Function `TryAgainLater` constructs an ErrAPIRequestFailed corresponding to a service availability failure
//
// Parameters:
//   - ...values: the sequence of string items to populate the details of the error
//
// Returns:
//   - `ErrAPIRequestFailed`: the error response with the appropriate underlying issue
func TryAgainLater(details ...string) ErrAPIRequestFailed {
	return ErrAPIRequestFailed{
		CausedBy: ErrServiceUnavailable,
		Details:  slicePairwiseToMapping(details...),
	}
}

// Function `slicePairwiseToMapping` takes pairs of values from a slice and interprets them as key:value pairs
//
// Parameters:
//   - ...values: the sequence of string items to populate the map
//
// Returns:
//   - `map[string]string`: the mapping of the input slice
func slicePairwiseToMapping(values ...string) map[string]string {
	data := make(map[string]string)

	for pair := range slices.Chunk(values, 2) {
		if len(pair) != 2 {
			break
		}
		key, value := pair[0], pair[1]
		data[key] = value
	}

	return data
}
