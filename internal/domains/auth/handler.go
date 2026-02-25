package auth

/*
 * File: internal/domains/auth/handler.go
 *
 * Purpose: HTTP layer wrapping of authentication and authorization
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
	"github.com/tournabyte/webapi/internal/utils"
)

// Function `CreateUserHandler` creates a HTTP handler function compatible with the gin HTTP framework for creating user records
//
// Parameters:
//   - conn: the database connection this handler function will utilize to complete the write operations
//
// Returns:
//   - `gin.HandlerFunc`: closure capable of handling HTTP requests through integration with the HTTP gin framework
func CreateUserHandler(conn *utils.DatabaseConnection, signer jose.Signer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body NewUserRequest
		var newUser AuthenticatedUser
		slog.InfoContext(ctx, "Invoked handler function for user creation")

		if err := ctx.ShouldBindJSON(&body); err != nil {
			slog.Error("Could not bind request body", slog.String("error", err.Error()))
			ctx.Error(err)
			utils.RespondWithError(ctx, utils.ErrCouldNotBindRequestBody)
			return
		}

		if sess, err := conn.Client().StartSession(); err != nil {
			slog.Error("Could not start mongo session", slog.String("error", err.Error()))
			ctx.Error(err)
			utils.RespondWithError(ctx, utils.ErrUpstreamDataUnavailable)
			return
		} else {
			defer sess.EndSession(ctx.Request.Context())
			if err := CreateUserRecord(ctx, sess.Client(), signer, &body, &newUser); err != nil {
				slog.Error("Could not create account", slog.String("error", err.Error()))
				ctx.Error(err)
				utils.RespondWithError(ctx, utils.ErrUpstreamDataUnavailable)
				return
			} else {
				utils.RespondWithRequestedData(ctx, newUser, http.StatusCreated)
				slog.InfoContext(ctx, "Handler invocation for user creation completed successfully")
			}
		}
	}
}

// Function `CheckLoginHandler` creates a HTTP handler function compatible with the gin HTTP framework for verifying login attempts
//
// Parameters:
//   - db: the database connection this handler function will utilize to complete the read operations
//   - tokenSigner: the JWT signing tool following an internal algorithm and private key
//
// Returns:
//   - `gin.HandlerFunc`: closure capable of handling HTTP requests through integration with the HTTP gin framework
func CheckLoginHandler(db *utils.DatabaseConnection, tokenSigner jose.Signer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		slog.InfoContext(ctx, "Invoked handler function for user login verification")
		var body LoginAttempt
		var user AuthenticatedUser

		if err := ctx.ShouldBindJSON(&body); err != nil {
			slog.Error("Could not bind request body", slog.String("error", err.Error()))
			ctx.Error(err)
			utils.RespondWithError(ctx, utils.ErrCouldNotBindRequestBody)
			return
		}

		if sess, err := db.Client().StartSession(); err != nil {
			slog.Error("Could not start mongo session", slog.String("error", err.Error()))
			ctx.Error(err)
			utils.RespondWithError(ctx, utils.ErrUpstreamDataUnavailable)
			return
		} else {
			defer sess.EndSession(ctx.Request.Context())
			if err := ValidateLoginCredentials(ctx, sess.Client(), tokenSigner, &body, &user); err != nil {
				slog.Error("Could not validate credentials", slog.String("error", err.Error()))
				ctx.Error(err)
				utils.RespondWithError(ctx, utils.ErrUpstreamDataUnavailable)
			} else {
				utils.RespondWithRequestedData(ctx, user, http.StatusOK)
				slog.InfoContext(ctx, "Handler invocation for credential verification completed successfully")
			}
		}
	}
}
