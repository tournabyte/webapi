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
)

type TournabyteAPIService struct {
	router *gin.Engine
	opts   *ApplicationOptions
}

func NewTournabyteService(options *ApplicationOptions) *TournabyteAPIService {
	return &TournabyteAPIService{
		router: gin.New(),
		opts:   options,
	}

}

func (srv *TournabyteAPIService) With(method string, path string, handlers ...gin.HandlerFunc) *TournabyteAPIService {
	srv.router.Handle(
		method,
		path,
		handlers...,
	)
	return srv
}

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
