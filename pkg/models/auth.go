package models

/*
 * File: pkg/models/auth.go
 *
 * Purpose: data modeling for authentication/authorization logic
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/tournabyte/webapi/pkg/dbx"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Variables storing query context associated with user account operations
var (
	UserAccountQueryContext = dbx.NewQueryContext(`tournabyte`, `users`)
	UserSessionQueryContext = dbx.NewQueryContext(`tournabyte`, `sessions`)
)

const (
	ValidatorObjectKey = "validatorObject"
)

// Type `AuthenticatedUser` represents the response structure for successfully authenticating as a user
//
// Fields:
//   - ID: the new user ID
//   - AccessToken: the JSON web token used for authorization to access protected resources
//   - RefreshToken: the token used for obtaining another access token once the current one has expired
type AuthenticatedUser struct {
	ID           string `json:"id"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// Type `AuthorizationTokenClaims` represents the custom claims present in a JWT produced by the Tournabyte API
//
// Fields:
//   - Owner: private claim expected to be the userID of the account this token was issued to
type AuthorizationTokenClaims struct {
	Me string `json:"whoami" validate:"required,mongodb"`
}

// Type `AuthorizationHeaderContent` represents a key:value pair specifically for the HTTP Authorization header
//
// Fields:
//   - Token: JWT value given as part of the HTTP authorization header
type AuthorizationHeaderContent struct {
	Token string `header:"Authorization" binding:"required"`
}

// Type `NewUserRequest` represents the minimum details required to create a user account
//
// Fields:
//   - Email: the email of the new user
//   - Password: the password of the new user (to be cryptographically secured before storing)
type AuthenticationRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// Type `SessionRefreshRequest` represents the details required to exchange a refresh token for a new access/refresh token pair
//
// Fields:
//   - RefreshToken: the current refresh token to exchange
type SessionID struct {
	RefreshToken string `json:"refresh" uri:"sessionid" binding:"required"`
}

// Type `UserAccount` represesnts a user within the Tournabyte platform (this is not the same as a player or team)
//
// Fields:
//   - ID: the user account id
//   - LoginEmail: the email associated with this user's login details
//   - PasswordHash: the hashed password associated with this user's login details
//   - Metadata: account document metadata
type UserAccount struct {
	ID           bson.ObjectID        `bson:"_id"`
	LoginEmail   string               `bson:"login_email"`
	PasswordHash string               `bson:"password_hash"`
	Metadata     dbx.DocumentMetadata `bson:"metadata"`
}

// Type `UserSession` represents the server-side session details needed to validate refresh tokens and reissue access tokens
//
// Fields:
//   - ID: the session ID
//   - NotValidBefore: the timestamp when the session refresh can be used
//   - NotValidAfter: the timestamp stopping when the session refresh can be used
//   - Authorizes: the user ID this session authorizes
//   - Rotated: indicates whether this session has already been used to rotate access tokens
type UserSession struct {
	ID             string        `bson:"_id"`
	NotValidBefore time.Time     `bson:"not_valid_before"`
	NotValidAfter  time.Time     `bson:"not_valid_after"`
	Authorizes     bson.ObjectID `bson:"authorizes"`
	Rotated        bool          `bson:"rotated"`
}

// Type `TokenOptions` groups the information needed to create and verify access tokens
//
// Fields:
//   - Issuer: the named token issuer
//   - Subject: the subject of the access token
//   - ExpiresIn: duration created access token should remain valid
//   - Signer: token signing tool
type TokenOptions struct {
	Issuer    string
	Subject   string
	ExpiresIn time.Duration
	Signer    jose.Signer
}

// Type `SessionOptions` groups the information needed to create and verify refresh tokens
//
// Fields:
//   - ExpiresIn: duration created refresh token should remain valid
type SessionOptions struct {
	ExpiresIn time.Duration
}
