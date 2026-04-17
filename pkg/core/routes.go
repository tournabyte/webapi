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
			handlerutil.AwaitAndRespondAs[models.EventRecord],
			http.StatusOK,
			eventRecordKey,
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
	eventGroup.DELETE(
		"/:eventid",
		srv.withMongoSession,
		srv.withMongoTransaction,
		handlerutil.HandlerTemplate(
			srv.initEventLookupWorkspace,
			eventDeletionPipeline,
			handlerutil.AwaitAndRespondAs[models.EventID],
			http.StatusOK,
			eventIDResponseKey,
			srv.errfmt,
		),
	)

	// POST /v1/events/{id}/participants
	eventGroup.POST(
		"/:eventid/participants",
		srv.withMongoSession,
		srv.withMongoTransaction,
		handlerutil.HandlerTemplate(
			srv.initParticipantCreationWorkspace,
			createParticipantPipeline,
			handlerutil.AwaitAndRespondAs[models.ParticipantID],
			http.StatusCreated,
			participatIDResponseKey,
			srv.errfmt,
		),
	)

	// GET /v1/events/{id}/participants
	eventGroup.GET(
		"/:eventid/participants",
		srv.withMongoSession,
		handlerutil.HandlerTemplate(
			srv.initEventLookupWorkspace,
			listParticipantsPipeline,
			handlerutil.AwaitAndRespondAs[[]models.EventParticipant],
			http.StatusOK,
			participantListRecordsKey,
			srv.errfmt,
		),
	)

	// GET /v1/events/{id}/participants/{id}
	eventGroup.GET(
		"/:eventid/participants/:playerid",
		srv.withMongoSession,
		handlerutil.HandlerTemplate(
			srv.initParticipantLookupWorkspace,
			getParticipantPipeline,
			handlerutil.AwaitAndRespondAs[models.EventParticipant],
			http.StatusOK,
			participantRecordKey,
			srv.errfmt,
		),
	)

	// PUT /v1/events/{id}/participants/{id}
	eventGroup.PUT(
		"/:eventid/participants/:playerid",
		srv.withMongoSession,
		srv.withMongoTransaction,
		handlerutil.HandlerTemplate(
			srv.initParticipantUpdateWorkspace,
			updateParticipantPipeline,
			handlerutil.AwaitAndRespondAs[models.ParticipantID],
			http.StatusOK,
			participatIDResponseKey,
			srv.errfmt,
		),
	)

	// DELETE /v1/events/{id}/participants/{id}
	eventGroup.DELETE(
		"/:eventid/participants/:playerid",
		srv.withMongoSession,
		srv.withMongoTransaction,
		handlerutil.HandlerTemplate(
			srv.initParticipantLookupWorkspace,
			removeParticipantPipeline,
			handlerutil.AwaitAndRespondAs[models.ParticipantID],
			http.StatusOK,
			participatIDResponseKey,
			srv.errfmt,
		),
	)

	// POST /v1/events/{id}/matches
	eventGroup.POST(
		"/:eventid/matches",
		srv.withMongoSession,
		srv.withMongoTransaction,
		handlerutil.HandlerTemplate(
			srv.initEventLookupWorkspace,
			createMatchSetPipeline,
			handlerutil.AwaitAndRespondAs[models.EventID],
			http.StatusCreated,
			eventIDResponseKey,
			srv.errfmt,
		),
	)

	// GET /v1/events/{id}/matches
	eventGroup.GET(
		"/:eventid/matches",
		srv.withMongoSession,
		srv.withMongoTransaction,
		handlerutil.HandlerTemplate(
			srv.initEventLookupWorkspace,
			getMatchSetPipeline,
			handlerutil.AwaitAndRespondAs[[]models.EventMatch],
			http.StatusOK,
			matchListRecordKey,
			srv.errfmt,
		),
	)

	// GET /v1/events/{id}/matches/{id}
	eventGroup.GET(
		"/:eventid/matches/:matchid",
		srv.withMongoSession,
		srv.withMongoTransaction,
		handlerutil.HandlerTemplate(
			srv.initMatchLookupWorkspace,
			getMatchPipeline,
			handlerutil.AwaitAndRespondAs[models.EventMatch],
			http.StatusOK,
			matchRecordKey,
			srv.errfmt,
		),
	)

	// PATCH /v1/events/{id}/matches/{id}/away-participant
	eventGroup.PATCH(
		"/:eventid/matches/:matchid/away-participant",
		srv.withMongoSession,
		srv.withMongoTransaction,
		handlerutil.HandlerTemplate(
			srv.initMatchLookupWorkspace,
			tryResolveAwayParticipantPipeline,
			handlerutil.AwaitAndRespondAs[models.MatchID],
			http.StatusOK,
			matchIDResponseKey,
			srv.errfmt,
		),
	)

	// PATCH /v1/events/{id}/matches/{id}/home-participant
	eventGroup.PATCH(
		"/:eventid/matches/:matchid/home-participant",
		srv.withMongoSession,
		srv.withMongoTransaction,
		handlerutil.HandlerTemplate(
			srv.initMatchLookupWorkspace,
			tryResolveHomeParticipantPipeline,
			handlerutil.AwaitAndRespondAs[models.MatchID],
			http.StatusOK,
			matchIDResponseKey,
			srv.errfmt,
		),
	)

	// PATCH /v1/events/{id}/matches/{id}/declared-winner
	eventGroup.PATCH(
		"/:eventid/matches/:matchid/declared-winner",
		srv.withMongoSession,
		srv.withMongoTransaction,
		handlerutil.HandlerTemplate(
			srv.initMatchLookupWorkspace,
			declareMatchWinnerPipeline,
			handlerutil.AwaitAndRespondAs[models.MatchID],
			http.StatusOK,
			matchIDResponseKey,
			srv.errfmt,
		),
	)
}
