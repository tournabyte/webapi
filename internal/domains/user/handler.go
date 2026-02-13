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
	"context"

	"github.com/gin-gonic/gin"
	"github.com/tournabyte/webapi/internal/utils"
)

func CreateUserHandler(conn *utils.DatabaseConnection) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body NewUser
		var newUserID string

		if err := ctx.ShouldBindJSON(&body); err != nil {
			panic(utils.ValidationFailed("validationFailure", err.Error()))
		}

		if sess, err := conn.Client().StartSession(); err != nil {
			panic(utils.TryAgainLater("database", "failed to start session"))
		} else {
			defer sess.EndSession(ctx.Request.Context())
			sess.WithTransaction(
				ctx.Request.Context(),
				func(ctx context.Context) (any, error) {
					if err := CreateUserRecord(ctx, sess.Client(), &body, &newUserID); err != nil {
						return false, err
					}
					return true, nil
				},
			)
		}

	}
}
