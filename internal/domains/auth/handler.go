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
	"github.com/tournabyte/webapi/internal/domains/user"
	"github.com/tournabyte/webapi/internal/utils"
)

// Function `UserRegistrationHandler` processes new user requests and instantiates them into the database
//
// Parameters:
//   - db: the database driver to use to insert user records
//
// Returns:
//   - `gin.HandlerFunc`: a closure capable of handling HTTP request through the gin framework
func UserRegistrationHandler(db *utils.DatabaseConnection, signer jose.Signer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		slog.InfoContext(ctx, "Invoked handler function for user registration")
		var body NewUserRequest
		var login user.LoginCredentials
		var response AuthenticatedUser

		if !utils.DidBindBody(ctx, &body) {
			slog.ErrorContext(ctx, "Failed to bind request body")
			utils.RespondWithError(ctx, utils.ErrCouldNotBindRequestBody)
			return
		}

		ops := []utils.MongoOperationFunc{
			Register(ctx, &body),
			FindLoginDetails(ctx, body.Email, &login, &response),
			Authenticate(ctx, body.Password, signer, &login, &response),
		}
		if !db.SessionCompleted(ctx, ops...) {
			slog.ErrorContext(ctx, "Failed to create user record")
			utils.RespondWithError(ctx, utils.ErrUpstreamDataUnavailable)
			return
		}

		slog.InfoContext(ctx, "User registration completed successfully")
		utils.RespondWithRequestedData(ctx, &response, http.StatusCreated)
	}
}

// Function `UserAuthenticationHandler` processes login attempts and creates the user session if the attempt is valid
//
// Parameters:
//   - db: the database driver used to find and verify login credentials
//
// Returns:
//   - `gin.HandlerFunc`: a closure capable of handling HTTP request through the gin framework
func UserAuthenticationHandler(db *utils.DatabaseConnection, signer jose.Signer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		slog.InfoContext(ctx, "Invoked handler function for user authentication")
		var body LoginAttempt
		var login user.LoginCredentials
		var response AuthenticatedUser

		if !utils.DidBindBody(ctx, &body) {
			slog.ErrorContext(ctx, "Failed to bind request body")
			utils.RespondWithError(ctx, utils.ErrCouldNotBindRequestBody)
			return
		}

		ops := []utils.MongoOperationFunc{
			FindLoginDetails(ctx, body.AuthenticateAs, &login, &response),
			Authenticate(ctx, body.Passphrase, signer, &login, &response),
		}
		if !db.SessionCompleted(ctx, ops...) {
			slog.ErrorContext(ctx, "Failed to create session")
			utils.RespondWithError(ctx, utils.ErrInvalidLoginAttempt)
			return
		}

		slog.InfoContext(ctx, "User authentication completed successfully")
		utils.RespondWithRequestedData(ctx, &response, http.StatusOK)
	}
}
