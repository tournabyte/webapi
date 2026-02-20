package user

/*
 * File: internal/domains/user/handler.go
 *
 * Purpose: HTTP layer wrapping of user management
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

		if err := ctx.ShouldBindJSON(&body); err != nil {
			slog.Error("Could not bind request body", slog.String("error", err.Error()))
			panic(utils.ValidationFailed("validationFailure", err.Error()))
		}

		if sess, err := conn.Client().StartSession(); err != nil {
			panic(utils.TryAgainLater("database", "failed to start session"))
		} else {
			defer sess.EndSession(ctx.Request.Context())
			if err := CreateUserRecord(ctx, sess.Client(), signer, &body, &newUser); err != nil {
				panic(utils.TryAgainLater("accountCreationFailed", err.Error()))
			} else {
				ctx.JSON(
					http.StatusCreated,
					utils.RespondWithRequestedData(newUser),
				)
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
		var body LoginAttempt
		var user AuthenticatedUser

		if err := ctx.ShouldBindJSON(&body); err != nil {
			panic(utils.NotAuthorized())
		}

		if sess, err := db.Client().StartSession(); err != nil {
			panic(utils.TryAgainLater("database", "failed to start session"))
		} else {
			defer sess.EndSession(ctx.Request.Context())
			if err := ValidateLoginCredentials(ctx, sess.Client(), tokenSigner, &body, &user); err != nil {
				panic(utils.NotAuthorized("credentials", "wrong email or password"))
			} else {
				ctx.JSON(
					http.StatusOK,
					utils.RespondWithRequestedData(user),
				)
			}
		}
	}
}

//func UpdateLoginHandler(db *utils.DatabaseConnection) gin.HandlerFunc

//func SessionRefreshHandler(db *utils.DatabaseConnection) gin.HandlerFunc

//func SessionRevokeHandler(db *utils.DatabaseConnection) gin.HandlerFunc

//func CreateProfileHandler(db *utils.DatabaseConnection) gin.HandlerFunc

//func FindProfileHandler(db *utils.DatabaseConnection, s3 *utils.MinioConnection) gin.HandlerFunc

//func UpdateProfileHandler(db *utils.DatabaseConnection, s3 *utils.MinioConnection) gin.HandlerFunc

// func UpdateProfileAvatarHandler(db *utils.DatabaseConnection, s3 *utils.MinioConnection) gin.HandlerFunc

// fjnc UpdateProfilePreferencesHandler(db *utils.DatabaseConnection) gin.HandlerFunc
