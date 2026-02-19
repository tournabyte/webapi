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
	"github.com/tournabyte/webapi/internal/domains/user"
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
		utils.MongoAppName("Tournabyte API"),
		utils.MongoCredentials(cfg.Database.Username, cfg.Database.Password),
		utils.MongoHosts(cfg.Database.Hosts...),
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
			Algorithm: jose.SignatureAlgorithm(cfg.Serve.Tokens.Algorithm),
			Key:       []byte(cfg.Serve.Tokens.PrivateKey),
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
	router *gin.Engine
	db     *utils.DatabaseConnection
	s3     *utils.MinioConnection
	sess   jose.Signer
	opts   *ApplicationOptions
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
		router: gin.New(),
		db:     db,
		s3:     s3,
		sess:   jwt,
		opts:   options,
	}, nil

}

// Function `(*TournabyteAPIService).addAuthGroup` setups the handler chains to respond to requests on the `/auth/` endpoint group
//
// Parameters:
//   - parentGroup: the router group to attach to
func (srv *TournabyteAPIService) addAuthGroup(parentGroup *gin.RouterGroup) {
	auth := parentGroup.Group("auth")

	// POST /auth/register
	auth.POST("register", user.CreateUserHandler(srv.db))

	// POST /auth/login
	auth.POST("login", user.CheckLoginHandler(srv.db, srv.sess))
}

// Function `(*TournabyteAPIService).RegisterRoutes` initializes the underlying engine with the appropriate routes for service
func (srv *TournabyteAPIService) RegisterRoutes() {
	srv.router.Use(utils.ErrorRecovery())

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
		slog.Debug("Using TLS: ", slog.Bool("useTLS", srv.opts.Serve.UseTLS))
		var startupError error

		if srv.opts.Serve.UseTLS {
			startupError = server.ListenAndServeTLS(srv.opts.Serve.CertFile, srv.opts.Serve.KeyFile)
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
