package core

/*
 * File: pkg/core/auth.go
 *
 * Purpose: authentication/authorization logic
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"context"
	"crypto/rand"
	"errors"
	"log"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/tournabyte/webapi/pkg/dbx"
	"github.com/tournabyte/webapi/pkg/handlerutil"
	"github.com/tournabyte/webapi/pkg/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Workspace keys associated with authentication/authorization workspace tasks
const (
	authRequestKey               = "authenticationRequest"
	userAuthorizationResponseKey = "authorizedUserData"
	authTokenOptionsKey          = "authorizationTokenOptions"
	authSessionOptionsKey        = "authorizationSessionOptions"
	userAccountRecordKey         = "userAccountRecord"
	userSessionRecordKey         = "userSessionRecord"
	accessTokenKey               = "userAccessToken"
	refreshTokenKey              = "userRefreshToken"
)

// Errors specific to authentication/authorization workflow tasks
var (
	ErrInvalidLogin = errors.New("invalid credentials presented")
)

// Function `(*tournabyteAPIService).initUserCreationWorkspace` initializes the handler workspace for a user creation request handling sequence
//
// Parameters:
//   - ctx: the request context to use during workspace initialization
//
// Returns:
//   - `*handlerutil.HandlerWorkspace`: the workspace for creating a user
func (srv *tournabyteAPIService) initUserCreationWorkspace(ctx *gin.Context) *handlerutil.HandlerWorkspace {
	space := handlerutil.DefaultWorkspace()

	bind := handlerutil.BindingsFromRequestContext(ctx, handlerutil.ShouldHaveJSONBody)
	space.Set(handlerutil.RequestBindings, bind)
	space.Set(authSessionOptionsKey, srv.getSessionConfig())
	space.Set(authTokenOptionsKey, srv.getTokenConfig())

	log.Printf("[HANDLER]: setup request bindings and token configurations")

	return &space
}

// Function `userCreationPipeline` initializes a handling pipeline for user creation
//
// Parameters:
//   - ctx: the parent context to control the created pipeline
//
// Returns:
//   - `context.Context`: the context controlling the created pipeline (derived from the given context.Context)
//   - `context.CancelCauseFunc`: the cancellation function controlling pipeline cancellation
//   - `chan<- *handlerutil.HandlerWorkspace`: the input channel for the pipeline (send-only)
//   - `<-chan *handlerutil.HandlerWorkspace`: the output channel for the pipeline (read-only)
func userCreationPipeline(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
	pipelineCtx, pipelineCancel := context.WithCancelCause(ctx)
	pipelineInput := make(chan *handlerutil.HandlerWorkspace)

	out1 := handlerutil.Stage(pipelineCtx, pipelineCancel, bindAuthenticationRequestFormat, pipelineInput)
	out2 := handlerutil.Stage(pipelineCtx, pipelineCancel, deriveAccountRecordFromRequest, out1)
	out3 := handlerutil.Stage(pipelineCtx, pipelineCancel, createAccountRecord, out2)
	out4 := handlerutil.Stage(pipelineCtx, pipelineCancel, validateCredentials, out3)
	out5 := handlerutil.Stage(pipelineCtx, pipelineCancel, createAccessToken, out4)
	out6 := handlerutil.Stage(pipelineCtx, pipelineCancel, createRefreshToken, out5)
	out7 := handlerutil.Stage(pipelineCtx, pipelineCancel, deriveSessionRecord, out6)
	out8 := handlerutil.Stage(pipelineCtx, pipelineCancel, createSessionRecord, out7)
	out9 := handlerutil.Stage(pipelineCtx, pipelineCancel, populateUserAuthorizationResponse, out8)

	return pipelineCtx, pipelineCancel, pipelineInput, out9
}

// Function `bindAuthenticationRequestFormat` binds the request body and saves it to the handler workspace for later processing
//
// Parameters:
//   - ctx: the context managing the handler lifecycle
//   - space: the handler workspace for the uer creation process
//
// Returns:
//   - `error`: error that occurred during this step of a pipeline
func bindAuthenticationRequestFormat(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var body models.AuthenticationRequest
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

	space.Set(authRequestKey, body)
	log.Printf("[HANDLER]: saved request body as variable of type %T within workspace under key %q", body, authRequestKey)
	return nil
}

func deriveAccountRecordFromRequest(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var req models.AuthenticationRequest
	var acct models.UserAccount

	log.Printf("[HANDLER]: loading authentication request information from workspace under key %q", authRequestKey)
	if err := space.Get(authRequestKey, &req); err != nil {
		log.Printf("[HANDLER]: error loading authentication request information (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: initializing account information")
	acct.ID = bson.NewObjectID()
	acct.LoginEmail = req.Email
	acct.Metadata = dbx.InitialMetadata()

	log.Printf("[HANDLER]: hashing password...")
	if hash, err := argon2id.CreateHash(req.Password, argon2id.DefaultParams); err != nil {
		log.Printf("[HANDLER]: error hashing password (%s)", err.Error())
		return err
	} else {
		acct.PasswordHash = hash
	}

	space.Set(userAccountRecordKey, acct)
	log.Printf("[HANDLER]: saved request body as variable of type %T within worspace under key %q", acct, userAccountRecordKey)
	return nil
}

// Function `createAccountRecord` populates the account record for insertion to the database
//
// Parameters:
//   - ctx: the context managing the handler lifecycle
//   - space: the handler workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this step of the pipeline (nil if none occurred)
func createAccountRecord(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var acct models.UserAccount
	var sess *mongo.Session
	var cfg *options.InsertOneOptionsBuilder
	var err error

	log.Printf("[HANDLER]: loading account record from worksapce...")
	if err = space.Get(userAccountRecordKey, &acct); err != nil {
		log.Printf("[HANDLER]: error loading account record from worksapce (%s)", err.Error())
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

	log.Printf("[HANDLER]: performing database insertion operation")
	_, err = sess.Client().
		Database(models.UserAccountQueryContext.Database).
		Collection(models.UserAccountQueryContext.Collection).
		InsertOne(ctx, acct, cfg)

	if err != nil {
		log.Printf("[HANDLER]: error during database insertion operation (%s)", err.Error())
	}

	return err
}

// Function `validateCredentials` validates the given username/password with the stored username/passwordhash
//
// Parameters:
//   - ctx: the context managing the lifecycle of the handler
//   - space: the workspace of the handler
//
// Returns:
//   - `error`: error that occurred during this step of the pipeline (nil if no error occurred)
func validateCredentials(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var acct models.UserAccount
	var req models.AuthenticationRequest
	var err error
	var match bool

	log.Printf("[HANDLER]: loading user account record from workspace...")
	if err = space.Get(userAccountRecordKey, &acct); err != nil {
		log.Printf("[HANDLER]: error loading account record from (%s)", err.Error())
		return nil
	}

	log.Printf("[HANDLER]: loading authentication attempt from workspace...")
	if err = space.Get(authRequestKey, &req); err != nil {
		log.Printf("[HANDLER]: error loading authentication attempt from workspace (%s)", err.Error())
		return nil
	}

	log.Printf("[HANDLER]: comparing password provided in authentication attempt and stored password hash...")
	if match, err = argon2id.ComparePasswordAndHash(req.Password, acct.PasswordHash); err != nil {
		log.Printf("[HANDLER]: error comparing password and hash (%s)", err.Error())
		return err
	} else if !match {
		log.Printf("[HANDLER]: mismatch comparing password and hash")
		return ErrInvalidLogin
	} else {
		log.Printf("[HANDLER]: password and hash match")
		return nil
	}
}

// Function `createAccessToken` creates a signed JWT for authorization to other protected resources
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func createAccessToken(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var issueTime = time.Now().UTC()
	var opts models.TokenOptions
	var publicClaims jwt.Claims
	var customClaims models.AuthorizationTokenClaims
	var acct models.UserAccount
	var err error
	var raw string

	log.Printf("[HANDLER]: loading token options for access token generation...")
	if err = space.Get(authTokenOptionsKey, &opts); err != nil {
		log.Printf("[HANDLER]: error loading token options (%s)", err.Error())
		return nil
	}

	log.Printf("[HANDLER]: loading account record from workspace...")
	if err = space.Get(userAccountRecordKey, &acct); err != nil {
		log.Printf("[HANDLER]: error loading account record from workspace (%s)", err.Error())
		return nil
	}

	log.Printf("[HANDLER]: initializing access token claims...")
	customClaims.Owner = acct.ID.Hex()
	publicClaims.Issuer = opts.Issuer
	publicClaims.Subject = opts.Subject
	publicClaims.IssuedAt = jwt.NewNumericDate(issueTime)
	publicClaims.NotBefore = jwt.NewNumericDate(issueTime)
	publicClaims.Expiry = jwt.NewNumericDate(issueTime.Add(opts.ExpiresIn))

	log.Printf("[HANDLER]: serializing token claims...")
	if raw, err = jwt.Signed(opts.Signer).Claims(publicClaims).Claims(customClaims).Serialize(); err != nil {
		log.Printf("[HANDLER]: error serializing token claims (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: saved signed access token into workspace")
	space.Set(accessTokenKey, raw)
	return nil
}

// Function `createRefreshToken` generates a random refresh token for the user session
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func createRefreshToken(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	raw := rand.Text()
	log.Printf("[HANDLER]: saved signed refresh token into workspace")
	space.Set(refreshTokenKey, raw)
	return nil
}

func deriveSessionRecord(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var opts models.SessionOptions
	var now = time.Now().UTC()
	var acct models.UserAccount
	var sess models.UserSession
	var err error
	var token string
	var hash string

	log.Printf("[HANDLER]: loading authorization session options...")
	if err = space.Get(authSessionOptionsKey, &opts); err != nil {
		log.Printf("[HANDLER]: error loading authorization session options (%s)", err.Error())
		return err
	}
	log.Printf("[HANDLER]: loading account record from workspace...")
	if err = space.Get(userAccountRecordKey, &acct); err != nil {
		log.Printf("[HANDLER]: error loading account record (%s)", err.Error())
		return err
	}
	log.Printf("[HANDLER]: loading refresh token from workspace...")
	if err = space.Get(refreshTokenKey, &token); err != nil {
		log.Printf("[HANDLER]: error loading refresh token (%s)", err.Error())
		return err
	}
	log.Printf("[HANDLER]: creating hash of refresh token...")
	if hash, err = argon2id.CreateHash(token, argon2id.DefaultParams); err != nil {
		log.Printf("[HANDLER]: error creating refresh token hase (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: initializing user session record...")
	sess.ID = hash
	sess.Authorizes = acct.ID
	sess.NotValidBefore = now
	sess.NotValidAfter = now.Add(opts.ExpiresIn)
	sess.Rotated = false

	log.Printf("[HANDLER]: saved user session record to workspace")
	space.Set(userSessionRecordKey, sess)
	return nil
}

// Function `createSessionRecord` populates the fields for a user session document for an authorization request
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func createSessionRecord(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var cfg *options.InsertOneOptionsBuilder
	var sess *mongo.Session
	var auth models.UserSession
	var err error

	log.Printf("[HANDLER]: loading database operation settings...")
	if cfg, err = dbx.NewOptions(dbx.ValidateInsertedDocument(true)); err != nil {
		log.Printf("[HANDLER]: error loading database settings (%s)", err.Error())
		return err
	}
	log.Printf("[HANDLER]: loading session details from workspace...")
	if err := space.Get(userSessionRecordKey, &auth); err != nil {
		log.Printf("[HANDLER]: error loading session details (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading database session from request context...")
	if sess, err = dbx.MongoFromContext(ctx); err != nil {
		log.Printf("[HANDLER]: error loading database session from request context (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: performing database insertion operation...")
	_, err = sess.Client().
		Database(models.UserSessionQueryContext.Database).
		Collection(models.UserSessionQueryContext.Collection).
		InsertOne(ctx, auth, cfg)

	if err != nil {
		log.Printf("[HANDLER]: error performing database insertion (%s)", err.Error())
	}
	return err
}

// Function `populateUserAuthorizationResponse` populates the response field for user authorization response
//
// Parameters:
//   - ctx: the context managing the lifecycle of this handler
//   - space: the workspace to utilize
//
// Returns:
//   - `error`: error that occurred during this processing step
func populateUserAuthorizationResponse(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var authorizationDetails models.AuthenticatedUser
	var userRecord models.UserAccount
	var accessToken string
	var refreshToken string
	var err error

	log.Printf("[HANDLER]: loading user record from workspace...")
	if err = space.Get(userAccountRecordKey, &userRecord); err != nil {
		log.Printf("[HANDLER]: error loading user record from workspace (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading the access token from workspace...")
	if err = space.Get(accessTokenKey, &accessToken); err != nil {
		log.Printf("[HANDLER]: error loading access token from workspace (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: loading the refresh token from workspace...")
	if err = space.Get(refreshTokenKey, &refreshToken); err != nil {
		log.Printf("[HANDLER] error loading refresh token from workspace (%s)", err.Error())
		return err
	}

	log.Printf("[HANDLER]: populating response fields")
	authorizationDetails.ID = userRecord.ID.Hex()
	authorizationDetails.AccessToken = accessToken
	authorizationDetails.RefreshToken = refreshToken

	log.Printf("[HANDLER]: saved response structure to workspace")
	space.Set(userAuthorizationResponseKey, authorizationDetails)
	return nil
}
