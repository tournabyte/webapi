package utils

/*
 * File: internal/utils/errors_test.go
 *
 * Purpose: unit test for errors emerging from service providers
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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationFailedConstructionAndJSON(t *testing.T) {
	// Test construction
	err := ValidationFailed("field", "name", "reason", "mandatory field")
	assert.Equal(t, ErrValidationFailed, err.CausedBy)
	assert.NotNil(t, err.Details)
	assert.Equal(t, "name", err.Details["field"])
	assert.Equal(t, "mandatory field", err.Details["reason"])

	// Test JSON marshalling
	jsonData, errMarshal := json.Marshal(&err)
	require.NoError(t, errMarshal, "Failed to marshal ValidationFailed")

	var result map[string]interface{}
	errUnmarshal := json.Unmarshal(jsonData, &result)
	require.NoError(t, errUnmarshal, "Failed to unmarshal JSON")

	assert.Equal(t, ErrValidationFailed.Error(), result["message"])
	assert.NotNil(t, result["details"])
	details := result["details"].(map[string]interface{})
	assert.Equal(t, "name", details["field"])
	assert.Equal(t, "mandatory field", details["reason"])
}

func TestNotAuthorizedConstructionAndJSON(t *testing.T) {
	// Test construction
	err := NotAuthorized("resource", "user_id")
	assert.Equal(t, ErrNotAuthorized, err.CausedBy)
	assert.NotNil(t, err.Details)
	assert.Equal(t, "user_id", err.Details["resource"])

	// Test JSON marshalling
	jsonData, errMarshal := json.Marshal(&err)
	require.NoError(t, errMarshal, "Failed to marshal NotAuthorized")

	var result map[string]interface{}
	errUnmarshal := json.Unmarshal(jsonData, &result)
	require.NoError(t, errUnmarshal, "Failed to unmarshal JSON")

	assert.Equal(t, ErrNotAuthorized.Error(), result["message"])
	assert.NotNil(t, result["details"])
	details := result["details"].(map[string]interface{})
	assert.Equal(t, "user_id", details["resource"])
}

func TestAccessDeniedConstructionAndJSON(t *testing.T) {
	// Test construction
	err := AccessDenied("user", "admin", "reason", "insufficient privilege")
	assert.Equal(t, ErrAccessDenied, err.CausedBy)
	assert.NotNil(t, err.Details)
	assert.Equal(t, "admin", err.Details["user"])
	assert.Equal(t, "insufficient privilege", err.Details["reason"])

	// Test JSON marshalling
	jsonData, errMarshal := json.Marshal(&err)
	require.NoError(t, errMarshal, "Failed to marshal AccessDenied")

	var result map[string]interface{}
	errUnmarshal := json.Unmarshal(jsonData, &result)
	require.NoError(t, errUnmarshal, "Failed to unmarshal JSON")

	assert.Equal(t, ErrAccessDenied.Error(), result["message"])
	assert.NotNil(t, result["details"])
	details := result["details"].(map[string]interface{})
	assert.Equal(t, "admin", details["user"])
	assert.Equal(t, "insufficient privilege", details["reason"])
}

func TestResourceNotFoundConstructionAndJSON(t *testing.T) {
	// Test construction
	err := ResourceNotFound("resource", "user_id", "reason", "not found")
	assert.Equal(t, ErrNoSuchResource, err.CausedBy)
	assert.NotNil(t, err.Details)
	assert.Equal(t, "user_id", err.Details["resource"])
	assert.Equal(t, "not found", err.Details["reason"])

	// Test JSON marshalling
	jsonData, errMarshal := json.Marshal(&err)
	require.NoError(t, errMarshal, "Failed to marshal ResourceNotFound")

	var result map[string]interface{}
	errUnmarshal := json.Unmarshal(jsonData, &result)
	require.NoError(t, errUnmarshal, "Failed to unmarshal JSON")

	assert.Equal(t, ErrNoSuchResource.Error(), result["message"])
	assert.NotNil(t, result["details"])
	details := result["details"].(map[string]interface{})
	assert.Equal(t, "user_id", details["resource"])
	assert.Equal(t, "not found", details["reason"])
}

func TestSlowDownConstructionAndJSON(t *testing.T) {
	// Test construction
	err := SlowDown("rate", "limit", "reason", "exceeded")
	assert.Equal(t, ErrRateLimitExceeded, err.CausedBy)
	assert.NotNil(t, err.Details)
	assert.Equal(t, "limit", err.Details["rate"])
	assert.Equal(t, "exceeded", err.Details["reason"])

	// Test JSON marshalling
	jsonData, errMarshal := json.Marshal(&err)
	require.NoError(t, errMarshal, "Failed to marshal SlowDown")

	var result map[string]interface{}
	errUnmarshal := json.Unmarshal(jsonData, &result)
	require.NoError(t, errUnmarshal, "Failed to unmarshal JSON")

	assert.Equal(t, ErrRateLimitExceeded.Error(), result["message"])
	assert.NotNil(t, result["details"])
	details := result["details"].(map[string]interface{})
	assert.Equal(t, "limit", details["rate"])
	assert.Equal(t, "exceeded", details["reason"])
}

func TestTryAgainLaterConstructionAndJSON(t *testing.T) {
	// Test construction
	err := TryAgainLater("service", "down")
	assert.Equal(t, ErrServiceUnavailable, err.CausedBy)
	assert.NotNil(t, err.Details)
	assert.Equal(t, "down", err.Details["service"])

	// Test JSON marshalling
	jsonData, errMarshal := json.Marshal(&err)
	require.NoError(t, errMarshal, "Failed to marshal TryAgainLater")

	var result map[string]interface{}
	errUnmarshal := json.Unmarshal(jsonData, &result)
	require.NoError(t, errUnmarshal, "Failed to unmarshal JSON")

	assert.Equal(t, ErrServiceUnavailable.Error(), result["message"])
	assert.NotNil(t, result["details"])
	details := result["details"].(map[string]interface{})
	assert.Equal(t, "down", details["service"])
}

func TestErrorStatusCode(t *testing.T) {
	// Test ValidationFailed status code
	err := ValidationFailed()
	statusCode := err.StatusCode()
	assert.Equal(t, http.StatusBadRequest, statusCode)

	// Test NotAuthorized status code
	err = NotAuthorized()
	statusCode = err.StatusCode()
	assert.Equal(t, http.StatusUnauthorized, statusCode)

	// Test AccessDenied status code
	err = AccessDenied()
	statusCode = err.StatusCode()
	assert.Equal(t, http.StatusForbidden, statusCode)

	// Test ResourceNotFound status code
	err = ResourceNotFound()
	statusCode = err.StatusCode()
	assert.Equal(t, http.StatusNotFound, statusCode)

	// Test SlowDown status code
	err = SlowDown()
	statusCode = err.StatusCode()
	assert.Equal(t, http.StatusTooManyRequests, statusCode)

	// Test TryAgainLater status code
	err = TryAgainLater()
	statusCode = err.StatusCode()
	assert.Equal(t, http.StatusServiceUnavailable, statusCode)

	// Test default case (should return 500)
	customErr := ErrAPIRequestFailed{
		CausedBy: errors.New("custom error"),
		Details:  map[string]string{"type": "custom"},
	}
	statusCode = customErr.StatusCode()
	assert.Equal(t, http.StatusInternalServerError, statusCode)
}

func TestErrorInterface(t *testing.T) {
	// Test Error() method implementation
	err := ValidationFailed("field", "name")
	assert.Equal(t, ErrValidationFailed.Error(), err.Error())

	// Test Unwrap() method implementation
	unwrapped := err.Unwrap()
	assert.Equal(t, ErrValidationFailed, unwrapped)
}
