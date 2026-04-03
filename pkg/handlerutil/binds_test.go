package handlerutil_test

/*
 * File: pkg/handlerutil/binds_test.go
 *
 * Purpose: unit tests for binding actions
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tournabyte/webapi/pkg/handlerutil"
)

type testBindDestination struct {
	Name string `header:"x-request-name" json:"displayName" uri:"name" form:"n" binding:"required"`
}

func setupTestBindServer() *gin.Engine {
	srv := gin.New()
	srv.GET("/header", func(ctx *gin.Context) {
		var dst testBindDestination
		bind := handlerutil.BindingsFromRequestContext(ctx, handlerutil.ShouldHaveHeaders)
		if err := bind.BindHeaders(&dst); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		ctx.String(http.StatusOK, "Binding OK")
	})

	srv.GET("/body", func(ctx *gin.Context) {
		var dst testBindDestination
		bind := handlerutil.BindingsFromRequestContext(ctx, handlerutil.ShouldHaveJSONBody)
		if err := bind.BindBodyAsJSON(&dst); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		ctx.String(http.StatusOK, "Binding OK")
	})

	srv.GET("/uri/:name", func(ctx *gin.Context) {
		var dst testBindDestination
		bind := handlerutil.BindingsFromRequestContext(ctx, handlerutil.ShouldHaveURIValues)
		if err := bind.BindURI(&dst); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		ctx.String(http.StatusOK, "Binding OK")
	})

	srv.GET("/query", func(ctx *gin.Context) {
		var dst testBindDestination
		bind := handlerutil.BindingsFromRequestContext(ctx, handlerutil.ShouldHaveQueryParameters)
		if err := bind.BindQueryParameters(&dst); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		ctx.String(http.StatusOK, "Binding OK")
	})

	return srv
}

func TestBindContextToValues(t *testing.T) {
	srv := setupTestBindServer()

	t.Run("Headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/header", nil)
		r.Header.Add("x-request-name", "Alice")

		srv.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, "Binding OK", responseBody)
	})

	t.Run("Body", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/body", bytes.NewBuffer([]byte(`{"displayName":"Alice"}`)))

		srv.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, "Binding OK", responseBody)
	})

	t.Run("URI", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/uri/Alice", nil)

		srv.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, "Binding OK", responseBody)
	})

	t.Run("Query", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/query?n=Alice", nil)

		srv.ServeHTTP(w, r)
		statusCode := w.Code
		responseBody := w.Body.String()

		assert.Equal(t, http.StatusOK, statusCode)
		assert.Equal(t, "Binding OK", responseBody)
	})
}

func TestBindContextWithoutBindingFunctions(t *testing.T) {
	srv := setupTestBindServer()

	t.Run("NilHeaders", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/body", nil)
		r.Header.Add("x-request-name", "Alice")

		srv.ServeHTTP(w, r)
		statusCode := w.Code

		assert.Equal(t, http.StatusBadRequest, statusCode)
	})

	t.Run("NilBody", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/header", nil)

		srv.ServeHTTP(w, r)
		statusCode := w.Code

		assert.Equal(t, http.StatusBadRequest, statusCode)
	})

	t.Run("NilURI", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/query", bytes.NewBuffer([]byte(`{"displayName":"Alice"}`)))

		srv.ServeHTTP(w, r)
		statusCode := w.Code

		assert.Equal(t, http.StatusBadRequest, statusCode)
	})

	t.Run("NilQuery", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/header", nil)

		srv.ServeHTTP(w, r)
		statusCode := w.Code

		assert.Equal(t, http.StatusBadRequest, statusCode)
	})
}
