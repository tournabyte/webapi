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

// Function `(*tournabyteAPIService).addGlobalMiddleware` configures the `gin.Engine` instance with global-level middleware
func (srv *tournabyteAPIService) addGlobalMiddleware() {
	srv.router.Use(gin.CustomRecovery(srv.recoverPanicAsFailure))
	srv.router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: srv.serviceLoggerFmt,
		Output:    log.Writer(),
	}))
	srv.router.Use(requestid.New())
}

// Function `(*tournabyteAPIService).registerRoutes` configures the `gin.Engine` instance with the application HTTP handlers
func (srv *tournabyteAPIService) registerRoutes() {
	srv.addGlobalMiddleware()

	{
		// /v1/...
		v1 := srv.router.Group("v1")
		srv.addAuthGroup(v1)
		srv.addEventGroup(v1)
	}
}

// Function `(*tournabyteAPIService).addAuthGroup` configures the `gin.Engine` instance with authentication/authorization related endpoints
//
// Parameters:
//   - parentGroup: the parent portion of the API endpoint these handlers will be attached to
func (srv *tournabyteAPIService) addAuthGroup(parentGroup *gin.RouterGroup) {
	authGroup := parentGroup.Group("users")

	// POST /v1/users
	authGroup.POST(
		"/",
		srv.withMongoSession,
		srv.withMongoTransaction,
		handlerutil.HandlerTemplate(
			srv.initAuthWorkspace,
			userCreationPipeline,
			handlerutil.AwaitAndRespondAs[models.AuthenticatedUser],
			http.StatusCreated,
			userAuthorizationResponseKey,
			srv.errfmt,
		),
	)

	// POST /v1/users/tokens
	authGroup.POST(
		"/tokens",
		srv.withMongoSession,
		srv.withMongoTransaction,
		handlerutil.HandlerTemplate(
			srv.initAuthWorkspace,
			userAuthenticationPipeline,
			handlerutil.AwaitAndRespondAs[models.AuthenticatedUser],
			http.StatusOK,
			userAuthorizationResponseKey,
			srv.errfmt,
		),
	)

	// PUT /v1/users/tokens
	authGroup.PUT(
		"/tokens",
		srv.withMongoSession,
		srv.withMongoTransaction,
		handlerutil.HandlerTemplate(
			srv.initAuthWorkspace,
			sessionRefreshPipeline,
			handlerutil.AwaitAndRespondAs[models.AuthenticatedUser],
			http.StatusOK,
			userAuthorizationResponseKey,
			srv.errfmt,
		),
	)

	// DELETE /v1/users/tokens/{id}
	authGroup.DELETE(
		"/tokens/:sessionid",
		srv.withMongoSession,
		srv.withMongoTransaction,
		handlerutil.HandlerTemplate(
			srv.initAuthWorkspace,
			sessionClosePipeline,
			handlerutil.AwaitAndRespondAs[gin.H],
			http.StatusOK,
			userLogoutResponseKey,
			srv.errfmt,
		),
	)

}

// Function `(*tournabyteAPIService).addEventGroup` configures the `gin.Engine` instance with event management related endpoints
//
// Parameters:
//   - parentGroup: the parent portion of the API endpoint these handlers will be attached to
func (srv *tournabyteAPIService) addEventGroup(parentGroup *gin.RouterGroup) {
	eventGroup := parentGroup.Group("events")

	// POST /v1/events
	eventGroup.POST(
		"/",
		srv.withMongoSession,
		srv.withMongoTransaction,
		handlerutil.HandlerTemplate(
			srv.initEventCreationWorkspace,
			eventCreationPipeline,
			handlerutil.AwaitAndRespondAs[models.EventID],
			http.StatusCreated,
			eventIDResponseKey,
			srv.errfmt,
		),
	)

	// GET /v1/events/{id}
	eventGroup.GET(
		"/:eventid",
		srv.withMongoSession,
		handlerutil.HandlerTemplate(
			srv.initEventLookupWorkspace,
			eventRetreivalPipeline,
			handlerutil.AwaitAndRespondAs[models.EventDetailsResponse],
			http.StatusOK,
			eventDetailsResponseKey,
			srv.errfmt,
		),
	)

	// PUT /v1/events/{id}
	eventGroup.PUT(
		"/:eventid",
		srv.withMongoSession,
		srv.withMongoTransaction,
		handlerutil.HandlerTemplate(
			srv.initEventUpdateWorkspace,
			eventModificiationPipeline,
			handlerutil.AwaitAndRespondAs[models.EventID],
			http.StatusOK,
			eventIDResponseKey,
			srv.errfmt,
		),
	)

	// DELETE /v1/events/{id}

	// POST /v1/events/{id}/participants
	// PUT /v1/events/{id}/participants/{name}
	// DELETE /v1/events/{id}/participants/{name}

	// POST /v1/events/{id}/bracket
	// GET /v1/events/{id}/bracket
}
