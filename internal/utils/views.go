package utils

/*
 * File: internal/utils/views.go
 *
 * Purpose: API layer helpers for API response structuring and construction
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Function `RespondWithRequestedData` produces a JSON mapping that indicates a successful response and sends it on the provided context
//
// Paramaters:
//   - ctx: the context to respond to
//   - data: the data to include in the `data` field of the JSON response body
//   - code: the status code to use with the response
//
// Encoding:
//
//	{
//		"ok": true,
//		"data": ...
//	}
func RespondWithRequestedData(ctx *gin.Context, data any, code int) {
	var body gin.H = gin.H{
		"ok":   true,
		"data": data,
	}
	ctx.JSON(code, body)

}

// Function `RespondWithError` produces a JSON mapping that indicates an unsuccessful response and sends in on the provided context
//
// Paramaters:
//   - ctx: the context to respond to
//   - err: the message to include in the `error.message` field of the JSON response body
//
// Encoding:
//
//	{
//		"ok": false,
//		"error": {
//			"message": "...",
//			"details": "{...}"
//		}
//	}
func RespondWithError(ctx *gin.Context, err error) {
	var body gin.H = gin.H{
		"ok": false,
		"error": gin.H{
			"message": err.Error(),
			"details": ctx.Errors.JSON(),
		},
	}
	if failure, exists := errors.AsType[HandlerFuncFailure](err); exists {
		ctx.AbortWithStatusJSON(failure.StatusCode, body)
	} else {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, body)
	}
}
