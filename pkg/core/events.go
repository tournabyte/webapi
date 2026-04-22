package core

/*
 * File: pkg/core/events.go
 *
 * Purpose: event management logic
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"context"
	"errors"
	"log"
	"math/bits"

	"github.com/gin-gonic/gin"
	"github.com/tournabyte/webapi/pkg/dbx"
	"github.com/tournabyte/webapi/pkg/handlerutil"
	"github.com/tournabyte/webapi/pkg/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Workspace keys associated with event management workspace tasks
const (
	eventCreationRequest       = "createEventRequest"
	eventLookupRequest         = "lookupEventRequest"
	eventUpdateRequest         = "updateEventRequest"
	eventRecordKey             = "eventRecord"
	eventIDResponseKey         = "eventIDResponse"
	participatIDResponseKey    = "participantIdResponse"
	participantCreationRequest = "createParticipantRequest"
	participantLookupRequest   = "lookupParticipantRequest"
	participantRecordKey       = "participantRecord"
	participantListRecordsKey  = "participantRecordList"
	matchRecordKey             = "matchRecord"
	matchListRecordKey         = "matchListRecords"
	matchLookupRequest         = "matchLookupRequest"
	matchDeclareWinnerRequest  = "matchWinnerDeclaredRequest"
	matchIDResponseKey         = "matchIDResponse"
)

// Function `(*tournabyteAPIService).initEventCreationWorkspace` initializes the handler workspace for an event creation request handling sequence
//
// Parameters:
//   - ctx: the request context to use during workspace initialization
//
// Returns:
//   - `*handlerutil.HandlerWorkspace`: the workspace for creating a user
func (srv *tournabyteAPIService) initEventCreationWorkspace(ctx *gin.Context) *handlerutil.HandlerWorkspace {
	space := handlerutil.DefaultWorkspace()
	binds := handlerutil.BindingsFromRequestContext(ctx, handlerutil.ShouldHaveJSONBody|handlerutil.ShouldHaveHeaders)

	space.Set(handlerutil.RequestBindings, binds)
	space.Set(authTokenOptionsKey, srv.getTokenConfig())
	space.Set(models.ValidatorObjectKey, srv.validationFunc)
	log.Printf("[HANDLER]: setup request bindings")
	return &space

}

// Function `(*tournabyteAPIService).initEventLookupWorkspace` initializes the handler workspace for an event lookup request handling sequence
//
// Parameters:
//   - ctx: the request context to use during workspace initialization
//
// Returns:
//   - `*handlerutil.HandlerWorkspace`: the workspace for finding an event
func (srv *tournabyteAPIService) initEventLookupWorkspace(ctx *gin.Context) *handlerutil.HandlerWorkspace {
	space := handlerutil.DefaultWorkspace()
	binds := handlerutil.BindingsFromRequestContext(ctx, handlerutil.ShouldHaveURIValues|handlerutil.ShouldHaveHeaders)

	space.Set(handlerutil.RequestBindings, binds)
	space.Set(authTokenOptionsKey, srv.getTokenConfig())
	space.Set(models.ValidatorObjectKey, srv.validationFunc)
	log.Printf("[HANDLER]: setup request bindings")
	return &space
}

// Function `(*tournabyteAPIService).initEventUpdateWorkspace` initializes the handler workspace for an event update request handling sequence
//
// Parameters:
//   - ctx: the request context to use during workspace initialization
//
// Returns:
//   - `*handlerutil.HandlerWorkspace`: the workspace for updating an event
func (srv *tournabyteAPIService) initEventUpdateWorkspace(ctx *gin.Context) *handlerutil.HandlerWorkspace {
	space := handlerutil.DefaultWorkspace()
	binds := handlerutil.BindingsFromRequestContext(ctx, handlerutil.ShouldHaveURIValues|handlerutil.ShouldHaveHeaders|handlerutil.ShouldHaveJSONBody)

	space.Set(handlerutil.RequestBindings, binds)
	space.Set(authTokenOptionsKey, srv.getTokenConfig())
	space.Set(models.ValidatorObjectKey, srv.validationFunc)
	log.Printf("[HANDLER]: setup request bindings")
	return &space
}

// Function `(*tournabyteAPIService).initParticipantCreationWorkspace` initializes the handler workspace for an participant request handling sequence
//
// Parameters:
//   - ctx: the request context to use during workspace initialization
//
// Returns:
//   - `*handlerutil.HandlerWorkspace`: the workspace for updating an event
func (srv *tournabyteAPIService) initParticipantCreationWorkspace(ctx *gin.Context) *handlerutil.HandlerWorkspace {
	space := handlerutil.DefaultWorkspace()
	binds := handlerutil.BindingsFromRequestContext(ctx, handlerutil.ShouldHaveURIValues|handlerutil.ShouldHaveHeaders|handlerutil.ShouldHaveJSONBody)

	space.Set(handlerutil.RequestBindings, binds)
	space.Set(authTokenOptionsKey, srv.getTokenConfig())
	space.Set(models.ValidatorObjectKey, srv.validationFunc)
	log.Printf("[HANDLER]: setup request bindings")
	return &space
}

// Function `(*tournabyteAPIService).initParticipantLookupWorkspace` initializes the handler workspace for an participant lookup request handling sequence
//
// Parameters:
//   - ctx: the request context to use during workspace initialization
//
// Returns:
//   - `*handlerutil.HandlerWorkspace`: the workspace for finding a participant
func (srv *tournabyteAPIService) initParticipantLookupWorkspace(ctx *gin.Context) *handlerutil.HandlerWorkspace {
	space := handlerutil.DefaultWorkspace()
	binds := handlerutil.BindingsFromRequestContext(ctx, handlerutil.ShouldHaveURIValues|handlerutil.ShouldHaveHeaders)

	space.Set(handlerutil.RequestBindings, binds)
	space.Set(authTokenOptionsKey, srv.getTokenConfig())
	space.Set(models.ValidatorObjectKey, srv.validationFunc)
	log.Printf("[HANDLER]: setup request bindings")
	return &space
}

// Function `(*tournabyteAPIService).initParticipantUpdateWorkspace` initializes the handler workspace for an participant update request handling sequence
//
// Parameters:
//   - ctx: the request context to use during workspace initialization
//
// Returns:
//   - `*handlerutil.HandlerWorkspace`: the workspace for finding a participant
func (srv *tournabyteAPIService) initParticipantUpdateWorkspace(ctx *gin.Context) *handlerutil.HandlerWorkspace {
	space := handlerutil.DefaultWorkspace()
	binds := handlerutil.BindingsFromRequestContext(ctx, handlerutil.ShouldHaveURIValues|handlerutil.ShouldHaveHeaders|handlerutil.ShouldHaveJSONBody)

	space.Set(handlerutil.RequestBindings, binds)
	space.Set(authTokenOptionsKey, srv.getTokenConfig())
	space.Set(models.ValidatorObjectKey, srv.validationFunc)
	log.Printf("[HANDLER]: setup request bindings")
	return &space
}

// Function `(*tournabyteAPIService).initMatchLookupWorkspace` initializes the handler workspace for an match lookup request handling sequence
//
// Parameters:
//   - ctx: the request context to use during workspace initialization
//
// Returns:
//   - `*handlerutil.HandlerWorkspace`: the workspace for finding a participant
func (srv *tournabyteAPIService) initMatchLookupWorkspace(ctx *gin.Context) *handlerutil.HandlerWorkspace {
	space := handlerutil.DefaultWorkspace()
	binds := handlerutil.BindingsFromRequestContext(ctx, handlerutil.ShouldHaveURIValues|handlerutil.ShouldHaveHeaders)

	space.Set(handlerutil.RequestBindings, binds)
	space.Set(authTokenOptionsKey, srv.getTokenConfig())
	space.Set(models.ValidatorObjectKey, srv.validationFunc)
	log.Printf("[HANDLER]: setup request bindings")
	return &space
}

// Function `(*tournabyteAPIService).initMatchUpdateWorkspace` initializes the handler workspace for an update match request handling sequence
//
// Parameters:
//   - ctx: the request context to use during workspace initialization
//
// Returns:
//   - `*handlerutil.HandlerWorkspace`: the workspace for finding a participant
func (srv *tournabyteAPIService) initMatchUpdateWorkspace(ctx *gin.Context) *handlerutil.HandlerWorkspace {
	space := handlerutil.DefaultWorkspace()
	binds := handlerutil.BindingsFromRequestContext(ctx, handlerutil.ShouldHaveURIValues|handlerutil.ShouldHaveHeaders|handlerutil.ShouldHaveJSONBody)

	space.Set(handlerutil.RequestBindings, binds)
	space.Set(authTokenOptionsKey, srv.getTokenConfig())
	space.Set(models.ValidatorObjectKey, srv.validationFunc)
	log.Printf("[HANDLER]: setup request bindings")
	return &space
}

// Function `eventCreationPipeline` initializes a handling pipeline for event creation
//
// Parameters:
//   - ctx: the parent context to control the created pipeline
//
// Returns:
//   - `context.Context`: the context controlling the created pipeline (derived from the given context.Context)
//   - `context.CancelCauseFunc`: the cancellation function controlling pipeline cancellation
//   - `chan<- *handlerutil.HandlerWorkspace`: the input channel for the pipeline (send-only)
//   - `<-chan *handlerutil.HandlerWorkspace`: the output channel for the pipeline (read-only)
func eventCreationPipeline(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
	pipelineCtx, pipelineCancel := context.WithCancelCause(ctx)
	pipelineInput := make(chan *handlerutil.HandlerWorkspace)

	out1 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindAccessTokenFromHeader, pipelineInput)
	out2 := handlerutil.Stage(pipelineCtx, pipelineCancel, validateAccessToken, out1)
	out3 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindEventCreationRequestFromBody, out2)
	out4 := handlerutil.Stage(pipelineCtx, pipelineCancel, deriveEventRecordFromRequest, out3)
	out5 := handlerutil.Stage(pipelineCtx, pipelineCancel, createEventRecord, out4)
	pipelineOutput := handlerutil.Stage(pipelineCtx, pipelineCancel, populateEventIDResponse, out5)

	return pipelineCtx, pipelineCancel, pipelineInput, pipelineOutput
}

// Function `eventRetreivalPipeline` initializes a handling pipeline for event retrieval
//
// Parameters:
//   - ctx: the parent context to control the created pipeline
//
// Returns:
//   - `context.Context`: the context controlling the created pipeline (derived from the given context.Context)
//   - `context.CancelCauseFunc`: the cancellation function controlling pipeline cancellation
//   - `chan<- *handlerutil.HandlerWorkspace`: the input channel for the pipeline (send-only)
//   - `<-chan *handlerutil.HandlerWorkspace`: the output channel for the pipeline (read-only)
func eventRetreivalPipeline(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
	pipelineCtx, pipelineCancel := context.WithCancelCause(ctx)
	pipelineInput := make(chan *handlerutil.HandlerWorkspace)

	out1 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindAccessTokenFromHeader, pipelineInput)
	out2 := handlerutil.Stage(pipelineCtx, pipelineCancel, validateAccessToken, out1)
	out3 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindEventLookupRequestFromURI, out2)
	pipelineOutput := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchEventRecordFromDatabaseByID, out3)

	return pipelineCtx, pipelineCancel, pipelineInput, pipelineOutput
}

// Function `eventModificiationPipeline` initializes a handling pipeline for event modification
//
// Parameters:
//   - ctx: the parent context to control the created pipeline
//
// Returns:
//   - `context.Context`: the context controlling the created pipeline (derived from the given context.Context)
//   - `context.CancelCauseFunc`: the cancellation function controlling pipeline cancellation
//   - `chan<- *handlerutil.HandlerWorkspace`: the input channel for the pipeline (send-only)
//   - `<-chan *handlerutil.HandlerWorkspace`: the output channel for the pipeline (read-only)
func eventModificiationPipeline(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
	pipelineCtx, pipelineCancel := context.WithCancelCause(ctx)
	pipelineInput := make(chan *handlerutil.HandlerWorkspace)

	out1 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindAccessTokenFromHeader, pipelineInput)
	out2 := handlerutil.Stage(pipelineCtx, pipelineCancel, validateAccessToken, out1)
	out3 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindEventLookupRequestFromURI, out2)
	out4 := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchEventRecordFromDatabaseByID, out3)
	out5 := handlerutil.Stage(pipelineCtx, pipelineCancel, verifyEventOwnership, out4)
	out6 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindEventModificationRequestFromBody, out5)
	out7 := handlerutil.Stage(pipelineCtx, pipelineCancel, applyEventRecordModificationByID, out6)
	pipelineOutput := handlerutil.Stage(pipelineCtx, pipelineCancel, populateEventIDResponse, out7)

	return pipelineCtx, pipelineCancel, pipelineInput, pipelineOutput
}

// Function `eventDeletionPipeline` initializes a handling pipeline for event deletion
//
// Parameters:
//   - ctx: the parent context to control the created pipeline
//
// Returns:
//   - `context.Context`: the context controlling the created pipeline (derived from the given context.Context)
//   - `context.CancelCauseFunc`: the cancellation function controlling pipeline cancellation
//   - `chan<- *handlerutil.HandlerWorkspace`: the input channel for the pipeline (send-only)
//   - `<-chan *handlerutil.HandlerWorkspace`: the output channel for the pipeline (read-only)
func eventDeletionPipeline(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
	pipelineCtx, pipelineCancel := context.WithCancelCause(ctx)
	pipelineInput := make(chan *handlerutil.HandlerWorkspace)

	out1 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindAccessTokenFromHeader, pipelineInput)
	out2 := handlerutil.Stage(pipelineCtx, pipelineCancel, validateAccessToken, out1)
	out3 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindEventLookupRequestFromURI, out2)
	out4 := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchEventRecordFromDatabaseByID, out3)
	out5 := handlerutil.Stage(pipelineCtx, pipelineCancel, verifyEventOwnership, out4)
	out6 := handlerutil.Stage(pipelineCtx, pipelineCancel, removeEventRecordByID, out5)
	pipelineOutput := handlerutil.Stage(pipelineCtx, pipelineCancel, populateEventIDResponse, out6)

	return pipelineCtx, pipelineCancel, pipelineInput, pipelineOutput
}

// Function `createParticipantPipeline` initializes a handling pipeline for adding to an event's participant list
//
// Parameters:
//   - ctx: the parent context to control the created pipeline
//
// Returns:
//   - `context.Context`: the context controlling the created pipeline (derived from the given context.Context)
//   - `context.CancelCauseFunc`: the cancellation function controlling pipeline cancellation
//   - `chan<- *handlerutil.HandlerWorkspace`: the input channel for the pipeline (send-only)
//   - `<-chan *handlerutil.HandlerWorkspace`: the output channel for the pipeline (read-only)
func createParticipantPipeline(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
	pipelineCtx, pipelineCancel := context.WithCancelCause(ctx)
	pipelineInput := make(chan *handlerutil.HandlerWorkspace)

	out1 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindAccessTokenFromHeader, pipelineInput)
	out2 := handlerutil.Stage(pipelineCtx, pipelineCancel, validateAccessToken, out1)
	out3 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindEventLookupRequestFromURI, out2)
	out4 := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchEventRecordFromDatabaseByID, out3)
	out5 := handlerutil.Stage(pipelineCtx, pipelineCancel, verifyEventOwnership, out4)
	out6 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindNewParticipantRequestFromBody, out5)
	out7 := handlerutil.Stage(pipelineCtx, pipelineCancel, deriveParticipantRecordFromRequest, out6)
	out8 := handlerutil.Stage(pipelineCtx, pipelineCancel, verifyEventModifiable, out7)
	pipelineOutput := handlerutil.Stage(pipelineCtx, pipelineCancel, createParticipantRecord, out8)

	return pipelineCtx, pipelineCancel, pipelineInput, pipelineOutput
}

// Function `listParticipantsPipeline` initializes a handling pipeline for retrieving an event's participant list
//
// Parameters:
//   - ctx: the parent context to control the created pipeline
//
// Returns:
//   - `context.Context`: the context controlling the created pipeline (derived from the given context.Context)
//   - `context.CancelCauseFunc`: the cancellation function controlling pipeline cancellation
//   - `chan<- *handlerutil.HandlerWorkspace`: the input channel for the pipeline (send-only)
//   - `<-chan *handlerutil.HandlerWorkspace`: the output channel for the pipeline (read-only)
func listParticipantsPipeline(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
	pipelineCtx, pipelineCancel := context.WithCancelCause(ctx)
	pipelineInput := make(chan *handlerutil.HandlerWorkspace)

	out1 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindAccessTokenFromHeader, pipelineInput)
	out2 := handlerutil.Stage(pipelineCtx, pipelineCancel, validateAccessToken, out1)
	out3 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindEventLookupRequestFromURI, out2)
	pipelineOutput := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchParticipantsFromDatabaseByEventID, out3)

	return pipelineCtx, pipelineCancel, pipelineInput, pipelineOutput
}

// Function `getParticipantPipeline` initializes a handling pipeline for retrieving a participant by its ID
//
// Parameters:
//   - ctx: the parent context to control the created pipeline
//
// Returns:
//   - `context.Context`: the context controlling the created pipeline (derived from the given context.Context)
//   - `context.CancelCauseFunc`: the cancellation function controlling pipeline cancellation
//   - `chan<- *handlerutil.HandlerWorkspace`: the input channel for the pipeline (send-only)
//   - `<-chan *handlerutil.HandlerWorkspace`: the output channel for the pipeline (read-only)
func getParticipantPipeline(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
	pipelineCtx, pipelineCancel := context.WithCancelCause(ctx)
	pipelineInput := make(chan *handlerutil.HandlerWorkspace)

	out1 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindAccessTokenFromHeader, pipelineInput)
	out2 := handlerutil.Stage(pipelineCtx, pipelineCancel, validateAccessToken, out1)
	out3 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindParticipantLookupRequestFromURI, out2)
	pipelineOutput := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchParticipantFromDatabaseByPlayerID, out3)

	return pipelineCtx, pipelineCancel, pipelineInput, pipelineOutput
}

// Function `updateParticipantPipeline` initializes a handling pipeline for retrieving a participant by its ID
//
// Parameters:
//   - ctx: the parent context to control the created pipeline
//
// Returns:
//   - `context.Context`: the context controlling the created pipeline (derived from the given context.Context)
//   - `context.CancelCauseFunc`: the cancellation function controlling pipeline cancellation
//   - `chan<- *handlerutil.HandlerWorkspace`: the input channel for the pipeline (send-only)
//   - `<-chan *handlerutil.HandlerWorkspace`: the output channel for the pipeline (read-only)
func updateParticipantPipeline(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
	pipelineCtx, pipelineCancel := context.WithCancelCause(ctx)
	pipelineInput := make(chan *handlerutil.HandlerWorkspace)

	out1 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindAccessTokenFromHeader, pipelineInput)
	out2 := handlerutil.Stage(pipelineCtx, pipelineCancel, validateAccessToken, out1)
	out3 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindParticipantLookupRequestFromURI, out2)
	out4 := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchEventRecordFromDatabaseByID, out3)
	out5 := handlerutil.Stage(pipelineCtx, pipelineCancel, verifyEventOwnership, out4)
	out6 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindNewParticipantRequestFromBody, out5)
	out7 := handlerutil.Stage(pipelineCtx, pipelineCancel, verifyEventModifiable, out6)
	pipelineOutput := handlerutil.Stage(pipelineCtx, pipelineCancel, updateParticipantRecord, out7)

	return pipelineCtx, pipelineCancel, pipelineInput, pipelineOutput
}

// Function `removeParticipantPipeline` initializes a handling pipeline for removing a participant by its ID
//
// Parameters:
//   - ctx: the parent context to control the created pipeline
//
// Returns:
//   - `context.Context`: the context controlling the created pipeline (derived from the given context.Context)
//   - `context.CancelCauseFunc`: the cancellation function controlling pipeline cancellation
//   - `chan<- *handlerutil.HandlerWorkspace`: the input channel for the pipeline (send-only)
//   - `<-chan *handlerutil.HandlerWorkspace`: the output channel for the pipeline (read-only)
func removeParticipantPipeline(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
	pipelineCtx, pipelineCancel := context.WithCancelCause(ctx)
	pipelineInput := make(chan *handlerutil.HandlerWorkspace)

	out1 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindAccessTokenFromHeader, pipelineInput)
	out2 := handlerutil.Stage(pipelineCtx, pipelineCancel, validateAccessToken, out1)
	out3 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindParticipantLookupRequestFromURI, out2)
	out4 := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchEventRecordFromDatabaseByID, out3)
	out5 := handlerutil.Stage(pipelineCtx, pipelineCancel, verifyEventOwnership, out4)
	out6 := handlerutil.Stage(pipelineCtx, pipelineCancel, verifyEventModifiable, out5)
	pipelineOutput := handlerutil.Stage(pipelineCtx, pipelineCancel, removeParticipantRecord, out6)

	return pipelineCtx, pipelineCancel, pipelineInput, pipelineOutput
}

// Function `createMatchSetPipeline` initializes a handling pipeline for creating a single-elimination match set for the given event ID
//
// Parameters:
//   - ctx: the parent context to control the created pipeline
//
// Returns:
//   - `context.Context`: the context controlling the created pipeline (derived from the given context.Context)
//   - `context.CancelCauseFunc`: the cancellation function controlling pipeline cancellation
//   - `chan<- *handlerutil.HandlerWorkspace`: the input channel for the pipeline (send-only)
//   - `<-chan *handlerutil.HandlerWorkspace`: the output channel for the pipeline (read-only)
func createMatchSetPipeline(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
	pipelineCtx, pipelineCancel := context.WithCancelCause(ctx)
	pipelineInput := make(chan *handlerutil.HandlerWorkspace)

	out1 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindAccessTokenFromHeader, pipelineInput)
	out2 := handlerutil.Stage(pipelineCtx, pipelineCancel, validateAccessToken, out1)
	out3 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindEventLookupRequestFromURI, out2)
	out4 := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchEventRecordFromDatabaseByID, out3)
	out5 := handlerutil.Stage(pipelineCtx, pipelineCancel, verifyEventOwnership, out4)
	out6 := handlerutil.Stage(pipelineCtx, pipelineCancel, verifyEventModifiable, out5)
	out7 := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchParticipantsFromDatabaseByEventID, out6)
	out8 := handlerutil.Stage(pipelineCtx, pipelineCancel, deriveMatchSetFromParticipantList, out7)
	pipelineOutput := handlerutil.Stage(pipelineCtx, pipelineCancel, createMatchSetRecord, out8)

	return pipelineCtx, pipelineCancel, pipelineInput, pipelineOutput
}

// Function `getMatchSetPipeline` initializes a handling pipeline for retrieving all matches associated with an event ID
//
// Parameters:
//   - ctx: the parent context to control the created pipeline
//
// Returns:
//   - `context.Context`: the context controlling the created pipeline (derived from the given context.Context)
//   - `context.CancelCauseFunc`: the cancellation function controlling pipeline cancellation
//   - `chan<- *handlerutil.HandlerWorkspace`: the input channel for the pipeline (send-only)
//   - `<-chan *handlerutil.HandlerWorkspace`: the output channel for the pipeline (read-only)
func getMatchSetPipeline(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
	pipelineCtx, pipelineCancel := context.WithCancelCause(ctx)
	pipelineInput := make(chan *handlerutil.HandlerWorkspace)

	out1 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindAccessTokenFromHeader, pipelineInput)
	out2 := handlerutil.Stage(pipelineCtx, pipelineCancel, validateAccessToken, out1)
	out3 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindEventLookupRequestFromURI, out2)
	pipelineOutput := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchMatchSetFromDatabaseByEventID, out3)

	return pipelineCtx, pipelineCancel, pipelineInput, pipelineOutput

}

// Function `getMatchPipeline` initializes a handling pipeline for finding a match by its ID
//
// Parameters:
//   - ctx: the parent context to control the created pipeline
//
// Returns:
//   - `context.Context`: the context controlling the created pipeline (derived from the given context.Context)
//   - `context.CancelCauseFunc`: the cancellation function controlling pipeline cancellation
//   - `chan<- *handlerutil.HandlerWorkspace`: the input channel for the pipeline (send-only)
//   - `<-chan *handlerutil.HandlerWorkspace`: the output channel for the pipeline (read-only)
func getMatchPipeline(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
	pipelineCtx, pipelineCancel := context.WithCancelCause(ctx)
	pipelineInput := make(chan *handlerutil.HandlerWorkspace)

	out1 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindAccessTokenFromHeader, pipelineInput)
	out2 := handlerutil.Stage(pipelineCtx, pipelineCancel, validateAccessToken, out1)
	out3 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindMatchLookupRequestFromURI, out2)
	pipelineOutput := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchMatchFromDatabaseByID, out3)

	return pipelineCtx, pipelineCancel, pipelineInput, pipelineOutput
}

// Function `tryResolveAwayParticipantPipeline` initializes a handling pipeline for resolving a match's away participant
//
// Parameters:
//   - ctx: the parent context to control the created pipeline
//
// Returns:
//   - `context.Context`: the context controlling the created pipeline (derived from the given context.Context)
//   - `context.CancelCauseFunc`: the cancellation function controlling pipeline cancellation
//   - `chan<- *handlerutil.HandlerWorkspace`: the input channel for the pipeline (send-only)
//   - `<-chan *handlerutil.HandlerWorkspace`: the output channel for the pipeline (read-only)
func tryResolveAwayParticipantPipeline(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
	pipelineCtx, pipelineCancel := context.WithCancelCause(ctx)
	pipelineInput := make(chan *handlerutil.HandlerWorkspace)

	out1 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindAccessTokenFromHeader, pipelineInput)
	out2 := handlerutil.Stage(pipelineCtx, pipelineCancel, validateAccessToken, out1)
	out3 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindMatchLookupRequestFromURI, out2)
	out4 := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchEventRecordFromDatabaseByID, out3)
	out5 := handlerutil.Stage(pipelineCtx, pipelineCancel, verifyEventOwnership, out4)
	out6 := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchMatchFromDatabaseByID, out5)
	pipelineOutput := handlerutil.Stage(pipelineCtx, pipelineCancel, updateAwayParticipantIfAvailable, out6)

	return pipelineCtx, pipelineCancel, pipelineInput, pipelineOutput
}

// Function `tryResolveHomeParticipantPipeline` initializes a handling pipeline for resolving a match's home participant
//
// Parameters:
//   - ctx: the parent context to control the created pipeline
//
// Returns:
//   - `context.Context`: the context controlling the created pipeline (derived from the given context.Context)
//   - `context.CancelCauseFunc`: the cancellation function controlling pipeline cancellation
//   - `chan<- *handlerutil.HandlerWorkspace`: the input channel for the pipeline (send-only)
//   - `<-chan *handlerutil.HandlerWorkspace`: the output channel for the pipeline (read-only)
func tryResolveHomeParticipantPipeline(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
	pipelineCtx, pipelineCancel := context.WithCancelCause(ctx)
	pipelineInput := make(chan *handlerutil.HandlerWorkspace)

	out1 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindAccessTokenFromHeader, pipelineInput)
	out2 := handlerutil.Stage(pipelineCtx, pipelineCancel, validateAccessToken, out1)
	out3 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindMatchLookupRequestFromURI, out2)
	out4 := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchEventRecordFromDatabaseByID, out3)
	out5 := handlerutil.Stage(pipelineCtx, pipelineCancel, verifyEventOwnership, out4)
	out6 := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchMatchFromDatabaseByID, out5)
	pipelineOutput := handlerutil.Stage(pipelineCtx, pipelineCancel, updateHomeParticipantIfAvailable, out6)

	return pipelineCtx, pipelineCancel, pipelineInput, pipelineOutput
}

// Function `declareMatchWinnerPipeline` initializes a handling pipeline for declaring a match winner
//
// Parameters:
//   - ctx: the parent context to control the created pipeline
//
// Returns:
//   - `context.Context`: the context controlling the created pipeline (derived from the given context.Context)
//   - `context.CancelCauseFunc`: the cancellation function controlling pipeline cancellation
//   - `chan<- *handlerutil.HandlerWorkspace`: the input channel for the pipeline (send-only)
//   - `<-chan *handlerutil.HandlerWorkspace`: the output channel for the pipeline (read-only)
func declareMatchWinnerPipeline(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
	pipelineCtx, pipelineCancel := context.WithCancelCause(ctx)
	pipelineInput := make(chan *handlerutil.HandlerWorkspace)

	out1 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindAccessTokenFromHeader, pipelineInput)
	out2 := handlerutil.Stage(pipelineCtx, pipelineCancel, validateAccessToken, out1)
	out3 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindMatchLookupRequestFromURI, out2)
	out4 := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchEventRecordFromDatabaseByID, out3)
	out5 := handlerutil.Stage(pipelineCtx, pipelineCancel, verifyEventOwnership, out4)
	out6 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindMatchWinnerDeclarationRequestFromBody, out5)
	pipelineOutput := handlerutil.Stage(pipelineCtx, pipelineCancel, updateMatchWinnerByID, out6)

	return pipelineCtx, pipelineCancel, pipelineInput, pipelineOutput
}

// Function `bindEventCreationRequestFromBody` binds the request body to the event create request format (and validates it)
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func bindEventCreationRequestFromBody(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var body models.CreateEventRequest
	var bindings handlerutil.Bindings

	log.Printf("[HANDLER]: loading request bindings from workspace...")
	if err := space.Get(handlerutil.RequestBindings, &bindings); err != nil {
		log.Printf("[HANDLER]: error loading request bindings from workspace (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: binding request body to variable of type %T", body)
	if err := bindings.BindBodyAsJSON(&body); err != nil {
		log.Printf("[HANDLER]: error binding request body (%s)", err.Error())
		return err
	}

	space.Set(eventCreationRequest, body)
	log.Printf("[HANDLER]: saved request body as variable of type %T within workspace under key %q", body, eventCreationRequest)
	return nil
}

// Function `bindEventLookupRequestFromURI` binds the request URI values to the event lookup request format (and validates it)
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func bindEventLookupRequestFromURI(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var lookup models.EventID
	var bindings handlerutil.Bindings

	log.Printf("[HANDLER]: loading request bindings from workspace...")
	if err := space.Get(handlerutil.RequestBindings, &bindings); err != nil {
		log.Printf("[HANDLER]: error loading request bindings from workspace (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: binding request URI to variable of typoe %T...", lookup)
	if err := bindings.BindURI(&lookup); err != nil {
		log.Printf("[HANDLER]: error binding request URI (%s)", err.Error())
		return err
	}

	space.Set(eventLookupRequest, lookup)
	log.Printf("[HANDLER]: saved request URI as variable of type %T within workspace under key %q", lookup, eventLookupRequest)
	return nil
}

// Function `bindParticipantLookupRequestFromURI` binds the request URI values to the participant lookup request format (and validates it)
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func bindParticipantLookupRequestFromURI(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var lookup models.ParticipantID
	var bindings handlerutil.Bindings

	log.Printf("[HANDLER]: loading request bindings from workspace...")
	if err := space.Get(handlerutil.RequestBindings, &bindings); err != nil {
		log.Printf("[HANDLER]: error loading request bindings from workspace (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: binding request URI to variable of type %T...", lookup)
	if err := bindings.BindURI(&lookup); err != nil {
		log.Printf("[HANDLER]: error binding request URI (%s)", err.Error())
		return err
	}

	space.Set(participantLookupRequest, lookup)
	space.Set(eventLookupRequest, models.EventID{ID: lookup.EID})
	log.Printf("[HANDLER]: saved request URI as variable of type %T within workspace under key %q", lookup, participantLookupRequest)
	return nil
}

// Function `bindEventModificationRequestFromBody` binds the request URI values to the event lookup request format (and validates it) and bindes the request body to the event update request format (and validates it)
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func bindEventModificationRequestFromBody(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var lookup models.EventID
	var modify models.UpdateEventRequest
	var bindings handlerutil.Bindings

	log.Printf("[HANDLER]: loading request bindings from workspace...")
	if err := space.Get(handlerutil.RequestBindings, &bindings); err != nil {
		log.Printf("[HANDLER]: error loading request bindings from workspace (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: binding request URI to variable of typoe %T...", lookup)
	if err := bindings.BindURI(&lookup); err != nil {
		log.Printf("[HANDLER]: error binding request URI (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: binding request body to variable of type %T...", modify)
	if err := bindings.BindBodyAsJSON(&modify); err != nil {
		log.Printf("[HANDLER]: error binding request body (%s)", err.Error())
		return err
	}

	space.Set(eventLookupRequest, lookup)
	space.Set(eventUpdateRequest, modify)
	log.Printf("[HANDLER]: saved request URI as variable of type %T within workspace under key %q", lookup, eventLookupRequest)
	return nil
}

// Function `bindNewParticipantRequestFromBody` binds the request body to the create participants request format (and validates it)
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func bindNewParticipantRequestFromBody(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var body models.CreateOrModifyParticipantRequest
	var bindings handlerutil.Bindings

	log.Printf("[HANDLER]: loading request bindings from workspace...")
	if err := space.Get(handlerutil.RequestBindings, &bindings); err != nil {
		log.Printf("[HANDLER]: error loading request bindings from workspace (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: binding request body to variable of type %T...", body)
	if err := bindings.BindBodyAsJSON(&body); err != nil {
		log.Printf("[HANDLER]: error binding request body (%s)", err.Error())
		return err
	}

	space.Set(participantCreationRequest, body)
	log.Printf("[HANDLER]: saved request body as variable of type %T within workspace under key %q", body, participantCreationRequest)
	return nil
}

// Function `bindMatchLookupRequestFromURI` binds the request URI to the match lookup request format (and validates it)
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func bindMatchLookupRequestFromURI(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var uri models.MatchID
	var bindings handlerutil.Bindings

	log.Printf("[HANDLER]: loading request bindings from workspace...")
	if err := space.Get(handlerutil.RequestBindings, &bindings); err != nil {
		log.Printf("[HANDLER]: error loading request bindings from workspace (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: binding request uri to variable of type %T...", uri)
	if err := bindings.BindURI(&uri); err != nil {
		log.Printf("[HANDLER]: error binding request uri (%s)", err.Error())
		return err
	}

	space.Set(matchLookupRequest, uri)
	space.Set(eventLookupRequest, models.EventID{ID: uri.EID})
	log.Printf("[HANDLER]: saved request body as variable of type %T within workspace under key %q", uri, matchLookupRequest)
	return nil
}

// Function `bindMatchWinnerDeclarationRequestFromBody` binds the request body to the match winner declaration format (and validates it)
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func bindMatchWinnerDeclarationRequestFromBody(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var body models.DeclarMatchWinnerRequest
	var bindings handlerutil.Bindings

	log.Printf("[HANDLER]: loading request bindings from workspace...")
	if err := space.Get(handlerutil.RequestBindings, &bindings); err != nil {
		log.Printf("[HANDLER]: error loading request bindings from workspace (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: binding request body to variable of type %T...", body)
	if err := bindings.BindBodyAsJSON(&body); err != nil {
		log.Printf("[HANDLER]: error binding request body (%s)", err.Error())
		return err
	}

	space.Set(matchDeclareWinnerRequest, body)
	log.Printf("[HANDLER]: saved request body as variable of type %T within workspace under key %q", body, matchDeclareWinnerRequest)
	return nil
}

// Function `verifyEventOwnership` checks that the owner of the event record is the same as presented in the access token
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func verifyEventOwnership(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var whoami string
	var userid bson.ObjectID
	var record models.EventRecord
	var err error

	log.Printf("[HANDLER]: loading user ID within access token under %q into variable of type %T...", activeUserID, whoami)
	if err = space.Get(activeUserID, &whoami); err != nil {
		log.Printf("[HANDLER]: error loading user ID (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: converting user ID hex to an ObjectID...")
	if userid, err = bson.ObjectIDFromHex(whoami); err != nil {
		log.Printf("[HANDLER]: error converting user ID hex to ObjectID (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading record data from workspace under the %q key into variable of type %T...", eventRecordKey, record)
	if err = space.Get(eventRecordKey, &record); err != nil {
		log.Printf("[HANDLER]: error loading event record data (%s)", err.Error())
		return err
	}

	log.Print("[HANDLER]: comparing token user ID to user ID associated with record...")
	if userid != record.Host {
		log.Print("[HANDLER]: ownership cannot be verified, rejecting update request")
		return errors.New("cannot update event that is not owned by you")
	}

	log.Print("[HANDLER]: ownership verified, proceeding with update")
	return nil
}

// Function `verifyEventModifiable` checks that an event record is writable (status is "PLANNED")
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func verifyEventModifiable(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var event models.EventRecord
	var err error

	log.Printf("[HANDLER]: loading event record from workspace under %q into variable of type %T...", eventRecordKey, event)
	if err = space.Get(eventRecordKey, &event); err != nil {
		log.Printf("[HANDLER]: error loading request data (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: checking if the event record status field is 'PLANNED'...")
	if event.Status != models.StatusPlanned {
		log.Printf("[HANDLER]: event status field is not 'PLANNED'; the event record (participants and matches) is not modifiable")
		return errors.New("event status disallwed modifying participants and matches")
	}

	log.Printf("[HANDLER]: event status allows for record modification")
	return nil
}

// Function `deriveEventRecordFromRequest` uses the request body within the workspace to initialize an event record
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func deriveEventRecordFromRequest(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var req models.CreateEventRequest
	var whoami string
	var id bson.ObjectID
	var record models.EventRecord
	var err error

	log.Printf("[HANDLER]: loading request data from workspace under %q into variable of type %T...", eventCreationRequest, req)
	if err = space.Get(eventCreationRequest, &req); err != nil {
		log.Printf("[HANDLER]: error loading request data (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading user ID within access token under %q into variable of type %T...", activeUserID, whoami)
	if err = space.Get(activeUserID, &whoami); err != nil {
		log.Printf("[HANDLER]: error loading user ID (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: converting user ID hex to an ObjectID...")
	if id, err = bson.ObjectIDFromHex(whoami); err != nil {
		log.Printf("[HANDLER]: error converting user ID hex to ObjectID (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: populating event record...")
	record.ID = bson.NewObjectID()
	record.Host = id
	record.Status = models.StatusPlanned
	record.Name = req.Name
	record.Game = req.Game
	record.Description = req.Description

	log.Printf("[HANDLER]: saved event record to workspace under the %q key", eventRecordKey)
	space.Set(eventRecordKey, record)
	return nil
}

// Function `deriveParticipantRecordFromRequest` uses the request body within workspace to initialize a participant record
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func deriveParticipantRecordFromRequest(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var req models.CreateOrModifyParticipantRequest
	var event models.EventRecord
	var participant models.EventParticipant
	var err error

	log.Printf("[HANDLER]: loading request data from workspace under %q into variable of type %T...", participantCreationRequest, req)
	if err = space.Get(participantCreationRequest, &req); err != nil {
		log.Printf("[HANDLER]: error loading request data (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading event record from workspace under %q into variable of type %T...", eventRecordKey, event)
	if err = space.Get(eventRecordKey, &event); err != nil {
		log.Printf("[HANDLER]: error loading request data (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: initializing participant record...")
	participant.ID = bson.NewObjectID()
	participant.DisplayName = req.DisplayName
	participant.ParticipatesIn = event.ID

	log.Printf("[HANDLER]: saved participant record to workspace under the %q key", participantRecordKey)
	space.Set(participantRecordKey, participant)
	return nil
}

// Function `deriveMatchSetFromParticipantList` uses the event participants within workspace to initialize a match bracket for an event
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func deriveMatchSetFromParticipantList(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var event models.EventRecord
	var participantList []models.EventParticipant = make([]models.EventParticipant, 0)
	var matchList []models.EventMatch = make([]models.EventMatch, 0)
	var participantCount uint
	var matchCount uint
	var err error

	log.Printf("[HANDLER]: loading participant list from workspace under %q into variable of type %T...", participantListRecordsKey, participantList)
	if err = space.Get(participantListRecordsKey, &participantList); err != nil {
		log.Printf("[HANDLER]: error loading participant list (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading event record from workspace under %q into variable of type %T...", eventRecordKey, event)
	if err = space.Get(eventRecordKey, &event); err != nil {
		log.Printf("[HANDLER]: error loading event record (%s)", err.Error())
		return err
	}

	log.Print("[HANDLER]: validating that there are enough participants for a bracket...")
	participantCount = uint(len(participantList))
	if participantCount < 2 {
		log.Printf("[HANDLER]: insufficient participants for a bracket (%d)", participantCount)
		return errors.New("insufficient number of participants for competition")
	}

	log.Print("[HANDLER]: validating that there are a reasonable number of participant for a bracket...")
	if participantCount > 256 {
		log.Printf("[HANDLER]: too many participants for a bracket (%d)", participantCount)
		return errors.New("too many participants for competition")
	}
	matchCount = (1 << bits.Len(participantCount-1)) - 1
	log.Printf("[HANDLER]: creating matchset for %d participants (%d matches needed)", participantCount, matchCount)

	log.Print("[HANDLER]: populating match list with unlinked match records...")
	for i := 0; i < int(matchCount); i++ {
		matchList = append(matchList, models.EventMatch{
			ID:               bson.NewObjectID(),
			TakesPlaceDuring: event.ID,
		})
	}

	log.Print("[HANDLER]: relating matches as a binary heap...")
	for i := 0; i < int(matchCount); i++ {
		awayIdx := 2*i + 1
		if awayIdx >= 0 && awayIdx < int(matchCount) {
			matchList[i].AwayParticipant = matchList[awayIdx].ID
			matchList[i].AwayRef = models.ParticipantFieldReferencesMatch
		}

		homeIdx := 2*i + 2
		if homeIdx >= 0 && homeIdx < int(matchCount) {
			matchList[i].HomeParticipant = matchList[homeIdx].ID
			matchList[i].HomeRef = models.ParticipantFieldReferencesMatch
		}
	}

	log.Print("[HANDLER]: seeding first round matches...")
	home := 0
	away := int(matchCount)
	for i := int(matchCount) / 2; i < int(matchCount); i++ {
		if home >= 0 && home < len(participantList) {
			matchList[i].HomeParticipant = participantList[home].ID
			matchList[i].HomeRef = models.ParticipantFieldReferencesPlayer
		} else {
			matchList[i].HomeParticipant = bson.NilObjectID
			matchList[i].HomeRef = models.ParticipantFieldReferencesBye
		}

		if away >= 0 && away < len(participantList) {
			matchList[i].AwayParticipant = participantList[away].ID
			matchList[i].AwayRef = models.ParticipantFieldReferencesPlayer
		} else {
			matchList[i].AwayParticipant = bson.NilObjectID
			matchList[i].AwayRef = models.ParticipantFieldReferencesBye
		}

		home++
		away--
	}

	log.Print("[HANDLER]: propogating BYE matches...")
	for i := int(matchCount) - 1; i >= 0; i-- {
		if matchList[i].AwayParticipant == bson.NilObjectID && matchList[i].HomeParticipant != bson.NilObjectID {
			matchList[i].Winner = matchList[i].HomeParticipant
		}
		if matchList[i].AwayParticipant != bson.NilObjectID && matchList[i].HomeParticipant == bson.NilObjectID {
			matchList[i].Winner = matchList[i].AwayParticipant
		}
		if matchList[i].HomeRef == models.ParticipantFieldReferencesMatch {
			feederIdx := 2*i + 2
			if feederIdx >= 0 && feederIdx < int(matchCount) && matchList[feederIdx].Winner != bson.NilObjectID {
				matchList[i].HomeParticipant = matchList[feederIdx].Winner
				matchList[i].HomeRef = models.ParticipantFieldReferencesPlayer
			}
		}
		if matchList[i].AwayRef == models.ParticipantFieldReferencesMatch {
			feederIdx := 2*i + 1
			if feederIdx >= 0 && feederIdx < int(matchCount) && matchList[feederIdx].Winner != bson.NilObjectID {
				matchList[i].AwayParticipant = matchList[feederIdx].Winner
				matchList[i].AwayRef = models.ParticipantFieldReferencesPlayer
			}
		}
	}

	log.Print("[HANDLER]: match set initialized")
	space.Set(matchListRecordKey, matchList)
	return nil
}

// Function `createEventRecord` inserts the event record within the workspace into the database
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func createEventRecord(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var sess *mongo.Session
	var event models.EventRecord
	var cfg *options.InsertOneOptionsBuilder
	var err error

	log.Printf("[HANDLER]: loading record data from workspace under the %q key into variable of type %T...", eventRecordKey, event)
	if err = space.Get(eventRecordKey, &event); err != nil {
		log.Printf("[HANDLER]: error loading event record data (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database operation settings...")
	if cfg, err = dbx.NewOptions(dbx.ValidateInsertedDocument(true)); err != nil {
		log.Printf("[HANDLER]: error configuration database operation (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database session from request context...")
	if sess, err = dbx.MongoFromContext(ctx); err != nil {
		log.Printf("[HANDLER]: error loading database session from request context (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: performing database insertion operation...")
	_, err = sess.Client().
		Database(models.EventQueryContext.Database).
		Collection(models.EventQueryContext.Collection).
		InsertOne(ctx, event, cfg)

	if err != nil {
		log.Printf("[HANDLER]: error during database insertion operation (%s)", err.Error())
	}

	return err
}

// Function `createParticipantRecord` inserts the participant record within the workspace into the database
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func createParticipantRecord(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var sess *mongo.Session
	var player models.EventParticipant
	var cfg *options.InsertOneOptionsBuilder
	var err error

	log.Printf("[HANDLER]: loading record data from workspace under the %q key into variable of type %T...", participantRecordKey, player)
	if err = space.Get(participantRecordKey, &player); err != nil {
		log.Printf("[HANDLER]: error loading participant record data (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database operation settings...")
	if cfg, err = dbx.NewOptions(dbx.ValidateInsertedDocument(true)); err != nil {
		log.Printf("[HANDLER]: error configuration database operation (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database session from request context...")
	if sess, err = dbx.MongoFromContext(ctx); err != nil {
		log.Printf("[HANDLER]: error loading database session from request context (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: performing database insertion operation...")
	_, err = sess.Client().
		Database(models.ParticipantQueryContext.Database).
		Collection(models.ParticipantQueryContext.Collection).
		InsertOne(ctx, player, cfg)

	if err != nil {
		log.Printf("[HANDLER]: error during database insertion operation (%s)", err.Error())
	}

	space.Set(participatIDResponseKey, models.ParticipantID{EID: player.ParticipatesIn.Hex(), PID: player.ID.Hex()})
	return err

}

// Function `createMatchSetRecord` creates the matches associated with an event
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func createMatchSetRecord(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var sess *mongo.Session
	var matches []models.EventMatch = make([]models.EventMatch, 0)
	var cfg *options.InsertManyOptionsBuilder
	var err error

	log.Printf("[HANDLER]: loading record data from workspace under the %q key into variable of type %T...", matchListRecordKey, matches)
	if err = space.Get(matchListRecordKey, &matches); err != nil {
		log.Printf("[HANDLER]: error loading participant record data (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database operation settings...")
	if cfg, err = dbx.NewOptions(dbx.ValidateInsertedDocuments(true), dbx.StopOnError(true)); err != nil {
		log.Printf("[HANDLER]: error configuration database operation (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database session from request context...")
	if sess, err = dbx.MongoFromContext(ctx); err != nil {
		log.Printf("[HANDLER]: error loading database session from request context (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: performing database insertion operation...")
	_, err = sess.Client().
		Database(models.MatchQueryContext.Database).
		Collection(models.MatchQueryContext.Collection).
		InsertMany(ctx, matches, cfg)

	if err != nil {
		log.Printf("[HANDLER]: error during insertion operation (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: inserted %d match records", len(matches))
	return nil
}

// Function `fetchEventRecordFromDatabaseByID` finds the event record within the workspace into the database
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func fetchEventRecordFromDatabaseByID(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var sess *mongo.Session
	var event models.EventRecord
	var req models.EventID
	var id bson.ObjectID
	var cfg *options.FindOneOptionsBuilder
	var err error

	log.Printf("[HANDLER]: loading event lookup request from workspace under %q key into variable of type %T...", eventLookupRequest, req)
	if err := space.Get(eventLookupRequest, &req); err != nil {
		log.Printf("[HANDLER]: error loading lookup request (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: interpreting ID presented in lookup request as an ObjectID...")
	if id, err = bson.ObjectIDFromHex(req.ID); err != nil {
		log.Printf("[HANDLER]: could not interpret provided ID as an ObjectID (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database operation settings...")
	if cfg, err = dbx.NewOptions(dbx.FindOneProjection(bson.E{Key: "participants", Value: 0}, bson.E{Key: "bracket", Value: 0})); err != nil {
		log.Printf("[HANDLER]: error configuration database operation (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database session from request context...")
	if sess, err = dbx.MongoFromContext(ctx); err != nil {
		log.Printf("[HANDLER]: error loading database session from request context (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: performing database lookup operation")
	filter := bson.D{{Key: "_id", Value: id}}
	err = sess.Client().
		Database(models.EventQueryContext.Database).
		Collection(models.EventQueryContext.Collection).
		FindOne(ctx, filter, cfg).
		Decode(&event)

	if err != nil {
		log.Printf("[HANDLER]: error during database lookup operation (%s)", err.Error())
		return err
	}

	space.Set(eventRecordKey, event)
	return nil

}

// Function `applyEventRecordModificationByID` updates the event record within the workspace into the database
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func applyEventRecordModificationByID(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var req models.UpdateEventRequest
	var cfg *options.UpdateOneOptionsBuilder
	var err error
	var update bson.D
	var which models.EventRecord
	var sess *mongo.Session
	var res *mongo.UpdateResult

	log.Printf("[HANDLER]: loading event update request from workspace under %q key into variable of type %T...", eventUpdateRequest, req)
	if err := space.Get(eventUpdateRequest, &req); err != nil {
		log.Printf("[HANDLER]: error loading update request (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading event record from workspace under %q key into variable of type %T...", eventRecordKey, which)
	if err := space.Get(eventRecordKey, &which); err != nil {
		log.Printf("[HANDLER]: error loading record (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database operation settings...")
	if cfg, err = dbx.NewOptions(dbx.ValidateUpdatedDocument(true), dbx.DoInsertOnNoMatchFound(false)); err != nil {
		log.Printf("[HANDLER]: error configuration database operation (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database session from request context...")
	if sess, err = dbx.MongoFromContext(ctx); err != nil {
		log.Printf("[HANDLER]: error loading database session from request context (%s)", err.Error())
		return err
	}

	if req.NewName != "" {
		update = append(update, bson.E{Key: "$set", Value: bson.D{{Key: "name", Value: req.NewName}}})
	}
	if req.NewGame != "" {
		update = append(update, bson.E{Key: "$set", Value: bson.D{{Key: "game", Value: req.NewGame}}})
	}
	if req.NewDescription != "" {
		update = append(update, bson.E{Key: "$set", Value: bson.D{{Key: "description", Value: req.NewDescription}}})
	}
	if req.NewStatus != "" {
		update = append(update, bson.E{Key: "$set", Value: bson.D{{Key: "status", Value: req.NewStatus}}})
	}
	log.Printf("[HANDLER]: configured update: %v", update)

	log.Print("[HANDLER]: running database update operation...")
	res, err = sess.Client().
		Database(models.EventQueryContext.Database).
		Collection(models.EventQueryContext.Collection).
		UpdateByID(
			ctx,
			which.ID,
			update,
			cfg,
		)

	if err != nil {
		log.Printf("[HANDLER]: error during database update operation (%s)", err.Error())
		return err
	}

	if res.ModifiedCount != 1 {
		log.Printf("[HANDLER]: incorrect number of documents updated (%d)", res.ModifiedCount)
		return errors.New("update not properly applied")
	}

	log.Printf("[HANDLER]: update applied to event (_id=%q)", which.ID.Hex())
	return nil
}

// Function `updateParticipantRecord` updates the specific participants record within the workspace into the database
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func updateParticipantRecord(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var modify models.CreateOrModifyParticipantRequest
	var cfg *options.UpdateOneOptionsBuilder
	var err error
	var which models.ParticipantID
	var playerID bson.ObjectID
	var sess *mongo.Session
	var res *mongo.UpdateResult

	log.Printf("[HANDLER]: loading event update request from workspace under %q key into variable of type %T...", participantCreationRequest, modify)
	if err := space.Get(participantCreationRequest, &modify); err != nil {
		log.Printf("[HANDLER]: error loading update request (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading event record from workspace under %q key into variable of type %T...", participantLookupRequest, which)
	if err := space.Get(participantLookupRequest, &which); err != nil {
		log.Printf("[HANDLER]: error loading record (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: interpreting ID presented in lookup request as an ObjectID...")
	if playerID, err = bson.ObjectIDFromHex(which.PID); err != nil {
		log.Printf("[HANDLER]: could not interpret provided ID as an ObjectID (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database operation settings...")
	if cfg, err = dbx.NewOptions(dbx.ValidateUpdatedDocument(true), dbx.DoInsertOnNoMatchFound(false)); err != nil {
		log.Printf("[HANDLER]: error configuration database operation (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database session from request context...")
	if sess, err = dbx.MongoFromContext(ctx); err != nil {
		log.Printf("[HANDLER]: error loading database session from request context (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: running database update operation...")
	res, err = sess.Client().
		Database(models.ParticipantQueryContext.Database).
		Collection(models.ParticipantQueryContext.Collection).
		UpdateByID(
			ctx,
			playerID,
			bson.D{{Key: "$set", Value: bson.D{{Key: "display_name", Value: modify.DisplayName}}}},
			cfg,
		)

	if err != nil {
		log.Printf("[HANDLER]: error during database update operation (%s)", err.Error())
		return err
	}

	if res.ModifiedCount != 1 {
		log.Printf("[HANDLER]: incorrect number of documents updated (found %d; update %d)", res.MatchedCount, res.ModifiedCount)
		return errors.New("update not properly applied")
	}

	space.Set(participatIDResponseKey, which)
	log.Printf("[HANDLER]: update applied to event (_id=%q)", which.EID)
	return nil
}

// Function `updateMatchWinnerByID` sets the winner field for a given match
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func updateMatchWinnerByID(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var modify models.DeclarMatchWinnerRequest
	var which models.MatchID
	var winner bson.ObjectID
	var matchID bson.ObjectID
	var eventID bson.ObjectID
	var cfg *options.UpdateOneOptionsBuilder
	var err error
	var sess *mongo.Session
	var res *mongo.UpdateResult

	log.Printf("[HANDLER]: loading match update request from workspace under %q key into variable of type %T...", matchDeclareWinnerRequest, modify)
	if err := space.Get(matchDeclareWinnerRequest, &modify); err != nil {
		log.Printf("[HANDLER]: error loading update request (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading match lookup request from workspace under %q key into variable of type %T...", matchLookupRequest, which)
	if err := space.Get(matchLookupRequest, &which); err != nil {
		log.Printf("[HANDLER]: error loading lookup request (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: interpreting ID presented in lookup request as an ObjectID...")
	if matchID, err = bson.ObjectIDFromHex(which.MID); err != nil {
		log.Printf("[HANDLER]: could not interpret provided ID as an ObjectID (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: interpreting ID presented in lookup request as an ObjectID...")
	if eventID, err = bson.ObjectIDFromHex(which.EID); err != nil {
		log.Printf("[HANDLER]: could not interpret provided ID as an ObjectID (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: interpreting ID presented in lookup request as an ObjectID...")
	if winner, err = bson.ObjectIDFromHex(modify.DeclareWinner); err != nil {
		log.Printf("[HANDLER]: could not interpret provided ID as an ObjectID (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database operation settings...")
	if cfg, err = dbx.NewOptions(dbx.ValidateUpdatedDocument(true), dbx.DoInsertOnNoMatchFound(false)); err != nil {
		log.Printf("[HANDLER]: error configuration database operation (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database session from request context...")
	if sess, err = dbx.MongoFromContext(ctx); err != nil {
		log.Printf("[HANDLER]: error loading database session from request context (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: running database update operation...")
	res, err = sess.Client().
		Database(models.MatchQueryContext.Database).
		Collection(models.MatchQueryContext.Collection).
		UpdateOne(
			ctx,
			bson.D{
				{Key: "_id", Value: matchID},
				{Key: "takes_place_during", Value: eventID},
				{Key: "$or", Value: bson.A{
					bson.D{{Key: "home", Value: winner}, {Key: "home_ref", Value: models.ParticipantFieldReferencesPlayer}},
					bson.D{{Key: "away", Value: winner}, {Key: "away_ref", Value: models.ParticipantFieldReferencesPlayer}},
				}},
			},
			bson.D{{Key: "$set", Value: bson.D{{Key: "winner", Value: winner}}}},
			cfg,
		)

	if err != nil {
		log.Printf("[HANDLER]: error during database update operation (%s)", err.Error())
		return err
	}

	if res.ModifiedCount != 1 {
		log.Printf("[HANDLER]: incorrect number of documents updated (found %d; update %d)", res.MatchedCount, res.ModifiedCount)
		return errors.New("update not properly applied")
	}

	log.Printf("[HANDLER]: declared winner for match (_id=%s)", matchID.Hex())
	space.Set(matchIDResponseKey, which)
	return nil
}

// Function `removeParticipantRecord` removes the specific participants record within the workspace into the database
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func removeParticipantRecord(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var err error
	var whichParticipant models.ParticipantID
	var playerID bson.ObjectID
	var eventID bson.ObjectID
	var sess *mongo.Session
	var res *mongo.DeleteResult

	log.Printf("[HANDLER]: loading participant lookup request from workspace under %q key into variable of type %T...", participantLookupRequest, whichParticipant)
	if err := space.Get(participantLookupRequest, &whichParticipant); err != nil {
		log.Printf("[HANDLER]: error loading record (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: interpreting ID presented in lookup request as an ObjectID...")
	if playerID, err = bson.ObjectIDFromHex(whichParticipant.PID); err != nil {
		log.Printf("[HANDLER]: could not interpret provided ID as an ObjectID (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: interpreting ID presented in lookup request as an ObjectID...")
	if eventID, err = bson.ObjectIDFromHex(whichParticipant.EID); err != nil {
		log.Printf("[HANDLER]: could not interpret provided ID as an ObjectID (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database session from request context...")
	if sess, err = dbx.MongoFromContext(ctx); err != nil {
		log.Printf("[HANDLER]: error loading database session from request context (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: running database removal operation...")
	res, err = sess.Client().
		Database(models.ParticipantQueryContext.Database).
		Collection(models.ParticipantQueryContext.Collection).
		DeleteOne(
			ctx,
			bson.D{{Key: "_id", Value: playerID}, {Key: "participates_in", Value: eventID}},
		)

	if err != nil {
		log.Printf("[HANDLER]: error during database delete operation (%s)", err.Error())
		return err
	}

	if res.DeletedCount != 1 {
		log.Printf("[HANDLER]: incorrect number of documents deleted (removed %d)", res.DeletedCount)
		return errors.New("delete not properly applied")
	}

	space.Set(participatIDResponseKey, whichParticipant)
	return nil
}

// Function `fetchParticipantsFromDatabaseByEventID` fetches the event participants record within the workspace into the database
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func fetchParticipantsFromDatabaseByEventID(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var sess *mongo.Session
	var cur *mongo.Cursor
	var participants []models.EventParticipant = make([]models.EventParticipant, 0)
	var req models.EventID
	var id bson.ObjectID
	var err error

	log.Printf("[HANDLER]: loading event lookup request from workspace under %q key into variable of type %T...", eventLookupRequest, req)
	if err := space.Get(eventLookupRequest, &req); err != nil {
		log.Printf("[HANDLER]: error loading lookup request (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: interpreting ID presented in lookup request as an ObjectID...")
	if id, err = bson.ObjectIDFromHex(req.ID); err != nil {
		log.Printf("[HANDLER]: could not interpret provided ID as an ObjectID (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database session from request context...")
	if sess, err = dbx.MongoFromContext(ctx); err != nil {
		log.Printf("[HANDLER]: error loading database session from request context (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: performing database lookup operation")
	filter := bson.D{{Key: "participates_in", Value: id}}
	cur, err = sess.Client().
		Database(models.ParticipantQueryContext.Database).
		Collection(models.ParticipantQueryContext.Collection).
		Find(ctx, filter)

	if err != nil {
		log.Printf("[HANDLER]: error during database lookup operation (%s)", err.Error())
		return err
	}

	if err = cur.All(ctx, &participants); err != nil {
		log.Printf("[HANDLER]: error during database lookup operation (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: found %d participants", len(participants))
	space.Set(participantListRecordsKey, participants)
	return nil
}

// Function `fetchParticipantFromDatabaseByPlayerID` finds the participant with the given ID
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func fetchParticipantFromDatabaseByPlayerID(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var sess *mongo.Session
	var reqPlayer models.ParticipantID
	var eventID bson.ObjectID
	var playerID bson.ObjectID
	var player models.EventParticipant
	var err error

	log.Printf("[HANDLER]: loading participant lookup request from workspace under %q key into variable of type %T...", participantLookupRequest, reqPlayer)
	if err := space.Get(participantLookupRequest, &reqPlayer); err != nil {
		log.Printf("[HANDLER]: error loading lookup request (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: interpreting ID presented in lookup request as an ObjectID...")
	if playerID, err = bson.ObjectIDFromHex(reqPlayer.PID); err != nil {
		log.Printf("[HANDLER]: could not interpret provided ID as an ObjectID (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: interpreting ID presented in lookup request as an ObjectID...")
	if eventID, err = bson.ObjectIDFromHex(reqPlayer.EID); err != nil {
		log.Printf("[HANDLER]: could not interpret provided ID as an ObjectID (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database session from request context...")
	if sess, err = dbx.MongoFromContext(ctx); err != nil {
		log.Printf("[HANDLER]: error loading database session from request context (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: performing database lookup operation")
	filter := bson.D{{Key: "_id", Value: playerID}, {Key: "participates_in", Value: eventID}}
	err = sess.Client().
		Database(models.ParticipantQueryContext.Database).
		Collection(models.ParticipantQueryContext.Collection).
		FindOne(ctx, filter).
		Decode(&player)

	if err != nil {
		log.Printf("[HANDLER]: error during database lookup operation (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: found participant (%v)", player)
	space.Set(participantRecordKey, player)
	return nil

}

// Function `fetchMatchSetFromDatabaseByEventID` finds the match set associated with the given event ID
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func fetchMatchSetFromDatabaseByEventID(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var sess *mongo.Session
	var cur *mongo.Cursor
	var matches []models.EventMatch = make([]models.EventMatch, 0)
	var req models.EventID
	var id bson.ObjectID
	var err error

	log.Printf("[HANDLER]: loading event lookup request from workspace under %q key into variable of type %T...", eventLookupRequest, req)
	if err := space.Get(eventLookupRequest, &req); err != nil {
		log.Printf("[HANDLER]: error loading lookup request (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: interpreting ID presented in lookup request as an ObjectID...")
	if id, err = bson.ObjectIDFromHex(req.ID); err != nil {
		log.Printf("[HANDLER]: could not interpret provided ID as an ObjectID (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database session from request context...")
	if sess, err = dbx.MongoFromContext(ctx); err != nil {
		log.Printf("[HANDLER]: error loading database session from request context (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: performing database lookup operation")
	filter := bson.D{{Key: "takes_place_during", Value: id}}
	cur, err = sess.Client().
		Database(models.MatchQueryContext.Database).
		Collection(models.MatchQueryContext.Collection).
		Find(ctx, filter)

	if err != nil {
		log.Printf("[HANDLER]: error during database lookup operation (%s)", err.Error())
		return err
	}

	if err = cur.All(ctx, &matches); err != nil {
		log.Printf("[HANDLER]: error during database lookup operation (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: found %d matches", len(matches))
	space.Set(matchListRecordKey, matches)
	return nil
}

// Function `fetchMatchFromDatabaseByID` fetches the match associated with the given ID
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func fetchMatchFromDatabaseByID(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var sess *mongo.Session
	var req models.MatchID
	var eventID bson.ObjectID
	var matchID bson.ObjectID
	var match models.EventMatch
	var err error

	log.Printf("[HANDLER]: loading match lookup request from workspace under %q key into variable of type %T...", matchLookupRequest, req)
	if err := space.Get(matchLookupRequest, &req); err != nil {
		log.Printf("[HANDLER]: error loading lookup request (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: interpreting ID presented in lookup request as an ObjectID...")
	if eventID, err = bson.ObjectIDFromHex(req.EID); err != nil {
		log.Printf("[HANDLER]: could not interpret provided ID as an ObjectID (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: interpreting ID presented in lookup request as an ObjectID...")
	if matchID, err = bson.ObjectIDFromHex(req.MID); err != nil {
		log.Printf("[HANDLER]: could not interpret provided ID as an ObjectID (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database session from request context...")
	if sess, err = dbx.MongoFromContext(ctx); err != nil {
		log.Printf("[HANDLER]: error loading database session from request context (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: performing database lookup operation")
	filter := bson.D{{Key: "_id", Value: matchID}, {Key: "takes_place_during", Value: eventID}}
	err = sess.Client().
		Database(models.MatchQueryContext.Database).
		Collection(models.MatchQueryContext.Collection).
		FindOne(ctx, filter).
		Decode(&match)

	if err != nil {
		log.Printf("[HANDLER]: error during database lookup operation (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: found match (%v)", match)
	space.Set(matchRecordKey, match)
	return nil
}

// Function `updateAwayParticipantIfAvailable` update the ID references in a matches away participant field if the reference points to another match with a "winner" existing
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func updateAwayParticipantIfAvailable(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var sess *mongo.Session
	var match models.EventMatch
	var feeder models.EventMatch
	var err error
	var res *mongo.UpdateResult

	log.Printf("[HANDLER]: loading match record from workspace under %q key into variable of type %T...", matchRecordKey, match)
	if err := space.Get(matchRecordKey, &match); err != nil {
		log.Printf("[HANDLER]: error loading match record (%s)", err.Error())
		return err
	}

	if match.AwayRef != models.ParticipantFieldReferencesMatch {
		log.Printf("[HANDLER]: target match record is not referencing another match for field %q", "away_ref")
		return errors.New("cannot resolve away participant when not referencing another match")
	}

	log.Printf("[HANDLER]: loading database session from request context...")
	if sess, err = dbx.MongoFromContext(ctx); err != nil {
		log.Printf("[HANDLER]: error loading database session from request context (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: performing database lookup operation (feeder match)")
	filter := bson.D{
		{Key: "_id", Value: match.AwayParticipant},
		{Key: "takes_place_during", Value: match.TakesPlaceDuring},
		{Key: "winner", Value: bson.D{{Key: "$exists", Value: true}}},
	}
	err = sess.Client().
		Database(models.MatchQueryContext.Database).
		Collection(models.MatchQueryContext.Collection).
		FindOne(ctx, filter).
		Decode(&feeder)

	if err != nil {
		log.Printf("[HANDLER]: error during database lookup operation (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: performing database update operation (target match)")
	res, err = sess.Client().
		Database(models.MatchQueryContext.Database).
		Collection(models.MatchQueryContext.Collection).
		UpdateByID(
			ctx,
			match.ID,
			bson.D{{Key: "$set", Value: bson.D{{Key: "away", Value: feeder.Winner}, {Key: "away_ref", Value: models.ParticipantFieldReferencesPlayer}}}},
		)

	if err != nil {
		log.Printf("[HANDLER]: error during database update operation (%s)", err.Error())
		return err
	}

	if res.ModifiedCount != 1 {
		log.Printf("[HANDLER]: incorrect number of documents updated (found %d; update %d)", res.MatchedCount, res.ModifiedCount)
		return errors.New("update not properly applied")
	}

	space.Set(matchIDResponseKey, models.MatchID{EID: match.TakesPlaceDuring.Hex(), MID: match.ID.Hex()})
	log.Printf("[HANDLER]: update applied to event (_id=%q)", match.ID.Hex())
	return nil
}

// Function `updateHomeParticipantIfAvailable` update the ID references in a matches home participant field if the reference points to another match with a "winner" existing
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func updateHomeParticipantIfAvailable(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var sess *mongo.Session
	var match models.EventMatch
	var feeder models.EventMatch
	var err error
	var res *mongo.UpdateResult

	log.Printf("[HANDLER]: loading match record from workspace under %q key into variable of type %T...", matchRecordKey, match)
	if err := space.Get(matchRecordKey, &match); err != nil {
		log.Printf("[HANDLER]: error loading match record (%s)", err.Error())
		return err
	}

	if match.HomeRef != models.ParticipantFieldReferencesMatch {
		log.Printf("[HANDLER]: target match record is not referencing another match for field %q", "home_ref")
		return errors.New("cannot resolve home participant when not referencing another match")
	}

	log.Printf("[HANDLER]: loading database session from request context...")
	if sess, err = dbx.MongoFromContext(ctx); err != nil {
		log.Printf("[HANDLER]: error loading database session from request context (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: performing database lookup operation (feeder match)")
	filter := bson.D{
		{Key: "_id", Value: match.HomeParticipant},
		{Key: "takes_place_during", Value: match.TakesPlaceDuring},
		{Key: "winner", Value: bson.D{{Key: "$exists", Value: true}}},
	}
	err = sess.Client().
		Database(models.MatchQueryContext.Database).
		Collection(models.MatchQueryContext.Collection).
		FindOne(ctx, filter).
		Decode(&feeder)

	if err != nil {
		log.Printf("[HANDLER]: error during database lookup operation (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: performing database update operation (target match)")
	res, err = sess.Client().
		Database(models.MatchQueryContext.Database).
		Collection(models.MatchQueryContext.Collection).
		UpdateByID(
			ctx,
			match.ID,
			bson.D{{Key: "$set", Value: bson.D{{Key: "home", Value: feeder.Winner}, {Key: "home_ref", Value: models.ParticipantFieldReferencesPlayer}}}},
		)

	if err != nil {
		log.Printf("[HANDLER]: error during database update operation (%s)", err.Error())
		return err
	}

	if res.ModifiedCount != 1 {
		log.Printf("[HANDLER]: incorrect number of documents updated (found %d; update %d)", res.MatchedCount, res.ModifiedCount)
		return errors.New("update not properly applied")
	}

	space.Set(matchIDResponseKey, models.MatchID{EID: match.TakesPlaceDuring.Hex(), MID: match.ID.Hex()})
	log.Printf("[HANDLER]: update applied to event (_id=%q)", match.ID.Hex())
	return nil
}

// Function `removeEventRecordByID` deletes the event participants record within the workspace into the database
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func removeEventRecordByID(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var which models.EventRecord
	var sess *mongo.Session
	var res *mongo.DeleteResult
	var err error

	log.Printf("[HANDLER]: loading event record from workspace under %q key into variable of type %T...", eventRecordKey, which)
	if err := space.Get(eventRecordKey, &which); err != nil {
		log.Printf("[HANDLER]: error loading record (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database session from request context...")
	if sess, err = dbx.MongoFromContext(ctx); err != nil {
		log.Printf("[HANDLER]: error loading database session from request context (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: running database delete...")
	res, err = sess.Client().
		Database(models.EventQueryContext.Database).
		Collection(models.EventQueryContext.Collection).
		DeleteOne(ctx, bson.D{{Key: "_id", Value: which.ID}})

	if err != nil {
		log.Printf("[HANDLER]: error during database delete operation (%s)", err.Error())
		return err
	}

	if res.DeletedCount != 1 {
		log.Printf("[HANDLER]: incorrect number of documents updated (%d)", res.DeletedCount)
		return errors.New("delete not properly applied")
	}

	log.Printf("[HANDLER]: delete applied to event (_id=%q)", which.ID.Hex())
	return nil

}

// Function `populateEventIDResponse` populates the fields for identifying an event by its ID
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func populateEventIDResponse(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var record models.EventRecord
	var response models.EventID
	var err error

	log.Printf("[HANDLER]: loading record data from workspace under the %q key into variable of type %T...", eventRecordKey, record)
	if err = space.Get(eventRecordKey, &record); err != nil {
		log.Printf("[HANDLER]: error loading event record data (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: populating response fields...")
	response.ID = record.ID.Hex()

	log.Printf("[HANDLER]: saved response data to workspace under the %q key", eventIDResponseKey)
	space.Set(eventIDResponseKey, response)
	return nil
}
