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
	"github.com/go-playground/validator/v10"
)

const AuthorizationClaims = "UserAuthorization"

// Type `AuthenticationTokenClaims` represents the custom claims present in a JWT produced by the Tournabyte API
//
// Fields:
//   - Owner: private claim expected to be the userID of the account this token was issued to
type AuthenticationTokenClaims struct {
	Owner string `json:"owner" validate:"required,mongodb"`
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
				case error:
					slog.Error("Recovering from error panic", slog.String("error", err.Error()))
					RespondWithError(ctx, err)
				default:
					slog.Error("Recovering from non error panic")
					RespondWithError(ctx, ErrUndisclosedHandlerFailure)
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
			RespondWithError(ctx, ErrInvalidAuthorizationToken)
		} else {
			slog.Debug("Found authorization token", slog.String("raw", raw))
			token, err := jwt.ParseSigned(raw, signingAlgorithms)
			if err != nil {
				slog.Error("Failed to parse signed token", slog.String("err", err.Error()))
				ctx.Error(err)
				RespondWithError(ctx, ErrInvalidAuthorizationToken)
				return
			}
			public := jwt.Claims{}
			custom := AuthenticationTokenClaims{}

			if err := token.Claims(key, &custom, &public); err != nil {
				slog.Error("Could not decode token claims", slog.String("err", err.Error()))
				ctx.Error(err)
				RespondWithError(ctx, ErrInvalidAuthorizationToken)
				return
			}

			expectedPublic := jwt.Expected{
				Issuer:  "api.tournabyte.com",
				Subject: "Tournabyte API authorization",
			}
			if err := public.Validate(expectedPublic); err != nil {
				slog.Error("Public claims validation failed", slog.String("err", err.Error()))
				ctx.Error(err)
				RespondWithError(ctx, ErrInvalidAuthorizationToken)
				return
			}
			if err := validator.New().Struct(custom); err != nil {
				slog.Error("Custom claims validation failed", slog.String("err", err.Error()))
				ctx.Error(err)
				RespondWithError(ctx, ErrInvalidAuthorizationToken)
				return
			}

			ctx.Set(AuthorizationClaims, custom.Owner)
			slog.DebugContext(ctx, "Set authorization token owner ID in request context", slog.String("owner", custom.Owner))
			ctx.Next()
		}
	}
}
