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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerFunctionFailureCreation(t *testing.T) {
	code := 100
	msg := "an error message"
	err := NewHandlerFuncFailure(code, msg)

	assert.Equal(t, code, err.StatusCode)
	assert.Equal(t, msg, err.ErrorMsg)
}

func TestPredefinedHandlerFunctionFailureInstances(t *testing.T) {
	t.Run("Request body", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, ErrCouldNotBindRequestBody.StatusCode)
		assert.Equal(t, failedToBindBodyMsg, ErrCouldNotBindRequestBody.ErrorMsg)
	})

	t.Run("Request parameters", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, ErrCouldNotBindRequestParameters.StatusCode)
		assert.Equal(t, failedToBindURIMsg, ErrCouldNotBindRequestParameters.ErrorMsg)
	})

	t.Run("Failed login", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, ErrInvalidLoginAttempt.StatusCode)
		assert.Equal(t, failedLoginMsg, ErrInvalidLoginAttempt.ErrorMsg)
	})

	t.Run("Request body", func(t *testing.T) {
		assert.Equal(t, http.StatusForbidden, ErrInvalidAuthorizationToken.StatusCode)
		assert.Equal(t, failedAuthVerificationMsg, ErrInvalidAuthorizationToken.ErrorMsg)
	})

	t.Run("Generic handler failure", func(t *testing.T) {
		assert.Equal(t, http.StatusInternalServerError, ErrUndisclosedHandlerFailure.StatusCode)
		assert.Equal(t, failedHandlerGenericMsg, ErrUndisclosedHandlerFailure.ErrorMsg)
	})
}
