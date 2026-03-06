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
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/go-playground/validator/v10"
	"github.com/tournabyte/webapi/internal/domains/user"
	"github.com/tournabyte/webapi/internal/utils"
)

// Function `CheckAuthorizationHeaderHandler` is a middleware handler that checks for and decodes the authorization header in the request
//
// Parameters:
//   - key: secret key for decoding access tokens
//   - issuer: expected value for the `iss` field
//   - subject: expected value for the `sub` field
//   - validate: validation dependency for verifying custom fields
//   - algorithms...: signing algorithms to try
//
// Returns:
//   - `gin.HandlerFunc`: closure capable of acting as a middleware function for a handlers chain
func CheckAuthorizationHeaderHandler(key string, issuer string, subject string, validate *validator.Validate, algorithms ...jose.SignatureAlgorithm) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		slog.InfoContext(ctx, "Invoked handler function for authorization check")
		var expectedHeaders AuthorizationHeaderContent
		var parsedToken jwt.JSONWebToken
		var publicClaims jwt.Claims = jwt.Claims{}
		var customClaims AuthorizationTokenClaims = AuthorizationTokenClaims{}

		if !utils.DidBindHeaders(ctx, &expectedHeaders) {
			slog.ErrorContext(ctx, "Failed to bind request headers")
			utils.RespondWithError(ctx, utils.ErrCouldNotBindRequestHeaders)
			return
		}

		slog.DebugContext(ctx, "Found authorization token", slog.String("raw", expectedHeaders.Token))
		if !parsedAuthorizationToken(ctx, expectedHeaders.Token, &parsedToken, algorithms...) {
			slog.ErrorContext(ctx, "Failed to parse signed token")
			utils.RespondWithError(ctx, utils.ErrInvalidAuthorizationToken)
			return
		}

		slog.DebugContext(ctx, "Decoding token claims")
		if !decodedTokenClaims(ctx, &parsedToken, []byte(key), &publicClaims, &customClaims) {
			slog.ErrorContext(ctx, "Failed to decode token claims")
			utils.RespondWithError(ctx, utils.ErrInvalidAuthorizationToken)
			return
		}

		slog.DebugContext(ctx, "Validating token claims", slog.String("iss", issuer), slog.String("sub", subject))
		if !validPublicClaims(ctx, &publicClaims, subject, issuer) {
			slog.ErrorContext(ctx, "Failed to validate public token claims")
			utils.RespondWithError(ctx, utils.ErrInvalidAuthorizationToken)
			return
		}

		if !validCustomClaims(ctx, &customClaims, validate) {
			slog.ErrorContext(ctx, "Failed to validate custom token claims")
			utils.RespondWithError(ctx, utils.ErrInvalidAuthorizationToken)
			return
		}

		ctx.Set(AuthorizationClaims, customClaims.Owner)
		slog.DebugContext(ctx, "Set authorization token owner ID in request context", slog.String("owner", customClaims.Owner))
		ctx.Next()
	}
}

// Function `UserRegistrationHandler` processes new user requests and instantiates them into the database
//
// Parameters:
//   - db: the database driver to use to insert user records
//   - signer: the JWT signing tool for signing access tokens
//   - sessionIssuer: access token issuers to include in token claims
//   - sessionSubject: access token subject to include in token claims
//
// Returns:
//   - `gin.HandlerFunc`: a closure capable of handling HTTP request through the gin framework
func UserRegistrationHandler(db *utils.DatabaseConnection, signer jose.Signer, sessionIssuer string, sessionSubject string) gin.HandlerFunc {
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
			Authenticate(ctx, body.Password, signer, sessionIssuer, sessionSubject, &login, &response),
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
//   - signer: the signing tool for creating JSON web tokens
//   - sessionIssuer: the access token issuer to include in token claims
//   - sessionSubject: the access token subject to include in token claims
//
// Returns:
//   - `gin.HandlerFunc`: a closure capable of handling HTTP request through the gin framework
func UserAuthenticationHandler(db *utils.DatabaseConnection, signer jose.Signer, sessionIssuer string, sessionSubject string) gin.HandlerFunc {
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
			Authenticate(ctx, body.Passphrase, signer, sessionIssuer, sessionSubject, &login, &response),
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
