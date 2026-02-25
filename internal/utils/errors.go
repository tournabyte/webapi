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
	"net/http"
)

// Type `HandlerFuncFailure` represents an error from a Tournabyte API handler function
//
// Fields:
//   - StatusCode: the HTTP status code to mark the response status (typically a 4xx or 5xx code)
//   - ErrorMsg: the contextual information about the error
type HandlerFuncFailure struct {
	StatusCode int
	ErrorMsg   string
}

// Function `NewHandlerFuncFailure` creates an handler func failure instance
//
// Parameters:
//   - code: the HTTP status code for the handler func failure instance
//   - msg: the contextual information about the error for the handler func failure
//
// Returns:
//   - `HandlerFuncFailure`: the failure instance representing a particular issue occuring within a handler function
func NewHandlerFuncFailure(code int, msg string) HandlerFuncFailure {
	return HandlerFuncFailure{
		StatusCode: code,
		ErrorMsg:   msg,
	}
}

// Function `(HandlerFuncFailure).Error` implements the error interface for the `HandlerFuncFailure` type
//
// Returns:
//   - `string`: the stringified error
func (f HandlerFuncFailure) Error() string {
	return f.ErrorMsg
}

// Constant error messages to display when handler functions encouter a failure
const (
	failedToBindBodyMsg       = "request body malformed"
	failedToBindURIMsg        = "request URI malformed"
	failedLoginMsg            = "invalid login credentials presented"
	failedAuthVerificationMsg = "invalid authorization presented"
	failedToReachUpstreamData = "upstream data service not available"

	failedHandlerGenericMsg = "try again later"
)

// Predefined handler function failure instances for convenience
var (
	ErrCouldNotBindRequestBody       = NewHandlerFuncFailure(http.StatusBadRequest, failedToBindBodyMsg)
	ErrCouldNotBindRequestParameters = NewHandlerFuncFailure(http.StatusBadRequest, failedToBindURIMsg)
	ErrInvalidLoginAttempt           = NewHandlerFuncFailure(http.StatusUnauthorized, failedLoginMsg)
	ErrInvalidAuthorizationToken     = NewHandlerFuncFailure(http.StatusForbidden, failedAuthVerificationMsg)
	ErrUpstreamDataUnavailable       = NewHandlerFuncFailure(http.StatusBadGateway, failedToReachUpstreamData)

	ErrUndisclosedHandlerFailure = NewHandlerFuncFailure(http.StatusInternalServerError, failedHandlerGenericMsg)
)
