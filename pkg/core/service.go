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
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
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
func initLogs(cfg *models.ApplicationOptions) (*slog.Logger, error) {
	var handlers []slog.Handler

	for _, logConf := range cfg.Log {
		if h, err := makeHandler(logConf.Level, logConf.Destination, logConf.UseJSON, logConf.UseSource); err != nil {
			return nil, err
		} else {
			handlers = append(handlers, h)
		}
	}

	return slog.New(slog.NewMultiHandler(handlers...)), nil
}

// Function `makeHandler` creates the logging record handler corresponding to the given configuration object
//
// Parameters:
//   - lvl: the logging level of the handler
//   - dst: the locations the handler will write to
//   - useJSON: flag to indicate whether JSON format should be used
//   - useSource: flag to indicate whether source code locaiton sould be included in logging records
//
// Returns:
//   - `slog.Handler`: the handler to process logging records
//   - `error`: the issue with producing the handler (nil if handler created successfully)
func makeHandler(lvl string, dst []string, useJSON bool, useSource bool) (slog.Handler, error) {
	var level slog.Level
	var outputs []io.Writer
	var handler slog.Handler
	var opts slog.HandlerOptions

	switch strings.ToLower(lvl) {
	case "debug":
		level = slog.LevelDebug
	case "error":
		level = slog.LevelError
	case "warn":
		level = slog.LevelWarn
	case "info":
		level = slog.LevelInfo
	default:
		return nil, errors.New("Invalid logging level provided")
	}

	for _, dst := range dst {
		switch dst {
		case "std.out":
			outputs = append(outputs, os.Stdout)
		case "std.err":
			outputs = append(outputs, os.Stderr)
		default:
			if filepath.IsAbs(dst) {
				if err := os.MkdirAll(filepath.Dir(dst), 0666); err != nil {
					return nil, err
				}
			}
			if f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666); err != nil {
				return nil, err
			} else {
				outputs = append(outputs, f)
			}
		}
	}

	opts = slog.HandlerOptions{Level: level, AddSource: useSource}
	if useJSON {
		handler = slog.NewJSONHandler(io.MultiWriter(outputs...), &opts)
	} else {
		handler = slog.NewTextHandler(io.MultiWriter(outputs...), &opts)
	}
	return handler, nil
}

func initErrorFormatter() *handlerutil.HandlerFailureFormatter {
	return &handlerutil.HandlerFailureFormatter{}
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
	logger         *slog.Logger
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
	logger, loggerErr := initLogs(options)
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
		logger:         logger,
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
