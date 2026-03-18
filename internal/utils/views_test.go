package utils

/*
 * File: internal/utils/views_test.go
 *
 * Purpose: Unit tests for views.go source file
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

func setupResponseTestRouter() *gin.Engine {
	router := gin.New()

	router.GET("/ok", func(ctx *gin.Context) {
		RespondWithRequestedData(ctx, gin.H{"someData": 5}, http.StatusOK)
	})
	router.GET("/fail", func(ctx *gin.Context) {
		RespondWithError(ctx, ErrCouldNotBindRequestBody)
	})
	router.GET("/error", func(ctx *gin.Context) {
		RespondWithError(ctx, errors.New("unanticipated error"))
	})

	return router
}

func TestResponseFormatting(t *testing.T) {
	server := setupResponseTestRouter()

	t.Run("GotData", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/ok", nil)

		server.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusOK, statusCode)
		assert.Contains(t, responseBody, `"ok":true`)
		assert.Contains(t, responseBody, `"data":{"someData":5}`)
	})

	t.Run("GotFailure", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/fail", nil)

		server.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusBadRequest, statusCode)
		assert.Contains(t, responseBody, `"ok":false`)
		assert.Contains(t, responseBody, `"error":`)
		assert.Contains(t, responseBody, `"message":"request body malformed"`)
	})

	t.Run("GotError", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/error", nil)

		server.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusInternalServerError, statusCode)
		assert.Contains(t, responseBody, `"ok":false`)
		assert.Contains(t, responseBody, `"error":`)
		assert.Contains(t, responseBody, `"message":"unanticipated error"`)
	})
}
