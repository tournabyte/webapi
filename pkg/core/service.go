package core

/*
 * File: pkg/core/service.go
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
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-playground/validator/v10"
	"github.com/tournabyte/webapi/pkg/dbx"
	"github.com/tournabyte/webapi/pkg/handlerutil"
	"github.com/tournabyte/webapi/pkg/models"
)

// Function `mongoClientFromConfig` creates the connection configured by the given application configuration
// Parameters:
//   - cfg: the configuration options to use
//
// Returns:
//   - `DatabaseConnection` the resulting establish mongodb connection with client and configuration information available
//   - `error`: the issue that occurred when attempting to create the specified connection (nil if no issue occurred)
func mongoClientFromConfig(cfg *models.ApplicationOptions) (*dbx.MongoConnection, error) {
	return dbx.NewMongoConnection(
		dbx.MongoClientAppName("Tournabyte API"),
		dbx.MongoClientCredentials(cfg.RecordStore.Username, cfg.RecordStore.Password),
		dbx.MongoClientHosts(cfg.RecordStore.Hosts...),
		dbx.MongoClientBSONOptions(dbx.NilSliceAsEmpty),
	)
}

// Function `minioClientFromConfig` creates the MinioConnection instance reflecting the options presented in the provided application options
//
// Parameters:
//   - cfg: the application configuration to extract minio options from
//
// Returns:
//   - `MinioConnection`: minio client lifecycle manager on success
//   - `error`: reported issue on failure
func minioClientFromConfig(cfg *models.ApplicationOptions) (*dbx.MinioConnection, error) {
	return dbx.NewMinioConnection(
		cfg.ObjectStore.Endpoint,
		dbx.MinioStaticCredentials(cfg.ObjectStore.AccessKey, cfg.ObjectStore.SecretKey),
		dbx.MinioUseSecureConnection(false),
		dbx.MinioMaxRetries(3),
	)
}

// Function `tokenSignerFromConfig` creates a token signer based on the tokens options in the given application configuration
//
// Parameters:
//   - cfg: the application configuration to extract token options from
//
// Returns:
//   - `jose.Signer`: token signer for application usage
//   - `error`: reported issue on failure
func tokenSignerFromConfig(cfg *models.ApplicationOptions) (jose.Signer, error) {
	return jose.NewSigner(
		jose.SigningKey{
			Algorithm: jose.SignatureAlgorithm(cfg.Serve.Sessions.Algorithm),
			Key:       []byte(cfg.Serve.Sessions.SigningKey),
		},
		nil,
	)
}

// Function `initLogs` initializes structured logging for the server
//
// Parameters:
//   - cfg: the application configuration to extract the logging options from
//
// Returns:
//   - `*slog.Logger`: the service logger
//   - `error`: issue with logging setup (if any)
func initLogs(cfg *models.ApplicationOptions) error {

	if output, err := parseLogOutputs(cfg.Log.Outputs...); err != nil {
		return err
	} else {
		log.SetFlags(cfg.Log.Flags)
		log.SetPrefix(cfg.Log.Prefix)
		log.SetOutput(io.MultiWriter(output...))
		return nil
	}
}

// Function `parseLogOutputs` creates the logging record destinations corresponding to the list of given outputs
//
// Parameters:
//   - outputs...: the variadic list of log destinations
//
// Returns:
//   - `[]io.Writer`: the list of writer objects to write log records to
//   - `error`: the issue with producing the handler (nil if handler created successfully)
func parseLogOutputs(outputs ...string) ([]io.Writer, error) {
	var writers []io.Writer
	for _, target := range outputs {
		switch {
		case target == "stdout":
			writers = append(writers, os.Stdout)
		case target == "stderr":
			writers = append(writers, os.Stderr)
		case strings.HasPrefix(target, "file:"):
			path := strings.TrimPrefix(target, "file:")
			if filepath.IsAbs(path) {
				if err := os.MkdirAll(filepath.Dir(path), 0666); err != nil {
					return nil, err
				}
			}
			if f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666); err != nil {
				return nil, err
			} else {
				writers = append(writers, f)
			}
		}
	}
	return writers, nil
}

func initErrorFormatter() *handlerutil.HandlerFailureFormatter {
	ffmt := handlerutil.FailureFormatter(
		dbx.IsDuplicateKeyError,
	)
	return &ffmt
}

// Type `tournabyteAPIService` represents the API server for the tournabyte platform
//
// Members:
//   - router: the HTTP multiplexer for the API endpoints
//   - db: the ephemeral database connection to a mongodb deployment
//   - s3: the ephemeral s3 connection to a minio deployment
//   - sess: the JWT signing tool for authorization checks
//   - validationFunc: the ephemeral validator for struct validation
//   - opts: the API configuration options for the API server
type tournabyteAPIService struct {
	router         *gin.Engine
	errfmt         *handlerutil.HandlerFailureFormatter
	db             *dbx.MongoConnection
	s3             *dbx.MinioConnection
	sess           jose.Signer
	validationFunc *validator.Validate
	opts           *models.ApplicationOptions
}

// Function `NewTournabyteService` creates a tournabyte API server instance for handling incoming requests
//
// Parameters:
//   - options: the configuration options to use for the server instance
//
// Returns:
//   - `*TournabyteAPIService`: pointer to the server instance
//   - `error`: issue that occurred during server instantiation (nil if instantiation was successful)
func NewTournabyteService(options *models.ApplicationOptions) (*tournabyteAPIService, error) {
	loggerErr := initLogs(options)
	db, dbErr := mongoClientFromConfig(options)
	s3, s3Err := minioClientFromConfig(options)
	jwt, jwtErr := tokenSignerFromConfig(options)

	if loggerErr != nil {
		log.Printf("Could not setup service logger: %s", loggerErr.Error())
	}

	if dbErr != nil {
		log.Printf("Could not establish connection to mongodb deployment: %s\n", dbErr.Error())
		return nil, dbErr
	}

	if s3Err != nil {
		log.Printf("Could not establish connection to minio deployment: %s\n", s3Err.Error())
		return nil, s3Err
	}

	if jwtErr != nil {
		log.Printf("Could not create the JWT signing tool: %s\n", jwtErr.Error())
		return nil, jwtErr
	}

	return &tournabyteAPIService{
		router:         gin.New(),
		errfmt:         initErrorFormatter(),
		db:             db,
		s3:             s3,
		sess:           jwt,
		validationFunc: validator.New(),
		opts:           options,
	}, nil

}

// Function `(*TournabyteAPIService).Run` starts the server instance in a separate goroutine and enables graceful shutdowns of the system
//
// Returns:
//   - `error`: issue that occurred during server shutdown
func (srv *tournabyteAPIService) Run() error {
	srv.registerRoutes()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", srv.opts.Serve.Port),
		Handler: srv.router,
	}

	quit, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Listening for requests on port %d\n", srv.opts.Serve.Port)
		log.Printf("TLS in use = %t\n", srv.opts.Serve.Security.TLSEnabled)
		var startupError error

		if srv.opts.Serve.Security.TLSEnabled {
			startupError = server.ListenAndServeTLS(srv.opts.Serve.Security.Certificate, srv.opts.Serve.Security.Keychain)
		} else {
			startupError = server.ListenAndServe()
		}

		if startupError != nil && startupError != http.ErrServerClosed {
			log.Printf("Failed to start service: %s\n", startupError.Error())
			stop()
		}

	}()

	<-quit.Done()
	log.Println("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Println("Could not shutdown the server, forcing it anyway...")
		server.Close()
		return err
	}

	srv.db.Disconnect(ctx)
	log.Println("Server exited gracefully")
	return nil
}

// Function `(*tournabyteAPIService).getSessionConfig` isolates the session configuration specific options from the service options
//
// Returns:
//   - `models.SessionOptions`: a structure housing information specific to session token generation and validation
func (srv *tournabyteAPIService) getSessionConfig() models.SessionOptions {
	return models.SessionOptions{
		ExpiresIn: srv.opts.Serve.Sessions.RefreshTokenTTL,
	}
}

// Function `(*tournabyteAPIService).getTokenConfig` isolates the token configuration specific options from the service options
//
// Returns:
//   - `models.TokenOptions`: a structure housing information specific to access token generation and validation
func (srv *tournabyteAPIService) getTokenConfig() models.TokenOptions {
	return models.TokenOptions{
		Signer:    srv.sess,
		Subject:   srv.opts.Serve.Sessions.Subject,
		Issuer:    srv.opts.Serve.Sessions.Issuer,
		ExpiresIn: srv.opts.Serve.Sessions.AccessTokenTTL,
	}
}
