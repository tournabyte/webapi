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
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

const AuthorizationClaims = "UserAuthorization"

// Type `AuthenticationTokenClaims` represents the custom claims present in a JWT produced by the Tournabyte API
//
// Fields:
//   - Owner: private claim expected to be the userID of the account this token was issued to
type AuthenticationTokenClaims struct {
	Owner string `json:"owner"`
}

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

// Function `VerifyAuthorization` provides a middleware that verifies the validity of the authorization header on incoming requests
//
// Returns:
//   - `gin.HandlerFunc`: a HTTP middleware closure that is compatible with the Gin HTTP framework
func VerifyAuthorization(key []byte, signingAlgorithms ...jose.SignatureAlgorithm) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if raw := ctx.GetHeader(http.CanonicalHeaderKey("Authorization")); len(raw) == 0 {
			slog.Error("Missing authorization header when it was expected")
			ctx.AbortWithStatusJSON(http.StatusBadRequest, RespondWithError(NotAuthorized(), nil))
		} else {
			token, err := jwt.ParseSigned(raw, signingAlgorithms)
			if err != nil {
				slog.Error("Failed to parse signed token", slog.String("err", err.Error()))
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, RespondWithError(NotAuthorized(), nil))
				return
			}
			slog.Debug("Token parsed", slog.Any("token", token))
			public := jwt.Claims{}
			custom := AuthenticationTokenClaims{}

			if err := token.Claims(key, &public, &custom); err != nil {
				slog.Error("Could not decode token claims", slog.String("err", err.Error()))
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, RespondWithError(NotAuthorized(), nil))
				return
			}

			expectedPublic := jwt.Expected{
				Issuer:  "api.tournabyte.com",
				Subject: "Tournabyte API authorization",
			}
			if err := public.Validate(expectedPublic); err != nil {
				slog.Error("Public claims validation failed", slog.String("err", err.Error()))
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, RespondWithError(NotAuthorized(), nil))
				return
			}

			ctx.Set(AuthorizationClaims, custom.Owner)
			ctx.Next()
		}
	}
}
