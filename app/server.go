package app

/*
 * File: app/server.go
 *
 * Purpose: definition of the server for the Tournabyte webapi
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-playground/validator/v10"
	"github.com/tournabyte/webapi/internal/domains/auth"
	"github.com/tournabyte/webapi/internal/utils"
)

// Function `MongoClientFromConfig` creates the connection configured by the given application configuration
// Parameters:
//   - cfg: the configuration options to use
//
// Returns:
//   - `DatabaseConnection` the resulting establish mongodb connection with client and configuration information available
//   - `error`: the issue that occurred when attempting to create the specified connection (nil if no issue occurred)
func MongoClientFromConfig(cfg *ApplicationOptions) (*utils.DatabaseConnection, error) {
	slog.Info("Creating mongodb client")
	return utils.NewMongoConnection(
		utils.MongoClientAppName("Tournabyte API"),
		utils.MongoClientCredentials(cfg.Database.Username, cfg.Database.Password),
		utils.MongoClientHosts(cfg.Database.Hosts...),
		utils.MongoClientBSONOptions(utils.NilSliceAsEmpty),
	)
}

// Function `MinioClientFromConfig` creates the MinioConnection instance reflecting the options presented in the provided application options
//
// Parameters:
//   - cfg: the application configuration to extract minio options from
//
// Returns:
//   - `MinioConnection`: minio client lifecycle manager on success
//   - `error`: reported issue on failure
func MinioClientFromConfig(cfg *ApplicationOptions) (*utils.MinioConnection, error) {
	slog.Info("Creating minio client")
	return utils.NewMinioConnection(
		cfg.Filestore.Endpoint,
		utils.MinioStaticCredentials(cfg.Filestore.AccessKey, cfg.Filestore.SecretKey),
		utils.MinioUseSecureConnection(false),
		utils.MinioMaxRetries(3),
	)
}

// Function `TokenSignerFromConfig` creates a token signer based on the tokens options in the given application configuration
//
// Parameters:
//   - cfg: the application configuration to extract token options from
//
// Returns:
//   - `jose.Signer`: token signer for application usage
//   - `error`: reported issue on failure
func TokenSignerFromConfig(cfg *ApplicationOptions) (jose.Signer, error) {
	return jose.NewSigner(
		jose.SigningKey{
			Algorithm: jose.SignatureAlgorithm(cfg.Serve.Sessions.Algorithm),
			Key:       []byte(cfg.Serve.Sessions.SigningKey),
		},
		nil,
	)
}

// Type `TournabyteAPIService` represents the API server for the tournabyte platform
//
// Fields:
//   - router: the HTTP multiplexer for the API endpoints
//   - db: the ephemeral database connection to a mongodb deployment
//   - s3: the ephemeral s3 connection to a minio deployment
//   - sess: the JWT signing tool for authorization checks
//   - opts: the API configuration options for the API server
type TournabyteAPIService struct {
	router         *gin.Engine
	db             *utils.DatabaseConnection
	s3             *utils.MinioConnection
	sess           jose.Signer
	validationFunc *validator.Validate
	opts           *ApplicationOptions
}

// Function `NewTournabyteService` creates a tournabyte API server instance for handling incoming requests
//
// Parameters:
//   - options: the configuration options to use for the server instance
//
// Returns:
//   - `*TournabyteAPIService`: pointer to the server instance
//   - `error`: issue that occurred during server instantiation (nil if instantiation was successful)
func NewTournabyteService(options *ApplicationOptions) (*TournabyteAPIService, error) {
	db, dbErr := MongoClientFromConfig(options)
	s3, s3Err := MinioClientFromConfig(options)
	jwt, jwtErr := TokenSignerFromConfig(options)

	if dbErr != nil {
		slog.Error("Could not establish connection to mongodb deployment", slog.String("err", dbErr.Error()))
		return nil, dbErr
	}

	if s3Err != nil {
		slog.Error("Could not establish connection to minio deployment", slog.String("err", s3Err.Error()))
		return nil, s3Err
	}

	if jwtErr != nil {
		slog.Error("Could not create the JWT signing tool", slog.String("err", jwtErr.Error()))
		return nil, jwtErr
	}

	return &TournabyteAPIService{
		router:         gin.New(),
		db:             db,
		s3:             s3,
		sess:           jwt,
		validationFunc: validator.New(),
		opts:           options,
	}, nil

}

// Function `(*TournabyteAPIService).addGlobalMiddleware` setups the base handler chain to respond to any requests received
func (srv *TournabyteAPIService) addGlobalMiddleware() {
	srv.router.Use(gin.CustomRecovery(
		func(ctx *gin.Context, err any) {
			if err != nil {
				switch err := err.(type) {
				case error:
					slog.Error("Recovering from error panic", slog.String("error", err.Error()))
					utils.RespondWithError(ctx, err)
				default:
					slog.Error("Recovering from non error panic")
					utils.RespondWithError(ctx, utils.ErrUndisclosedHandlerFailure)
				}
			}
		},
	))
}

// Function `(*TournabyteAPIService).addAuthGroup` setups the handler chains to respond to requests on the `/auth/` endpoint group
//
// Parameters:
//   - parentGroup: the router group to attach to
func (srv *TournabyteAPIService) addAuthGroup(parentGroup *gin.RouterGroup) {
	authGroup := parentGroup.Group("auth")

	// POST /auth/register
	authGroup.POST("register", auth.UserRegistrationHandler(srv.db, srv.sess, srv.opts.Serve.Sessions.Issuer, srv.opts.Serve.Sessions.Subject))

	// POST /auth/login
	authGroup.POST("login", auth.UserAuthenticationHandler(srv.db, srv.sess, srv.opts.Serve.Sessions.Issuer, srv.opts.Serve.Sessions.Subject))

	// GET /auth/:userid
	authGroup.GET(
		"/:userid",
		auth.CheckAuthorizationHeaderHandler(
			srv.opts.Serve.Sessions.SigningKey,
			srv.opts.Serve.Sessions.Issuer,
			srv.opts.Serve.Sessions.Subject,
			srv.validationFunc,
			jose.SignatureAlgorithm(srv.opts.Serve.Sessions.Algorithm),
		),
		func(ctx *gin.Context) {
			type R struct {
				ID string `uri:"userid" binding:"required,mongodb"`
			}
			var r R
			if err := ctx.ShouldBindUri(&r); err != nil {
				slog.Error("Could not bind URI parameter(s)")
				ctx.AbortWithStatusJSON(400, gin.H{"msg": err.Error()})
				return
			}
			if r.ID != ctx.GetString(auth.AuthorizationClaims) {
				slog.ErrorContext(ctx, "Could not validate authorization claims", auth.AuthorizationClaims, ctx.GetString(auth.AuthorizationClaims))
				ctx.AbortWithStatusJSON(401, gin.H{"msg": "Unauthorized"})
				return
			}
			ctx.JSON(200, gin.H{"user": r.ID, "msg": "successfully accessed protected resource"})
		},
	)
}

// Function `(*TournabyteAPIService).RegisterRoutes` initializes the underlying engine with the appropriate routes for service
func (srv *TournabyteAPIService) RegisterRoutes() {
	srv.addGlobalMiddleware()

	{
		// /v1/...
		v1 := srv.router.Group("v1")
		srv.addAuthGroup(v1)
	}

}

// Function `(*TournabyteAPIService).Run` starts the server instance in a separate goroutine and enables graceful shutdowns of the system
//
// Returns:
//   - `error`: issue that occurred during server shutdown
func (srv *TournabyteAPIService) Run() error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", srv.opts.Serve.Port),
		Handler: srv.router,
	}

	quit, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("Starting API server on", slog.Int("port", int(srv.opts.Serve.Port)))
		slog.Debug("Security options ", slog.Bool("using TLS", srv.opts.Serve.Security.TLSEnabled))
		var startupError error
		defer srv.db.Disconnect(context.Background())

		if srv.opts.Serve.Security.TLSEnabled {
			startupError = server.ListenAndServeTLS(srv.opts.Serve.Security.Certificate, srv.opts.Serve.Security.Keychain)
		} else {
			startupError = server.ListenAndServe()
		}

		if startupError != nil && startupError != http.ErrServerClosed {
			slog.Error("Failed to start service", slog.Any("error", startupError), slog.Any("config", srv.opts.Serve))
			stop()
		}

	}()

	<-quit.Done()
	slog.Info("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		server.Close()
		slog.Error("Could not shutdown the server, forcing it anyway")
		return err
	}
	slog.Info("Server exited gracefully")
	return nil
}
