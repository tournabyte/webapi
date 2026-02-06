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
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Type `APIError` is the base error structure for the Tournabyte API responses
//
// Struct members:
//   - Cause: the underlying error from the API service (expected to be an error string for seemless encoding)
//   - Timestamp: the time the error emerged from the service
type APIError struct {
	Cause     error `json:"message"`
	Timestamp int64 `json:"timestamp"`
}

// Function `APIError.Error` implements the error interface for `APIError`
func (err APIError) Error() string {
	if err.Cause != nil {
		return err.Cause.Error()
	}
	return fmt.Sprintf("Request encountered an error at %d", err.Timestamp)
}

// Function `APIError.Unwrap` implements the error interface for `APIError`
func (err APIError) Unwrap() error {
	return err.Cause
}

// Type `ValidationError` represents a issue occurred during validation when processing a request
//
// Struct members
//   - APIError: (inherited)
//   - Details: mapping of issue source to issue explanation (i.e. {"some_field": "could not be parsed as a number"})
type ValidationError struct {
	APIError
	Details gin.H `json:"details"`
}

// Type `AuthorizationError` represents an authorization issue occurred during request processing
//
// Struct members:
//   - APIError: (inherited)
//   - Reconcile: indicates whether the issue can be reconciled (true means client should reauthenticate)
type AuthorizationError struct {
	APIError
	Reconcile bool `json:"retryable"`
}

// Type `NoSuchResource` represents a resource not found error during request processing
//
// Struct members:
//   - APIError: (inherited)
//   - Resource: the resource that was requested and not found
type NoSuchResource struct {
	APIError
	Resource string `json:"resource"`
}

// Type `ServiceUnavailable` represents a service availability issue encountered during request processing
//
// Struct members:
//   - APIError: (inherited)
//   - ServiceName: service that is currently unavailable
type ServiceUnavailable struct {
	APIError
	ServiceName string `json:"service"`
}

// Type `ServiceUnavailable` represents a service timeout issue encountered during request processing
//
// Struct members:
//   - APIError: (inherited)
//   - Waited: the timeout duration that was exceeded
type ServiceTimedOut struct {
	APIError
	Waited time.Duration `json:"timeoutDuration"`
}
