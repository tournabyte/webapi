package utils

/*
 * File: internal/utils/middleware.go
 *
 * Purpose: HTTP layer middleware handlers for API endpoints
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRecoverTestRouter() *gin.Engine {
	router := gin.New()

	router.Use(ErrorRecovery())
	router.GET("/error400", func(ctx *gin.Context) {
		panic(ValidationFailed("testing", "panic 400"))
	})
	router.GET("/error401", func(ctx *gin.Context) {
		panic(NotAuthorized("testing", "panic 401"))
	})
	router.GET("error403", func(ctx *gin.Context) {
		panic(AccessDenied("testing", "panic 403"))
	})
	router.GET("/error404", func(ctx *gin.Context) {
		panic(ResourceNotFound("testing", "panic 404"))
	})
	router.GET("/error429", func(ctx *gin.Context) {
		panic(SlowDown("testing", "panic 429"))
	})
	router.GET("/error503", func(ctx *gin.Context) {
		panic(TryAgainLater("testing", "panic 503"))
	})
	router.GET("/error500", func(ctx *gin.Context) {
		panic(errors.New("panic 500"))
	})
	router.GET("/any500", func(ctx *gin.Context) {
		panic("panic 500")
	})

	return router
}

func TestErrorRecoveryMiddleware(t *testing.T) {
	server := setupRecoverTestRouter()

	t.Run("HTTP400", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/error400", nil)

		server.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusBadRequest, statusCode)
		assert.Contains(t, responseBody, `"ok":false`)
		assert.Contains(t, responseBody, `"message":"validation checks did not pass"`)
		assert.Contains(t, responseBody, `"details":{"testing":"panic 400"}`)
	})

	t.Run("HTTP401", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/error401", nil)

		server.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusUnauthorized, statusCode)
		assert.Contains(t, responseBody, `"ok":false`)
		assert.Contains(t, responseBody, `"message":"insufficient claim to the requested resource"`)
		assert.Contains(t, responseBody, `"details":{"testing":"panic 401"}`)
	})

	t.Run("HTTP403", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/error403", nil)

		server.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusForbidden, statusCode)
		assert.Contains(t, responseBody, `"ok":false`)
		assert.Contains(t, responseBody, `"message":"insufficient permissions to access the requested resource"`)
		assert.Contains(t, responseBody, `"details":{"testing":"panic 403"}`)
	})

	t.Run("HTTP404", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/error404", nil)

		server.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusNotFound, statusCode)
		assert.Contains(t, responseBody, `"ok":false`)
		assert.Contains(t, responseBody, `"message":"requested resource does not exist"`)
		assert.Contains(t, responseBody, `"details":{"testing":"panic 404"}`)
	})

	t.Run("HTTP429", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/error429", nil)

		server.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusTooManyRequests, statusCode)
		assert.Contains(t, responseBody, `"ok":false`)
		assert.Contains(t, responseBody, `"message":"rate limit exceeded"`)
		assert.Contains(t, responseBody, `"details":{"testing":"panic 429"}`)
	})

	t.Run("HTTP503", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/error503", nil)

		server.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusServiceUnavailable, statusCode)
		assert.Contains(t, responseBody, `"ok":false`)
		assert.Contains(t, responseBody, `"message":"try again later"`)
		assert.Contains(t, responseBody, `"details":{"testing":"panic 503"}`)
	})

	t.Run("HTTP500", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/error500", nil)

		server.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusInternalServerError, statusCode)
		assert.Contains(t, responseBody, `"ok":false`)
		assert.Contains(t, responseBody, `"message":"panic 500"`)
		assert.NotContains(t, responseBody, `"details":`)
	})

	t.Run("HTTP5XX", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/any500", nil)

		server.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusInternalServerError, statusCode)
		assert.Contains(t, responseBody, `"ok":false`)
		assert.Contains(t, responseBody, `"message":"try again later"`)
		assert.NotContains(t, responseBody, `"details":`)
	})
}
