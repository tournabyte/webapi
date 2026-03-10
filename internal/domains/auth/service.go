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
func Register(ctx *gin.Context, src *AuthenticationRequest) utils.MongoOperationFunc {
	account := newAccountFromRequest(ctx, src)
	slog.Debug("New account instance", slog.String("email", src.Email))

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
			Database(UserAccountQueryContext.Database).
			Collection(UserAccountQueryContext.Collection).
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
//   - oracle: the correct login credentials to compare against the attempt
//   - dst: the location to write the session and token information for client response
//   - sessLen: the duration the session should be valid for
//   - opts: options for the access token generation
//
// Returns:
//   - `utils.MongoOperationFunc`: a closure capable of recording the necessary server-side session information in the database
func Authenticate(ctx *gin.Context, attempt *AuthenticationRequest, oracle *UserAccount, dst *AuthenticatedUser, tokenCfg *TokenOptions, sessCfg *SessionOptions) utils.MongoOperationFunc {
	return func(timeout context.Context, conn *mongo.Client) error {

		passwordMatches(ctx, attempt, oracle)
		access := makeAccessToken(ctx, oracle.ID, tokenCfg)
		session, refresh := makeSession(ctx, oracle.ID, sessCfg.ExpiresIn)

		if truthy.ValueSlice(ctx.Errors) {
			return ErrInvalidSession
		}

		cfg, cfgErr := utils.InsertOneOptsWith(utils.ValidateInsertedDocument(true))
		if cfgErr != nil {
			return cfgErr
		}

		_, opErr := conn.
			Database(UserSessionQueryContext.Database).
			Collection(UserSessionQueryContext.Collection).
			InsertOne(
				timeout,
				session,
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
func FindLoginDetails(ctx *gin.Context, email string, creds *UserAccount, dst *AuthenticatedUser) utils.MongoOperationFunc {
	return func(timeout context.Context, conn *mongo.Client) error {
		if truthy.ValueSlice(ctx.Errors) {
			return ErrInvalidSession
		}

		cfg, cfgErr := utils.FindOneOptsWith(
			utils.FindOneProjectSpec(
				bson.E{Key: "_id", Value: true},
				bson.E{Key: "login_email", Value: true},
				bson.E{Key: "password_hash", Value: true},
			),
		)
		if cfgErr != nil {
			return cfgErr
		}

		opErr := conn.
			Database(UserAccountQueryContext.Database).
			Collection(UserAccountQueryContext.Collection).
			FindOne(
				timeout,
				bson.D{bson.E{Key: "login_email", Value: email}},
				cfg,
			).Decode(creds)
		if opErr != nil {
			return opErr
		}

		dst.ID = creds.ID.Hex()
		return nil
	}
}

// Function `newAccountFromRequest` creates a `UserAccount` instance based on the given `NewUserRequest` and reports any errors that occur to the request context
//
// Parameters:
//   - ctx: the context managing the lifetime of this request
//   - req: the user account details to initialize the instance with
//
// Returns:
//   - `UserAccount`: the created instance based on the input details
func newAccountFromRequest(ctx *gin.Context, req *AuthenticationRequest) UserAccount {
	var account UserAccount

	account.ID = bson.NewObjectID()
	account.LoginEmail = req.Email
	account.Metadata = utils.InitialMetadata()
	if hash, err := argon2id.CreateHash(req.Password, argon2id.DefaultParams); err != nil {
		ctx.Error(err)
	} else {
		account.PasswordHash = hash
	}

	return account
}

// Function `makeAccessToken` generates the JWT that can be used to access protected resources with an Authorization header
//
// Parameters:
//   - ctx: the context managing the lifetime of this request
//   - userid: the user this access token is for
//   - opts: options to customize the token's claims
//
// Returns:
//   - `string`: the generated access token
func makeAccessToken(ctx *gin.Context, userid bson.ObjectID, opts *TokenOptions) string {
	issueTime := time.Now().UTC()
	public := jwt.Claims{
		Issuer:    opts.Issuer,
		Subject:   opts.Subject,
		Expiry:    jwt.NewNumericDate(issueTime.Add(opts.ExpiresIn)),
		IssuedAt:  jwt.NewNumericDate(issueTime),
		NotBefore: jwt.NewNumericDate(issueTime),
	}

	custom := AuthorizationTokenClaims{
		Owner: userid.Hex(),
	}

	raw, err := jwt.Signed(opts.Signer).Claims(public).Claims(custom).Serialize()
	if err != nil {
		ctx.Error(err)
	}
	return raw
}

// Function `makeSession` creates a session instance for the given user ID that expires in the given duration
//
// Parameters:
//   - ctx: the context managing the lifetime of this request
//   - userid: the user id this session belongs to
//   - expiresIn: the duration the session should remain valid
//
// Returns:
//   - `UserSession`: the structured session infomration to be stored server-side
func makeSession(ctx *gin.Context, userid bson.ObjectID, expiresIn time.Duration) (UserSession, string) {
	var sess UserSession
	raw := rand.Text()
	now := time.Now().UTC()

	hash, hashErr := argon2id.CreateHash(raw, argon2id.DefaultParams)
	if hashErr != nil {
		ctx.Error(hashErr)
	}

	sess.ID = hash
	sess.Authorizes = userid
	sess.NotValidBefore = now
	sess.NotValidAfter = now.Add(expiresIn)
	sess.Rotated = false

	return sess, raw
}

// Function `passwordMatches` compares the provided raw string to the stored hash string
//
// Parameters:
//   - ctx: the context managing the lifetime of the request
//   - provided: the authentication attempt
//   - stored: the hashed oracle value
func passwordMatches(ctx *gin.Context, provided *AuthenticationRequest, stored *UserAccount) {
	match, err := argon2id.ComparePasswordAndHash(provided.Password, stored.PasswordHash)
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
	if parsed, err := jwt.ParseSigned(raw, algorithms); err != nil {
		ctx.Error(err)
		return false
	} else {
		*dst = *parsed
		return true
	}
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
	if token == nil {
		ctx.Error(ErrNilAuthorizationToken)
		return false
	}
	if err := token.Claims(key, addrs...); err != nil {
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
