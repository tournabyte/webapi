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

// Function `ErrorFallback` provides a recovery middleware that delivers appropriate HTTP response status codes based on the error being recovered from
//
// Returns:
//   - `gin.HandlerFunc`: a HTTP middleware closure that is compatible with the Gin HTTP framework
func ErrorFallback() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			err := recover()
			if err != nil {
				switch err := err.(type) {
				case ValidationError:
					slog.Error("Recovering from incoming validation error")
					ctx.JSON(http.StatusBadRequest, WriteErrorResponse(err))
				case AuthorizationError:
					slog.Error("Recovering from incoming authorization error")
					if err.Reconcile {
						ctx.JSON(http.StatusUnauthorized, WriteErrorResponse(err))
					} else {
						ctx.JSON(http.StatusForbidden, WriteErrorResponse(err))
					}
				case ServiceUnavailable:
					slog.Error("Recovering from incoming service availability error")
					ctx.JSON(http.StatusServiceUnavailable, WriteErrorResponse(err))
				case ServiceTimedOut:
					slog.Error("Recovering from a timeout exceeded error")
					ctx.JSON(http.StatusGatewayTimeout, WriteErrorResponse(err))
				default:
					slog.Error("Recovering from incoming unknown error", slog.Any("err", err))
					ctx.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": gin.H{"message": err}})
				}
			}
		}()
		ctx.Next()
	}
}
