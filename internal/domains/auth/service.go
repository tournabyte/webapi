package auth

/*
 * File: internal/domains/auth/service.go
 *
 * Purpose: implementing authentication and authorization business logic as functions
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"context"
	"crypto/rand"
	"log/slog"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/carlmjohnson/truthy"
	"github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/go-playground/validator/v10"
	"github.com/tournabyte/webapi/internal/domains/user"
	"github.com/tournabyte/webapi/internal/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Function `Register` instantiates the data records for a new user and prepares it for insertion to the database
//
// Parameters:
//   - ctx: the request context for managing the request lifetime
//   - src: the data source for the new account date record
//
// Returns:
//   - `utils.MongoOperationFunc`: a closure capable of performing the database insertion operation with the newly instantiated account data records
func Register(ctx *gin.Context, src *NewUserRequest) utils.MongoOperationFunc {
	account := user.NewAccountFromRequest(ctx, src.Email, src.DisplayName, src.Password)
	slog.Debug("New account instance", slog.String("email", src.Email), slog.String("display name", src.DisplayName))

	return func(timeout context.Context, conn *mongo.Client) error {
		if truthy.ValueSlice(ctx.Errors) {
			slog.ErrorContext(timeout, "abort mongo operation: request context contains errors", slog.String("errors", ctx.Errors.String()))
			return ErrInvalidUserRegistration
		}

		cfg, cfgErr := utils.InsertOneOptsWith(utils.ValidateInsertedDocument(true))
		if cfgErr != nil {
			slog.ErrorContext(timeout, "abort mongo operation: error configuring operation options", slog.String("error", cfgErr.Error()))
			return cfgErr
		}

		_, opErr := conn.
			Database(`tournabyte`).
			Collection(`users`).
			InsertOne(timeout, account, cfg)
		if opErr != nil {
			slog.ErrorContext(timeout, "abort mongo operation: error executing operation", slog.String("error", opErr.Error()))
			return opErr
		}

		slog.DebugContext(timeout, "mongo operation completed: account record inserted", slog.String("id", account.ID.Hex()))
		return nil
	}
}

// Function `Authenticate` creates a closure to validate a login attempt and attach the server-side session details to the appropriate account record
//
// Parameters:
//   - ctx: the context managing the lifetime of the request
//   - attempt: the client attempt to respond to the authentication challenge
//   - signer: the signing tool for producing the JWT for later authorization
//   - oracle: the correct login credentials to compare against the attempt
//   - dst: the location to write the session token information for client response
//
// Returns:
//   - `utils.MongoOperationFunc`: a closure capable of recording the necessary server-side session information in the database
func Authenticate(ctx *gin.Context, attempt string, signer jose.Signer, oracle *user.LoginCredentials, dst *AuthenticatedUser) utils.MongoOperationFunc {
	return func(timeout context.Context, conn *mongo.Client) error {

		passwordMatches(ctx, attempt, oracle.PasswordHash)
		access, refresh := makeSessionTokens(ctx, dst.ID, signer)
		session := user.SecureSessionToken(ctx, refresh)

		if truthy.ValueSlice(ctx.Errors) {
			return ErrInvalidSession
		}

		cfg, cfgErr := utils.UpdateOneOptsWith(utils.ValidateUpdatedDocument(true), utils.DoInsertOnNoMatchFound(false))
		if cfgErr != nil {
			return cfgErr
		}

		id, idErr := bson.ObjectIDFromHex(dst.ID)
		if idErr != nil {
			return idErr
		}

		_, opErr := conn.
			Database(`tournabyte`).
			Collection(`users`).
			UpdateByID(
				timeout,
				id,
				bson.M{"$push": bson.M{"active_sessions": session}},
				cfg,
			)
		if opErr != nil {
			return opErr
		}

		dst.AccessToken = access
		dst.RefreshToken = refresh
		return nil
	}
}

// Function `FindLoginDetails` creates a closure to lookup a user's login details based on a unique email address
//
// Parameters:
//   - ctx: the context managing the lifetime of this request
//   - email: the unique identifier to look up
//   - creds: the location to decode the discovered login credentials to
//   - dst: the location to decode the discovered userid to
//
// Returns:
//   - `utils.MongoOperationFunc`: a closure capable of performing the specified lookup and decoding the result
func FindLoginDetails(ctx *gin.Context, email string, creds *user.LoginCredentials, dst *AuthenticatedUser) utils.MongoOperationFunc {
	var account user.FullAccountDetails

	return func(timeout context.Context, conn *mongo.Client) error {
		if truthy.ValueSlice(ctx.Errors) {
			return ErrInvalidSession
		}

		cfg, cfgErr := utils.FindOneOptsWith(
			utils.FindOneProjectSpec(utils.Retain("login_credentials"), utils.Retain("_id")),
		)
		if cfgErr != nil {
			return cfgErr
		}

		opErr := conn.
			Database(`tournabyte`).
			Collection(`users`).
			FindOne(
				timeout,
				utils.Directives(
					utils.Eq("login_credentials.email", email),
				),
				cfg,
			).Decode(&account)
		if opErr != nil {
			return opErr
		}

		creds.Email = account.Credentials.Email
		creds.PasswordHash = account.Credentials.PasswordHash
		dst.ID = account.ID.Hex()
		return nil
	}
}

// Function `makeSessionTokens` creates the access and refresh tokens associated with a new session
//
// Parameters:
//   - ctx: the context managing the lifetime of the request
//   - userid: the ID of the user the new session is for
//   - signer: the signing tool for signing the resulting token
//
// Returns:
//   - `string`: the JWT token portion
//   - `string`: the refresh token portion
func makeSessionTokens(ctx *gin.Context, userid string, signer jose.Signer) (string, string) {
	issueTime := time.Now().UTC()
	public := jwt.Claims{
		Issuer:    "api.tournabyte.com",
		Subject:   "Tournabyte API authorization",
		Expiry:    jwt.NewNumericDate(issueTime.Add(10 * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(issueTime),
		NotBefore: jwt.NewNumericDate(issueTime.Add(15 * time.Second)),
	}

	custom := AuthorizationTokenClaims{
		Owner: userid,
	}

	raw, err := jwt.Signed(signer).Claims(public).Claims(custom).Serialize()
	if err != nil {
		ctx.Error(err)
	}
	return raw, rand.Text()
}

// Function `passwordMatches` compares the provided raw string to the stored hash string
//
// Parameters:
//   - ctx: the context managing the lifetime of the request
//   - provided: the raw authentication attempt
//   - stored: the hashed oracle value
func passwordMatches(ctx *gin.Context, provided string, stored string) {
	match, err := argon2id.ComparePasswordAndHash(provided, stored)
	if err != nil {
		ctx.Error(err)
	}
	if !match {
		ctx.Error(ErrWrongCredentials)
	}
}

// Function `parsedAuthorizationToken` parses the signed JWT and writes the parsed data to the specified address
//
// Parameters:
//   - ctx: the context managing the lifetime of the request
//   - raw: the raw token to parse
//   - dst: the location to write the parsed token to
//   - algorithms...: the signing algorithms to try
//
// Returns:
//   - `bool`: value indicating whether the parsing process succeeded or not
func parsedAuthorizationToken(ctx *gin.Context, raw string, dst *jwt.JSONWebToken, algorithms ...jose.SignatureAlgorithm) bool {
	var parsingError error
	dst, parsingError = jwt.ParseSigned(raw, algorithms)

	if parsingError != nil {
		ctx.Error(parsingError)
		return false
	}
	return true
}

// Function `decodedTokenClaims` decodes the given JWT claims into the sequence of specified addresses
//
// Parameters:
//   - ctx: the context managing the lifetime of the request
//   - token: the token whose claims should be decoded
//   - key: the secret key for decoding token claims
//   - addrs: locations to decode claims to
//
// Returns:
//   - `bool`: value indicating whether the claim decoding process succeeded or not
func decodedTokenClaims(ctx *gin.Context, token *jwt.JSONWebToken, key []byte, addrs ...any) bool {
	if err := token.Claims(key, addrs); err != nil {
		ctx.Error(err)
		return false
	}
	return true
}

// Function `validPublicClaims` validates the utilized public JWT claims match the expected values
//
// Parameters:
//   - ctx: the context manaing the lifetime of the request
//   - claims: the JWT claims to validate
//   - expectedSubject: the expected value for the `sub` field
//   - expectedIssuer: the expected value for the `iss` field
//
// Returns:
//   - `bool`: value indicating whether claim validation succeeded or not
func validPublicClaims(ctx *gin.Context, claims *jwt.Claims, expectedSubject string, expectedIssuer string) bool {
	expected := jwt.Expected{
		Issuer:  expectedIssuer,
		Subject: expectedSubject,
	}
	if err := claims.Validate(expected); err != nil {
		ctx.Error(err)
		return false
	}
	return true
}

// Function `validCustomClaims` validates the custom claims decoded from a JWT against a validator
//
// Parameters:
//   - ctx: the context managing the lifetime of the request
//   - claims: the JWT claims to validate
//   - validatorFunc: custom validation logic
//
// Returns:
//   - `bool`: value indicating whether claim validation succeeded or not
func validCustomClaims(ctx *gin.Context, claims any, validatorFunc *validator.Validate) bool {
	if err := validatorFunc.Struct(claims); err != nil {
		ctx.Error(err)
		return false
	}
	return true
}
