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

func (srv *tournabyteAPIService) initUserCreationWorkspace(ctx *gin.Context) *handlerutil.HandlerWorkspace {
	space := handlerutil.DefaultWorkspace()

	bind := handlerutil.BindingsFromRequestContext(ctx, handlerutil.ShouldHaveJSONBody)
	space.Set(handlerutil.RequestBindings, bind)
	space.Set(authSessionOptionsKey, srv.getSessionConfig())
	space.Set(authTokenOptionsKey, srv.getTokenConfig())

	return &space
}

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

	return pipelineCtx, pipelineCancel, pipelineInput, out8
}

func bindAuthenticationRequestFormat(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var body models.AuthenticationRequest
	var bindings handlerutil.Bindings

	if err := space.Get(handlerutil.RequestBindings, &bindings); err != nil {
		return err
	}

	if err := bindings.BindBodyAsJSON(&body); err != nil {
		return err
	}

	space.Set(authRequestKey, body)
	return nil
}

func deriveAccountRecordFromRequest(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var req models.AuthenticationRequest
	var acct models.UserAccount

	if err := space.Get(authRequestKey, &req); err != nil {
		return err
	}

	acct.ID = bson.NewObjectID()
	acct.LoginEmail = req.Email
	acct.Metadata = dbx.InitialMetadata()

	if hash, err := argon2id.CreateHash(req.Password, argon2id.DefaultParams); err != nil {
		return err
	} else {
		acct.PasswordHash = hash
	}

	space.Set(userAccountRecordKey, acct)
	return nil
}

func createAccountRecord(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var acct models.UserAccount
	var sess *mongo.Session
	var cfg *options.InsertOneOptionsBuilder
	var err error

	err = space.Get(userAccountRecordKey, &acct)
	if err != nil {
		return err
	}

	cfg, err = dbx.NewOptions(dbx.ValidateInsertedDocument(true))
	if err != nil {
		return err
	}

	sess, err = dbx.MongoFromContext(ctx)
	if err != nil {
		return err
	}

	_, err = sess.Client().
		Database(models.UserAccountQueryContext.Database).
		Collection(models.UserAccountQueryContext.Collection).
		InsertOne(ctx, acct, cfg)

	return nil
}

func validateCredentials(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var acct models.UserAccount
	var req models.AuthenticationRequest
	var err error
	var match bool

	if err = space.Get(userAccountRecordKey, &acct); err != nil {
		return nil
	}

	if err = space.Get(authRequestKey, &req); err != nil {
		return nil
	}

	if match, err = argon2id.ComparePasswordAndHash(req.Password, acct.PasswordHash); err != nil {
		return err
	} else if !match {
		return ErrInvalidLogin
	} else {
		return nil
	}
}

func createAccessToken(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var issueTime = time.Now().UTC()
	var opts models.TokenOptions
	var publicClaims jwt.Claims
	var customClaims models.AuthorizationTokenClaims
	var acct models.UserAccount
	var err error
	var raw string

	if err = space.Get(authTokenOptionsKey, &opts); err != nil {
		return nil
	}

	if err = space.Get(userAccountRecordKey, &acct); err != nil {
		return nil
	}

	customClaims.Owner = acct.ID.Hex()
	publicClaims.Issuer = opts.Issuer
	publicClaims.Subject = opts.Subject
	publicClaims.IssuedAt = jwt.NewNumericDate(issueTime)
	publicClaims.NotBefore = jwt.NewNumericDate(issueTime)
	publicClaims.Expiry = jwt.NewNumericDate(issueTime.Add(opts.ExpiresIn))

	if raw, err = jwt.Signed(opts.Signer).Claims(publicClaims).Claims(customClaims).Serialize(); err != nil {
		return err
	}

	space.Set(accessTokenKey, raw)
	return nil
}

func createRefreshToken(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	raw := rand.Text()
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

	if err = space.Get(authSessionOptionsKey, &opts); err != nil {
		return err
	}
	if err = space.Get(userAccountRecordKey, &acct); err != nil {
		return err
	}
	if err = space.Get(refreshTokenKey, &token); err != nil {
		return err
	}

	if hash, err = argon2id.CreateHash(token, argon2id.DefaultParams); err != nil {
		return err
	}
	sess.ID = hash
	sess.Authorizes = acct.ID
	sess.NotValidBefore = now
	sess.NotValidAfter = now.Add(opts.ExpiresIn)
	sess.Rotated = false

	space.Set(userSessionRecordKey, sess)
	return nil
}

func createSessionRecord(ctx context.Context, space *handlerutil.HandlerWorkspace) error {
	var cfg *options.InsertOneOptionsBuilder
	var sess *mongo.Session
	var auth models.UserSession
	var err error

	if cfg, err = dbx.NewOptions(dbx.ValidateInsertedDocument(true)); err != nil {
		return err
	}
	if err := space.Get(userSessionRecordKey, &auth); err != nil {
		return err
	}

	if sess, err = dbx.MongoFromContext(ctx); err != nil {
		return err
	}

	_, err = sess.Client().
		Database(models.UserSessionQueryContext.Database).
		Collection(models.UserSessionQueryContext.Collection).
		InsertOne(ctx, auth, cfg)
	return err
}
