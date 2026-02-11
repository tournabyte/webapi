package utils

/*
 * File: internal/utils/middleware.go
 *
 * Purpose: HTTP layer middleware handlers for API endpoints
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Function `ErrorRecovery` provides a recovery middleware that delivers appropriate HTTP response status codes based on the error being recovered from
//
// Returns:
//   - `gin.HandlerFunc`: a HTTP middleware closure that is compatible with the Gin HTTP framework
func ErrorRecovery() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			err := recover()
			if err != nil {
				switch err := err.(type) {
				case ErrAPIRequestFailed:
					slog.Error("Recovering from structured API error", slog.String("err", err.Error()))
					ctx.AbortWithStatusJSON(err.StatusCode(), RespondWithError(err.CausedBy, &err.Details))
				case error:
					slog.Error("Recovering from unstructured API error", slog.String("err", err.Error()))
					ctx.AbortWithStatusJSON(http.StatusInternalServerError, RespondWithError(err, nil))
				default:
					slog.Error("Recovering from incoming unknown error", slog.Any("err", err))
					ctx.AbortWithStatusJSON(http.StatusInternalServerError, RespondWithError(TryAgainLater(), nil))
				}
			}
		}()
		ctx.Next()
	}
}
