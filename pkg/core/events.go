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
	"log"

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
	eventCreationRequest = "createEventRequest"
	eventRecordKey       = "eventRecord"
	eventIDResponseKey   = "eventIDResponse"
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
	space.Set(authTokenSigningAlgorithm, srv.opts.Serve.Sessions.Algorithm)
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
	var record models.EventRecord
	var err error

	log.Printf("[HANDLER]: loading request data from workspace under %q into variable of type %T...", eventCreationRequest, req)
	if err = space.Get(eventCreationRequest, &req); err != nil {
		log.Printf("[HANDLER]: error loading request data (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: populating event record...")
	record.ID = bson.NewObjectID()
	record.Name = req.Name
	record.Game = req.Game
	record.Description = req.Description
	record.Participants = make([]string, 0)
	record.Bracket = make(map[bson.ObjectID][]bson.ObjectID)

	log.Printf("[HANDLER]: saved event record to workspace under the %q key", eventRecordKey)
	space.Set(eventRecordKey, record)
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
	if err = space.Get(eventCreationRequest, &event); err != nil {
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

// Function `populateEventIDResponse` uses the request body within the workspace to initialize an event record
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func populateEventIDResponse(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var record models.EventRecord
	var response models.EventIDResponse
	var err error

	log.Printf("[HANDLER]: loading record data from workspace under the %q key into variable of type %T...", eventRecordKey, record)
	if err = space.Get(eventCreationRequest, &record); err != nil {
		log.Printf("[HANDLER]: error loading event record data (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: populating response fields...")
	response.ID = record.ID.Hex()

	log.Printf("[HANDLER]: saved response data to workspace under the %q key", eventIDResponseKey)
	space.Set(eventIDResponseKey, response)
	return nil
}
