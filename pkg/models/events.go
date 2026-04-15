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
	EventQueryContext       = dbx.NewQueryContext(`tournabyte`, `events`)
	ParticipantQueryContext = dbx.NewQueryContext(`tournabyte`, `participants`)
	MatchQueryContext       = dbx.NewQueryContext(`tournabyte`, `matches`)
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

// Type `EventID` represents a response to an successful event (created/updated/deleted) endpoint usage
//
// Fields:
//   - ID: the event created/updated/deleted
type EventID struct {
	ID string `json:"id" uri:"eventid" binding:"required,mongodb"`
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
type EventRecord struct {
	ID          bson.ObjectID `json:"id" bson:"_id"`
	Host        bson.ObjectID `json:"hostedBy" bson:"host"`
	Status      string        `json:"status" bson:"status"`
	Name        string        `json:"name" bson:"name"`
	Game        string        `json:"game" bson:"game"`
	Description string        `json:"description" bson:"description"`
}

// Type `CreateOrModifyParticipantRequest` represents the request body for a new participant
//
// Fields:
//   - DisplayName: the name to use for the participant's display name
type CreateOrModifyParticipantRequest struct {
	DisplayName string `json:"name" binding:"required,min=4,max=64"`
}

// Type `ParticipantLookupRequest` represents the request URI for looking up a participant
//
// Fields:
//   - EID: event ID participant is associated with
//   - PID: participant identifier
type ParticipantID struct {
	EID string `uri:"eventid" binding:"required,mongodb" json:"eventid"`
	PID string `uri:"playerid" binding:"required,mongodb" json:"playerid"`
}

// Type `EventParticipant` represent a record of a participant in an event
//
// Fields:
//   - ID: the unique ID of the participant
//   - DisplayName: the name shown in the UI for this participant
//   - ParticipatesIn: references the ID of the event this participant takes part in
type EventParticipant struct {
	ID             bson.ObjectID `json:"id" bson:"_id"`
	DisplayName    string        `json:"displayName" bson:"display_name"`
	ParticipatesIn bson.ObjectID `json:"participatesIn" bson:"participates_in"`
}

// Type `EventMatch` represents a match record associated with an event
//
// Fields:
//   - ID: the ID of the match
//   - AwayParticipant: the ID of the away participant or the match it is sourced from
//   - AwayRef: states whether `AwayParticipant` refers to a participant ID or a match ID
//   - HomeParticipant: the ID of the home participant or the match it is sourced from
//   - HomeRef: states whether `HomeParticipant` refers to a participant ID or a match ID
//   - Winner: the declared winner of the match (used by match referencing this match to populate participants)
//   - TakesPlaceDuring: references the ObjectID of the event this match is associated with
type EventMatch struct {
	ID               bson.ObjectID `json:"id" bson:"_id"`
	AwayParticipant  bson.ObjectID `json:"away" bson:"away"`
	AwayRef          string        `json:"-" bson:"away_ref"`
	HomeParticipant  bson.ObjectID `json:"home" bson:"home"`
	HomeRef          string        `json:"-" bson:"home_ref"`
	Winner           bson.ObjectID `json:"winner,omitempty" bson:"winner,omitempty"`
	TakesPlaceDuring bson.ObjectID `json:"takesPlaceDuring" bson:"takes_place_during"`
}
