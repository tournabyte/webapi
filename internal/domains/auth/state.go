package auth

/*
 * File: internal/domains/auth/state.go
 *
 * Purpose: data modeling for HTTP handler state variables
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"github.com/tournabyte/webapi/internal/utils"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Type `RegistrationHandlerState` represents the state that is passed along between steps of the user registration handler
//
// Fields:
//   - bindings: binder functions for other fields
//   - request: the request structure to bind
//   - response: the response to fill out
//   - record: the record to interface with the database
type AuthenticationHandlerState struct {
	bindings  utils.Bindings
	request   AuthenticationRequest
	response  AuthenticatedUser
	record    UserAccount
	mongosess *mongo.Session
}
