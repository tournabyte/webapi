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
	eventCreationRequest           = "createEventRequest"
	eventLookupRequest             = "lookupEventRequest"
	eventUpdateRequest             = "updateEventRequest"
	eventUpdateParticipantsRequest = "updateParticipantsRequest"
	eventRecordKey                 = "eventRecord"
	eventIDResponseKey             = "eventIDResponse"
	eventDetailsResponseKey        = "eventDetailsResponse"
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

func (srv *tournabyteAPIService) initEventLookupWorkspace(ctx *gin.Context) *handlerutil.HandlerWorkspace {
	space := handlerutil.DefaultWorkspace()
	binds := handlerutil.BindingsFromRequestContext(ctx, handlerutil.ShouldHaveURIValues|handlerutil.ShouldHaveHeaders)

	space.Set(handlerutil.RequestBindings, binds)
	space.Set(authTokenOptionsKey, srv.getTokenConfig())
	space.Set(models.ValidatorObjectKey, srv.validationFunc)
	log.Printf("[HANDLER]: setup request bindings")
	return &space
}

func (srv *tournabyteAPIService) initEventUpdateWorkspace(ctx *gin.Context) *handlerutil.HandlerWorkspace {
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
	out4 := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchEventRecordFromDatabaseByID, out3)
	pipelineOutput := handlerutil.Stage(pipelineCtx, pipelineCancel, populateEventDetailsResponse, out4)

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

// Function `eventParticipantSetterPipeline` initializes a handling pipeline for populating an event's participant list
//
// Parameters:
//   - ctx: the parent context to control the created pipeline
//
// Returns:
//   - `context.Context`: the context controlling the created pipeline (derived from the given context.Context)
//   - `context.CancelCauseFunc`: the cancellation function controlling pipeline cancellation
//   - `chan<- *handlerutil.HandlerWorkspace`: the input channel for the pipeline (send-only)
//   - `<-chan *handlerutil.HandlerWorkspace`: the output channel for the pipeline (read-only)
func eventParticipantSetterPipeline(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
	pipelineCtx, pipelineCancel := context.WithCancelCause(ctx)
	pipelineInput := make(chan *handlerutil.HandlerWorkspace)

	out1 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindAccessTokenFromHeader, pipelineInput)
	out2 := handlerutil.Stage(pipelineCtx, pipelineCancel, validateAccessToken, out1)
	out3 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindEventLookupRequestFromURI, out2)
	out4 := handlerutil.Stage(pipelineCtx, pipelineCancel, fetchEventRecordFromDatabaseByID, out3)
	out5 := handlerutil.Stage(pipelineCtx, pipelineCancel, verifyEventOwnership, out4)
	out6 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindNewParticipantListFromBody, out5)
	out7 := handlerutil.Stage(pipelineCtx, pipelineCancel, patchEventParticipantList, out6)
	pipelineOutput := handlerutil.Stage(pipelineCtx, pipelineCancel, populateEventIDResponse, out7)

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

func bindNewParticipantListFromBody(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var modify models.UpdateParticipantListRequest
	var bindings handlerutil.Bindings

	log.Printf("[HANDLER]: loading request bindings from workspace...")
	if err := space.Get(handlerutil.RequestBindings, &bindings); err != nil {
		log.Printf("[HANDLER]: error loading request bindings from workspace (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: binding request body to variable of type %T...", modify)
	if err := bindings.BindURI(&modify); err != nil {
		log.Printf("[HANDLER]: error binding request body (%s)", err.Error())
		return err
	}

	space.Set(eventUpdateParticipantsRequest, modify)
	log.Printf("[HANDLER]: saved request body as variable of type %T within workspace under key %q", modify, eventUpdateParticipantsRequest)
	return nil
}

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

func patchEventParticipantList(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var req models.UpdateParticipantListRequest
	var cfg *options.UpdateOneOptionsBuilder
	var err error
	var which models.EventRecord
	var sess *mongo.Session
	var res *mongo.UpdateResult

	log.Printf("[HANDLER]: loading event update request from workspace under %q key into variable of type %T...", eventUpdateRequest, req)
	if err := space.Get(eventUpdateParticipantsRequest, &req); err != nil {
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

	log.Print("[HANDLER]: running database update operation...")
	res, err = sess.Client().
		Database(models.EventQueryContext.Database).
		Collection(models.EventQueryContext.Collection).
		UpdateByID(
			ctx,
			which.ID,
			bson.D{{Key: "$set", Value: bson.D{{Key: "participants", Value: req.NewParticipants}}}},
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

func populateEventDetailsResponse(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var record models.EventRecord
	var response models.EventDetailsResponse
	var err error

	log.Printf("[HANDLER]: loading record data from workspace under the %q key into variable of type %T...", eventRecordKey, record)
	if err = space.Get(eventRecordKey, &record); err != nil {
		log.Printf("[HANDLER]: error loading event record data (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: populating response fields")
	response.ID = record.ID.Hex()
	response.Host = record.Host.Hex()
	response.Name = record.Name
	response.Game = record.Game
	response.Description = record.Description
	response.Status = record.Status

	log.Printf("[HANDLER]: saved response data to workspace under the %q key", eventDetailsResponseKey)
	space.Set(eventDetailsResponseKey, response)
	return nil
}
