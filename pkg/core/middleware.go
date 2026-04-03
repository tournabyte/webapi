package core

/*
 * File: pkg/core/middleware.go
 *
 * Purpose: definition of the server for the Tournabyte webapi
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tournabyte/webapi/pkg/handlerutil"
)

// Function `(*tournabyteAPIService).recoverPanicAsFailure` creates a custom recovery middleware for the underlying gin HTTP server
//
// Parameters:
//   - ctx: the request context to recover from
//   - e: the value (typically an error) causing the panic
func (srv *tournabyteAPIService) recoverPanicAsFailure(ctx *gin.Context, e any) {
	if e == nil {
		return
	}

	switch e := e.(type) {
	case error:
		srv.logger.Error("Recovering from panic from error", slog.String("err", e.Error()))
		handlerutil.RespondWithError(
			ctx,
			handlerutil.ErrInternalServerError(
				handlerutil.NewDetail("recover", "something went wrong..."),
			),
		)
	default:
		srv.logger.Error("Recovering from panic from non-error", slog.Any("panic_value", e))
		handlerutil.RespondWithError(
			ctx,
			handlerutil.ErrInternalServerError(
				handlerutil.NewDetail("recover", "this definitely should not have happened..."),
			),
		)
	}
}

// Function `(*tournabyteAPIService).assignRequestIdentifier` creates unique IDs for an incoming request
//
// Parameters:
//   - ctx: the request context to assign a unique request ID to
func (srv *tournabyteAPIService) assignRequestIdentifier(ctx *gin.Context) {
	requestID := uuid.New().String()
	ctx.Set("RequestID", requestID)
	ctx.Writer.Header().Set("X-Request-ID", requestID)
	srv.logger.Debug("Assigning identifier to request", slog.String("RequestID", requestID))
	ctx.Next()
}

// Function `(*tournabyteAPIService).markRequestStartTimestamp` marks the timestamp request processing begins
//
// Parameters:
//   - ctx: the request context to begin timing
func (srv *tournabyteAPIService) markRequestStartTimestamp(ctx *gin.Context) {
	startTime := time.Now().UTC()
	ctx.Writer.Header().Set("X-Request-Received-At", startTime.String())
	ctx.Set("RequestTime", startTime)
	srv.logger.Debug("Marked the timestamp request began processing", slog.Time("t", startTime))
	ctx.Next()
}

// Function `(*tournabyteAPIService).withMongoSession` sets up a mongo session within the given request context
//
// Parameters:
//   - ctx: the context that requires a mongo session
func (srv *tournabyteAPIService) withMongoSession(ctx *gin.Context) {
	if sessCtx, err := srv.db.SetUpSession(ctx.Request.Context()); err != nil {
		srv.logger.Error("Failed to start session", slog.String("error", err.Error()))
		handlerutil.RespondWithError(ctx, err)
	} else {
		defer srv.db.TearDownSession(sessCtx)
		ctx.Request = ctx.Request.WithContext(sessCtx)
		ctx.Next()
	}
}

// Function `(*tournabyteAPIService).withMongoTransaction` sets up a mongo transaction within the given request context
//
// Parameters:
//   - ctx: the context requiring a mongo transaction
//
// Warnings:
//   - the context to start a transaction should already have a session intializes within the context. It is illegal to start a transaction without a session
//   - in a handlers chain, withMongoTransaction should be ordered after withMongoSession, i.e. [..., withMongoSession, withMongoTransaction, ...]
func (srv *tournabyteAPIService) withMongoTransaction(ctx *gin.Context) {
	if err := srv.db.BeginTransaction(ctx.Request.Context()); err != nil {
		srv.logger.Error("Transaction failed to start", slog.String("error", err.Error()))
		handlerutil.RespondWithError(ctx, err)
	} else {
		ctx.Next()

		if len(ctx.Errors) > 0 {
			srv.logger.Debug("Transaction aborted")
			srv.db.AbortTransaction(ctx.Request.Context())
		} else {
			srv.logger.Debug("Transaction committed")
			srv.db.CommitTransaction(ctx.Request.Context())
		}
	}
}
