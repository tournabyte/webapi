package core

/*
 * File: pkg/core/routes.go
 *
 * Purpose: definition of the server routing rules
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"log"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/tournabyte/webapi/pkg/handlerutil"
	"github.com/tournabyte/webapi/pkg/models"
)

func (srv *tournabyteAPIService) addGlobalMiddleware() {
	srv.router.Use(gin.CustomRecovery(srv.recoverPanicAsFailure))
	srv.router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: srv.serviceLoggerFmt,
		Output:    log.Writer(),
	}))
	srv.router.Use(requestid.New())
}

func (srv *tournabyteAPIService) registerRoutes() {
	srv.addGlobalMiddleware()

	{
		// /v1/...
		v1 := srv.router.Group("v1")
		srv.addAuthGroup(v1)
	}
}

func (srv *tournabyteAPIService) addAuthGroup(parentGroup *gin.RouterGroup) {
	authGroup := parentGroup.Group("users")

	// POST /v1/users
	authGroup.POST(
		"/",
		srv.withMongoSession,
		srv.withMongoTransaction,
		handlerutil.HandlerTemplate(
			srv.initUserCreationWorkspace,
			userCreationPipeline,
			handlerutil.AwaitAndRespondAs[models.AuthenticatedUser],
			http.StatusCreated,
			userAuthorizationResponseKey,
			srv.errfmt,
		),
	)

	// POST /v1/users/tokens

	// DELETE /v1/users/tokens/{id}

	// PUT /v1/users/credentials
}
