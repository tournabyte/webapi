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
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStructuredErrorCreation(t *testing.T) {

	t.Run("ValidationFailed", func(t *testing.T) {
		err := ValidationFailed()
		assert.Equal(t, http.StatusBadRequest, err.StatusCode())
	})

	t.Run("ValidationFailedWithDetails", func(t *testing.T) {
		err := ValidationFailed("some field", "is required", "another field", "must be a number")
		assert.Equal(t, http.StatusBadRequest, err.StatusCode())
		assert.Equal(t, 2, len(err.Details))
	})

	t.Run("NotAuthorized", func(t *testing.T) {
		err := NotAuthorized()
		assert.Equal(t, http.StatusUnauthorized, err.StatusCode())
	})

	t.Run("NotAuthorizedWithDetails", func(t *testing.T) {
		err := NotAuthorized("authorization token", "is expired")
		assert.Equal(t, http.StatusUnauthorized, err.StatusCode())
		assert.Equal(t, 1, len(err.Details))
	})

	t.Run("AccessDenied", func(t *testing.T) {
		err := AccessDenied()
		assert.Equal(t, http.StatusForbidden, err.StatusCode())
	})

	t.Run("AccessDeniedWithDetails", func(t *testing.T) {
		err := AccessDenied("permissions", "rw", "granted", "no")
		assert.Equal(t, http.StatusForbidden, err.StatusCode())
		assert.Equal(t, 2, len(err.Details))
	})

	t.Run("NoSuchResource", func(t *testing.T) {
		err := ResourceNotFound()
		assert.Equal(t, http.StatusNotFound, err.StatusCode())
	})

	t.Run("NoSuchResourceWithDetails", func(t *testing.T) {
		err := ResourceNotFound("userid", "no user with id 1")
		assert.Equal(t, http.StatusNotFound, err.StatusCode())
		assert.Equal(t, 1, len(err.Details))
	})

	t.Run("TooManyRequests", func(t *testing.T) {
		err := SlowDown()
		assert.Equal(t, http.StatusTooManyRequests, err.StatusCode())
	})

	t.Run("TooManyRequestsWithDetails", func(t *testing.T) {
		err := SlowDown("tryAgainAt", time.Now().UTC().Add(time.Hour).String())
		assert.Equal(t, http.StatusTooManyRequests, err.StatusCode())
		assert.Equal(t, 1, len(err.Details))
	})

	t.Run("ServiceUnavailable", func(t *testing.T) {
		err := TryAgainLater()
		assert.Equal(t, http.StatusServiceUnavailable, err.StatusCode())
	})

	t.Run("ServiceUnavailableWithDetails", func(t *testing.T) {
		err := TryAgainLater("tryAgainAt", time.Now().UTC().Add(time.Hour).String())
		assert.Equal(t, http.StatusServiceUnavailable, err.StatusCode())
		assert.Equal(t, 1, len(err.Details))
	})

	t.Run("UnstructuredError", func(t *testing.T) {
		err := ErrAPIRequestFailed{
			CausedBy: errors.New("custom error"),
		}
		assert.Equal(t, http.StatusInternalServerError, err.StatusCode())
	})
}

func TestErrorInterface(t *testing.T) {
	// Test Error() method implementation
	err := ValidationFailed("field", "name")
	assert.Equal(t, ErrValidationFailed.Error(), err.Error())

	// Test Unwrap() method implementation
	unwrapped := err.Unwrap()
	assert.Equal(t, ErrValidationFailed, unwrapped)
}
