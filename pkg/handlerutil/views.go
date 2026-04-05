package handlerutil

/*
 * File: pkg/handlerutil/views.go
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
//		"data": {...}
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
//   - err: the information to include in the `error` field of the JSON response body
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
//
//	!!! PANICS !!!
//		- When the received error does not contain a `handlerFailure` error within its `Unwrap` tree
//		- In practice, this panic should be handled by a recovery middleware
func RespondWithError(ctx *gin.Context, err error) {
	if failure, exists := errors.AsType[handlerFailure](err); exists {
		var body = gin.H{
			"ok": false,
			"error": gin.H{
				"message": failure.Message,
				"details": failure.DetailMapping(),
			},
		}
		ctx.AbortWithStatusJSON(failure.statusCode, body)
		ctx.Error(err)
	} else {
		panic(err)
	}
}
