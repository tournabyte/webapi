package models

/*
 * File: pkg/models/events.go
 *
 * Purpose: data models for event management
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"github.com/tournabyte/webapi/pkg/dbx"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Variables storing query context associated with event management operations
var (
	EventQueryContext = dbx.NewQueryContext(`tournabyte`, `events`)
)

// Constants storing status values for an event's status field
const (
	StatusPlanned    = "PLANNED"
	StatusInProgress = "IN-PROGRESS"
	StatusConcluded  = "CONCLUDED"
)

// Type `CreateEventRequest` represents the request body format for the create event endpoint
//
// Fields:
//   - Name: the name of the event (to show instead of an ID)
//   - Game: the game the event is focused around
//   - Description: the description of the event
type CreateEventRequest struct {
	Name        string `json:"name" binding:"required,min=4,max=128"`
	Game        string `json:"game" binding:"required,min=4,max=128"`
	Description string `json:"description" binding:"max=1024"`
}

// Type `UpdateEventRequest` represents the request body format for the update event endpoint
//
// Fields:
//   - NewName: the new name of the event
//   - NewGame: the new game of the event
//   - NewDescription: the new description of the event
//   - NewStatus: the new status of the event
type UpdateEventRequest struct {
	NewName        string `json:"name" binding:"max=128"`
	NewGame        string `json:"game" binding:"max=128"`
	NewDescription string `json:"description" binding:"max=1024"`
	NewStatus      string `json:"status" binding:"max=128"`
}

// Type `EventParticipant` represents the request body format for the update event participant endpoint
//
// Fields:
//   - NewParticipants: the list of new participants
type EventParticipants struct {
	List []string `json:"participants" bson:"participants"`
}

// Type `EventID` represents a response to an successful event (created/updated/deleted) endpoint usage
//
// Fields:
//   - ID: the event created/updated/deleted
type EventID struct {
	ID string `json:"id" uri:"eventid" binding:"required,mongodb"`
}

// Type `EventDetailsResponse` represents a subset of the database record for an event delivered when looking up a specific event
//
// Fields:
//   - ID: the unique ID for this event
//   - Host: the user ID that created this event
//   - Status: the status of this event
//   - Name: name of the event
//   - Game: the game the event is focused around
//   - Description: description of the event
type EventDetailsResponse struct {
	ID          string `json:"id"`
	Host        string `json:"hostBy"`
	Status      string `json:"status"`
	Name        string `json:"name"`
	Game        string `json:"game"`
	Description string `json:"description"`
}

// Type `EventRecord` represents a database record for an event
//
// Fields:
//   - ID: the unique ID for this event
//   - Host: the user ID that created this event
//   - Status: the status of this event (i.e. planned, in-progress, concluded)
//   - Name: the name of the event
//   - Game: the game the event is focused around
//   - Description: the description of the event
//   - Participants: the participant list for this event
//   - Bracket: the match bracket for this event
type EventRecord struct {
	ID           bson.ObjectID                     `bson:"_id"`
	Host         bson.ObjectID                     `bson:"host"`
	Status       string                            `bson:"status"`
	Name         string                            `bson:"name"`
	Game         string                            `bson:"game"`
	Description  string                            `bson:"description"`
	Participants []string                          `bson:"participants"`
	Bracket      map[bson.ObjectID][]bson.ObjectID `bson:"bracket"`
}
