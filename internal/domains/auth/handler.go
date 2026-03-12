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
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tournabyte/webapi/internal/utils"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func UserRegistrationHandler(tokenOpts *TokenOptions, sessionOpts *SessionOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		slog.Info("Invoked handler for user registration")

		ctx, cancelf := context.WithCancelCause(c.Request.Context())
		defer cancelf(nil)
		state := AuthenticationHandlerState{
			bindings:  utils.BindingsFromRequestContext(c, utils.ShouldHaveJSONBody),
			mongosess: mongo.SessionFromContext(ctx),
		}
		slog.Debug("Initialized handler state and context variables")

		slog.Debug("Feeding pipeline...")
		out1 := utils.Stage(ctx, cancelf, bindAuthenticationRequest, utils.Feed(ctx, state))
		out2 := utils.Stage(ctx, cancelf, deriveAccountRecordFromRequest, out1)
		out3 := utils.Stage(ctx, cancelf, saveAccountRecord, out2)

		slog.Debug("Awaiting context cancellation or pipeline conclusion")
		select {
		case <-ctx.Done():
			slog.Debug("Context cancelled before deferred call")
			if err := context.Cause(ctx); err != nil {
				slog.Error("Cancellation has attached cause", slog.String("error", err.Error()))
				c.Error(err)
				utils.RespondWithError(c, err)
			}
		case res, ok := <-out3:
			if !ok {
				slog.Error("Handler state could not be read")
				c.Error(utils.ErrUndisclosedHandlerFailure)
				utils.RespondWithError(c, utils.ErrUndisclosedHandlerFailure)
			} else {
				slog.Info("Handler state received from pipeline")
				utils.RespondWithRequestedData(c, res.response, http.StatusCreated)
			}
		}
	}
}

// func UserAuthenticationHandler()
// func SessionRefreshHandler()
// func AuthenticationUpdateHandler()
// func SessionRevocationHandler()
