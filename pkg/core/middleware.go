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
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
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
		log.Printf("[RECOVER] panic from error: %s", e.Error())
		handlerutil.RespondWithError(
			ctx,
			handlerutil.ErrInternalServerError(
				handlerutil.NewDetail("recover", "something went wrong..."),
			),
		)
	default:
		log.Printf("[RECOVER] panic caused by value: %v", e)
		handlerutil.RespondWithError(
			ctx,
			handlerutil.ErrInternalServerError(
				handlerutil.NewDetail("recover", "this definitely should not have happened..."),
			),
		)
	}
}

// Function `(*tournabyteAPIService).serviceLoggerFmt` is a gin.LogFormatter that defines a custom log format for server level logs
//
// Parameters:
//   - param: the gin specific logging parameters
//
// Returns:
//   - `string`: a format specified string indicating what gin server logs should look like
func (srv *tournabyteAPIService) serviceLoggerFmt(param gin.LogFormatterParams) string {
	return fmt.Sprintf(
		"[GIN] %s <%s | %s %s | %s> -  <%d | took: %s>\n",
		param.TimeStamp.Format(time.RFC1123Z),
		param.ClientIP,
		param.Method,
		param.Path,
		param.Request.Proto,
		param.StatusCode,
		param.Latency.String(),
	)
}

// Function `(*tournabyteAPIService).withMongoSession` sets up a mongo session within the given request context
//
// Parameters:
//   - ctx: the context that requires a mongo session
func (srv *tournabyteAPIService) withMongoSession(ctx *gin.Context) {
	if sessCtx, err := srv.db.SetUpSession(ctx.Request.Context()); err != nil {
		log.Printf("[MIDDLEWARE]: error starting mongo session: %s", err.Error())
		handlerutil.RespondWithError(ctx, err)
	} else {
		defer log.Printf("[MIDDLEWARE]: mongo session closed")
		defer srv.db.TearDownSession(sessCtx)
		ctx.Request = ctx.Request.WithContext(sessCtx)
		log.Printf("[MIDDLEWARE]: mongo session injected into request context (session ending deferred)")
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
		log.Printf("[MIDDLEWARE]: error starting mongo transaction: %s", err.Error())
		handlerutil.RespondWithError(ctx, err)
	} else {
		ctx.Next()

		if len(ctx.Errors) > 0 {
			log.Printf("[MIDDLEWARE]: error in request context, rolling back transaction")
			srv.db.AbortTransaction(ctx.Request.Context())
		} else {
			log.Printf("[MIDDLEWARE]: commiting transaction")
			srv.db.CommitTransaction(ctx.Request.Context())
		}
	}
}
