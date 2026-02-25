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
