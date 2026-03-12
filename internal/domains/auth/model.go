package auth

/*
 * File: internal/domains/auth/model.go
 *
 * Purpose: data modeling for authentication and authorization
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"errors"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/tournabyte/webapi/internal/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrInvalidUserRegistration = errors.New("user cannot register due invalid input data")
	ErrInvalidSession          = errors.New("session could not be created")
	ErrWrongCredentials        = errors.New("provided credentials do not match any stored credentials")
	ErrNilAuthorizationToken   = errors.New("cannot decode claims of a missing token")
)

const (
	AuthorizationClaims = "AuthorizedID"
)

var (
	UserAccountQueryContext = utils.NewQueryContext(`tournabyte`, `users`)
	UserSessionQueryContext = utils.NewQueryContext(`tournabyte`, `sessions`)
)

// Type `AuthorizationTokenClaims` represents the custom claims present in a JWT produced by the Tournabyte API
//
// Fields:
//   - Owner: private claim expected to be the userID of the account this token was issued to
type AuthorizationTokenClaims struct {
	Owner string `json:"owner" validate:"required,mongodb"`
}

// Type `UserAccount` represesnts a user within the Tournabyte platform (this is not the same as a player or team)
//
// Fields:
//   - ID: the user account id
//   - LoginEmail: the email associated with this user's login details
//   - PasswordHash: the hashed password associated with this user's login details
//   - Metadata: account document metadata
type UserAccount struct {
	ID           bson.ObjectID          `bson:"_id"`
	LoginEmail   string                 `bson:"login_email"`
	PasswordHash string                 `bson:"password_hash"`
	Metadata     utils.DocumentMetadata `bson:"metadata"`
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
