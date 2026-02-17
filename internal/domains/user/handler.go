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
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/tournabyte/webapi/internal/utils"
)

// Function `CreateUserHandler` creates a HTTP handler function compatible with the gin HTTP framework for creating user records
//
// Parameters:
//   - conn: the database connection this handler function will utilize to complete the write operations
//
// Returns:
//   - `gin.HandlerFunc`: closure capable of handling HTTP requests through integration with the HTTP gin framework
func CreateUserHandler(conn *utils.DatabaseConnection) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body NewUser
		var newUserID string

		if err := ctx.ShouldBindJSON(&body); err != nil {
			slog.Error("Could not bind request body", slog.Any("expectedStructure", reflect.TypeOf(body)))
			panic(utils.ValidationFailed("validationFailure", err.Error()))
		}

		if sess, err := conn.Client().StartSession(); err != nil {
			panic(utils.TryAgainLater("database", "failed to start session"))
		} else {
			defer sess.EndSession(ctx.Request.Context())
			if err := CreateUserRecord(ctx, sess.Client(), &body, &newUserID); err != nil {
				panic(utils.TryAgainLater("accountCreationFailed", err.Error()))
			} else {
				ctx.JSON(
					http.StatusCreated,
					utils.RespondWithRequestedData(gin.H{"id": newUserID}),
				)
			}
		}
	}
}
