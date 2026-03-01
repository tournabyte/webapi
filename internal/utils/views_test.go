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
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func setupResponseTestRouter() *gin.Engine {
	router := gin.New()

	router.Use(ErrorRecovery())
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

func setupBindingTestRouter() *gin.Engine {
	router := gin.New()

	router.Use(ErrorRecovery())
	router.GET("/items/:id", func(ctx *gin.Context) {
		params := struct {
			ID string `uri:"id" binding:"required,mongodb"`
		}{}

		if !DidBindURI(ctx, &params) {
			RespondWithError(ctx, ErrCouldNotBindRequestParameters)
			return
		}
		RespondWithRequestedData(ctx, gin.H{"URI ID": params.ID}, http.StatusOK)
	})
	router.POST("/items", func(ctx *gin.Context) {
		body := struct {
			Name  string  `json:"name" binding:"required"`
			Price float32 `json:"price" binding:"required,gt=0"`
		}{}

		if !DidBindBody(ctx, &body) {
			RespondWithError(ctx, ErrCouldNotBindRequestBody)
			return
		}
		RespondWithRequestedData(ctx, gin.H{"body": gin.H{"name": body.Name, "price": body.Price}}, http.StatusOK)
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

func TestURIBinding(t *testing.T) {
	server := setupBindingTestRouter()

	t.Run("DidBind", func(t *testing.T) {
		id := bson.NewObjectID()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", fmt.Sprintf("/items/%s", id.Hex()), nil)

		server.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusOK, statusCode)
		assert.Contains(t, responseBody, `"ok":true`)
		assert.Contains(t, responseBody, id.Hex())
	})

	t.Run("DidNotBind", func(t *testing.T) {
		id := uuid.New()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", fmt.Sprintf("/items/%s", id.String()), nil)

		server.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusBadRequest, statusCode)
		assert.Contains(t, responseBody, `"ok":false`)
	})
}

func TestRequestBodyBinding(t *testing.T) {
	server := setupBindingTestRouter()

	t.Run("DidBind", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/items", bytes.NewBuffer([]byte(`{"name": "water bottle", "price": 2.75}`)))

		server.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusOK, statusCode)
		assert.Contains(t, responseBody, `"ok":true`)
		assert.Contains(t, responseBody, `"name":"water bottle"`)
		assert.Contains(t, responseBody, `"price":2.75`)
	})

	t.Run("DidNotBind", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/items", bytes.NewBuffer([]byte(`{"name": "water bottle", "price": -2.75}`)))

		server.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusBadRequest, statusCode)
		assert.Contains(t, responseBody, `"ok":false`)
	})
}
